//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/vorkauteer/itsproto-windows-client/internal/config"
	"github.com/vorkauteer/itsproto-windows-client/internal/runner"
)

const (
	cwUseDefault = 0x80000000

	wmCreate  = 0x0001
	wmDestroy = 0x0002
	wmCommand = 0x0111
	wmTimer   = 0x0113
	wmClose   = 0x0010

	wsOverlapped = 0x00000000
	wsCaption    = 0x00C00000
	wsSysMenu    = 0x00080000
	wsMinimize   = 0x20000000
	wsVisible    = 0x10000000
	wsChild      = 0x40000000
	wsBorder     = 0x00800000
	wsTabStop    = 0x00010000
	wsVScroll    = 0x00200000
	wsDisabled   = 0x08000000

	esMultiline   = 0x0004
	esReadOnly    = 0x0800
	esAutoVScroll = 0x0040

	bsPushButton = 0x00000000
	ssLeft       = 0x00000000

	idConnect = 1001
	idStop    = 1002
	idReload  = 1003
	idTimer   = 2001

	emSetSel        = 0x00B1
	emReplaceSel    = 0x00C2
	wmSetText       = 0x000C
	wmGetTextLength = 0x000E

	swShow = 5
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}

type msg struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	procRegisterClassEx = user32.NewProc("RegisterClassExW")
	procCreateWindowEx  = user32.NewProc("CreateWindowExW")
	procDefWindowProc   = user32.NewProc("DefWindowProcW")
	procShowWindow      = user32.NewProc("ShowWindow")
	procUpdateWindow    = user32.NewProc("UpdateWindow")
	procGetMessage      = user32.NewProc("GetMessageW")
	procTranslateMsg    = user32.NewProc("TranslateMessage")
	procDispatchMsg     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage = user32.NewProc("PostQuitMessage")
	procSetWindowText   = user32.NewProc("SetWindowTextW")
	procSendMessage     = user32.NewProc("SendMessageW")
	procSetTimer        = user32.NewProc("SetTimer")
	procKillTimer       = user32.NewProc("KillTimer")
	procMessageBox      = user32.NewProc("MessageBoxW")
	procLoadCursor      = user32.NewProc("LoadCursorW")

	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
	procGetStockObject  = gdi32.NewProc("GetStockObject")

	mainHwnd    syscall.Handle
	statusHwnd  syscall.Handle
	logHwnd     syscall.Handle
	connectHwnd syscall.Handle
	stopHwnd    syscall.Handle

	cfgPath    string
	currentCfg config.Config
	procRunner runner.ProcessRunner
	logCh      = make(chan string, 256)
)

func main() {
	flag.StringVar(&cfgPath, "config", config.DefaultPath(), "path to client.yaml")
	flag.Parse()
	if err := run(); err != nil {
		messageBox(0, "ITSProto Windows Client", err.Error())
		os.Exit(1)
	}
}

func run() error {
	instance := getModuleHandle()
	className := utf16Ptr("ITSProtoWindowClass")
	cursor, _, _ := procLoadCursor.Call(0, uintptr(32512)) // IDC_ARROW
	bg, _, _ := procGetStockObject.Call(0)                 // WHITE_BRUSH
	wc := wndClassEx{
		Size:       uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:    syscall.NewCallback(wndProc),
		Instance:   instance,
		Cursor:     syscall.Handle(cursor),
		Background: syscall.Handle(bg),
		ClassName:  className,
	}
	atom, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return fmt.Errorf("RegisterClassExW failed: %v", err)
	}

	hwnd, _, err := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("ITSProto"))),
		uintptr(wsOverlapped|wsCaption|wsSysMenu|wsMinimize|wsVisible),
		uintptr(200), uintptr(200), uintptr(430), uintptr(260),
		0, 0, uintptr(instance), 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed: %v", err)
	}
	mainHwnd = syscall.Handle(hwnd)
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)
	procSetTimer.Call(hwnd, idTimer, 200, 0)

	logAsync("config path: " + cfgPath)
	loadConfig()

	var m msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMsg.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMsg.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

