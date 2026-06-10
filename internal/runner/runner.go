package runner

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ProcessRunner struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func (r *ProcessRunner) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cmd != nil && r.cmd.Process != nil
}

func (r *ProcessRunner) Start(command string, argumentLine string, logf func(string)) error {
	return r.StartWithSearch(command, argumentLine, nil, logf)
}

func (r *ProcessRunner) StartWithSearch(command string, argumentLine string, searchDirs []string, logf func(string)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil && r.cmd.Process != nil {
		return errors.New("client backend is already running")
	}
	resolvedCommand, err := resolveCommand(command, searchDirs)
	if err != nil {
		return err
	}
	args := splitArgs(argumentLine)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, resolvedCommand, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}
	r.cmd = cmd
	r.cancel = cancel
	logf(fmt.Sprintf("started: %s %s", resolvedCommand, strings.Join(args, " ")))
	go copyLines("stdout", stdout, logf)
	go copyLines("stderr", stderr, logf)
	go func() {
		err := cmd.Wait()
		r.mu.Lock()
		if r.cmd == cmd {
			r.cmd = nil
			r.cancel = nil
		}
		r.mu.Unlock()
		if err != nil {
			logf("backend exited: " + err.Error())
			return
		}
		logf("backend exited cleanly")
	}()
	return nil
}

func resolveCommand(command string, searchDirs []string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", errors.New("backend command is empty")
	}
	if filepath.IsAbs(command) || strings.ContainsAny(command, `/\\`) {
		if _, err := os.Stat(command); err != nil {
			return "", fmt.Errorf("backend executable not found: %s: %w", command, err)
		}
		return command, nil
	}
	for _, dir := range searchDirs {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, command)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	found, err := exec.LookPath(command)
	if err == nil {
		return found, nil
	}
	return "", fmt.Errorf("backend executable %q not found. Put it next to the GUI exe, put it in C:\\ProgramData\\ITSProto, set an absolute command path in client.yaml, or add it to PATH", command)
}

func (r *ProcessRunner) Stop(logf func(string)) error {
	r.mu.Lock()
	cmd := r.cmd
	cancel := r.cancel
	r.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	logf("stopping backend")
	if cancel != nil {
		cancel()
	}
	done := make(chan struct{})
	go func() {
		_, _ = cmd.Process.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
	}
	return nil
}

func copyLines(prefix string, r io.Reader, logf func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logf(prefix + ": " + scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		logf(prefix + " read error: " + err.Error())
	}
}

// splitArgs is a small Windows-friendly argument splitter for config-provided
// command lines. It supports single and double quotes and backslash escaping.
func splitArgs(s string) []string {
	var args []string
	var cur strings.Builder
	inQuote := false
	quote := rune(0)
	escaped := false
	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '\'' || r == '"' {
			if !inQuote {
				inQuote = true
				quote = r
				continue
			}
			if quote == r {
				inQuote = false
				continue
			}
		}
		if (r == ' ' || r == '\t' || r == '\n') && !inQuote {
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteRune(r)
	}
	if escaped {
		cur.WriteRune('\\')
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}
