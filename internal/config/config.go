package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config describes the Windows client launcher profile.
//
// The GUI stays intentionally small and delegates protocol work to a backend
// executable. For the current MVP, route/full-tunnel modes use the native
// pvpn-client-wintun.exe backend from the ITSProto repository.
type Config struct {
	Server    string
	Token     string
	DeviceID  string
	ServerPub string
	Mode      string
	Route     string
	Command   string
	Arguments string

	TunName  string
	TunCIDR  string
	TunnelIP string
	Gateway  string
	MTU      string
	Debug    string
}

func DefaultPath() string {
	if runtime.GOOS == "windows" {
		if programData := os.Getenv("ProgramData"); programData != "" {
			return filepath.Join(programData, "ITSProto", "client.yaml")
		}
		return `C:\ProgramData\ITSProto\client.yaml`
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "itsproto", "client.yaml")
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".config", "itsproto", "client.yaml")
	}
	return "client.yaml"
}

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	cfg := Config{
		Mode:     "route",
		Route:    "1.1.1.1/32",
		TunName:  "ITSProto",
		TunCIDR:  "10.77.0.2/24",
		TunnelIP: "10.77.0.2",
		Gateway:  "10.77.0.1",
		MTU:      "1000",
		Debug:    "true",
	}
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := stripComment(scanner.Text())
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			return Config{}, fmt.Errorf("%s:%d: expected key: value", path, lineNo)
		}
		key := strings.TrimSpace(line[:idx])
		val := cleanValue(strings.TrimSpace(line[idx+1:]))
		switch key {
		case "server":
			cfg.Server = val
		case "token":
			cfg.Token = val
		case "device_id":
			cfg.DeviceID = val
		case "server_pub":
			cfg.ServerPub = val
		case "mode":
			cfg.Mode = val
		case "route":
			cfg.Route = val
		case "command":
			cfg.Command = val
		case "arguments":
			cfg.Arguments = val
		case "tun_name":
			cfg.TunName = val
		case "tun_cidr":
			cfg.TunCIDR = val
		case "tunnel_ip":
			cfg.TunnelIP = val
		case "gateway":
			cfg.Gateway = val
		case "mtu":
			cfg.MTU = val
		case "debug":
			cfg.Debug = val
		default:
			return Config{}, fmt.Errorf("%s:%d: unknown key %q", path, lineNo, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var missing []string
	if c.Server == "" {
		missing = append(missing, "server")
	}
	if c.Token == "" {
		missing = append(missing, "token")
	}
	if c.DeviceID == "" {
		missing = append(missing, "device_id")
	}
	if c.ServerPub == "" {
		missing = append(missing, "server_pub")
	}
	if c.Command == "" {
		missing = append(missing, "command")
	}
	if c.Arguments == "" {
		missing = append(missing, "arguments")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config keys: %s", strings.Join(missing, ", "))
	}
	switch c.Mode {
	case "route", "full-tunnel", "transport-check":
	default:
		return errors.New("mode must be one of: route, full-tunnel, transport-check")
	}
	if c.Mode != "transport-check" {
		if c.TunName == "" || c.TunCIDR == "" || c.TunnelIP == "" || c.Gateway == "" || c.MTU == "" {
			return errors.New("route/full-tunnel mode requires tun_name, tun_cidr, tunnel_ip, gateway, mtu")
		}
	}
	return nil
}

func (c Config) ExpandArguments() string {
	debugArg := ""
	if strings.EqualFold(c.Debug, "true") || c.Debug == "1" || strings.EqualFold(c.Debug, "yes") {
		debugArg = "-debug"
	}
	modeArg := ""
	if c.Mode == "full-tunnel" {
		modeArg = "-full-tunnel"
	}
	repl := strings.NewReplacer(
		"{server}", c.Server,
		"{token}", c.Token,
		"{device_id}", c.DeviceID,
		"{server_pub}", c.ServerPub,
		"{mode}", c.Mode,
		"{route}", c.Route,
		"{tun_name}", c.TunName,
		"{tun_cidr}", c.TunCIDR,
		"{tunnel_ip}", c.TunnelIP,
		"{gateway}", c.Gateway,
		"{mtu}", c.MTU,
		"{debug_arg}", debugArg,
		"{full_tunnel_arg}", modeArg,
	)
	return strings.Join(strings.Fields(repl.Replace(c.Arguments)), " ")
}

func stripComment(s string) string {
	inQuote := false
	quote := rune(0)
	for i, r := range s {
		if (r == '\'' || r == '"') && (i == 0 || s[i-1] != '\\') {
			if !inQuote {
				inQuote = true
				quote = r
			} else if quote == r {
				inQuote = false
			}
		}
		if r == '#' && !inQuote {
			return s[:i]
		}
	}
	return s
}

func cleanValue(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