func wndProc(hwnd syscall.Handle, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		createControls(hwnd)
		return 0
	case wmCommand:
		switch int(wParam & 0xffff) {
		case idConnect:
			connect()
		case idStop:
			stop()
		case idReload:
			loadConfig()
		}
		return 0
	case wmTimer:
		drainLogs()
		return 0
	case wmClose:
		stop()
		procKillTimer.Call(uintptr(hwnd), idTimer)
		procPostQuitMessage.Call(0)
		return 0
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return ret
}

func createControls(parent syscall.Handle) {
	statusHwnd = createControl("STATIC", "Status: not connected", wsChild|wsVisible|ssLeft, 18, 18, 380, 22, parent, 0)
	connectHwnd = createControl("BUTTON", "Connect", wsChild|wsVisible|wsTabStop|bsPushButton, 18, 48, 118, 32, parent, idConnect)
	stopHwnd = createControl("BUTTON", "Disconnect", wsChild|wsVisible|wsTabStop|bsPushButton, 148, 48, 118, 32, parent, idStop)
	_ = createControl("BUTTON", "Reload config", wsChild|wsVisible|wsTabStop|bsPushButton, 278, 48, 120, 32, parent, idReload)
	logHwnd = createControl("EDIT", "", wsChild|wsVisible|wsBorder|wsVScroll|esMultiline|esReadOnly|esAutoVScroll, 18, 94, 380, 118, parent, 0)
}

func createControl(class, title string, style uint32, x, y, w, height int32, parent syscall.Handle, id int) syscall.Handle {
	hwnd, _, _ := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(class))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(height),
		uintptr(parent), uintptr(id), 0, 0,
	)
	return syscall.Handle(hwnd)
}

func loadConfig() {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		setStatus("Status: config error")
		logAsync("config error: " + err.Error())
		return
	}
	currentCfg = cfg
	setStatus("Status: ready, server " + cfg.Server)
	logAsync("config loaded: " + cfg.Server + ", mode=" + cfg.Mode)
}

func connect() {
	if currentCfg.Command == "" {
		loadConfig()
	}
	if currentCfg.Command == "" {
		return
	}
	if procRunner.Running() {
		logAsync("backend is already running")
		return
	}
	setStatus("Status: connecting")
	if err := procRunner.StartWithSearch(currentCfg.Command, currentCfg.ExpandArguments(), backendSearchDirs(), logAsync); err != nil {
		setStatus("Status: failed")
		logAsync("start error: " + err.Error())
		return
	}
	setStatus("Status: connected/running")
}

func backendSearchDirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil && exe != "" {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if cfgPath != "" {
		dirs = append(dirs, filepath.Dir(cfgPath))
	}
	if programData := os.Getenv("ProgramData"); programData != "" {
		dirs = append(dirs, filepath.Join(programData, "ITSProto"))
	}
	return dirs
}

func stop() {
	_ = procRunner.Stop(logAsync)
	setStatus("Status: stopped")
}

func setStatus(s string) {
	if statusHwnd == 0 {
		return
	}
	procSetWindowText.Call(uintptr(statusHwnd), uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func logAsync(s string) {
	line := time.Now().Format("15:04:05") + " " + s
	select {
	case logCh <- line:
	default:
	}
}

func drainLogs() {
	for {
		select {
		case line := <-logCh:
			appendLog(line + "\r\n")
		default:
			return
		}
	}
}

func appendLog(s string) {
	if logHwnd == 0 {
		return
	}
	length, _, _ := procSendMessage.Call(uintptr(logHwnd), wmGetTextLength, 0, 0)
	procSendMessage.Call(uintptr(logHwnd), emSetSel, length, length)
	procSendMessage.Call(uintptr(logHwnd), emReplaceSel, 0, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func getModuleHandle() syscall.Handle {
	h, _, _ := procGetModuleHandle.Call(0)
	return syscall.Handle(h)
}

func messageBox(hwnd syscall.Handle, title, text string) {
	procMessageBox.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(title))), 0)
}

func utf16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(strings.ReplaceAll(s, "\n", "\r\n"))
	if err != nil {
		return syscall.StringToUTF16Ptr("")
	}
	return p
}
