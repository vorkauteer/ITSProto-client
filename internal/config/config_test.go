package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRouteModeConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "client.yaml")
	data := `
server: "72.56.68.200:9443"
token: "dev-token"
device_id: "windows-client-1"
server_pub: "abc"
mode: "route"
route: "1.1.1.1/32"
tun_name: "ITSProto"
tun_cidr: "10.77.0.2/24"
tunnel_ip: "10.77.0.2"
gateway: "10.77.0.1"
mtu: "1000"
debug: "true"
command: "pvpn-client-wintun.exe"
arguments: "-server {server} -route {route} -tun-name {tun_name} {debug_arg}"
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	args := cfg.ExpandArguments()
	for _, want := range []string{"-server 72.56.68.200:9443", "-route 1.1.1.1/32", "-tun-name ITSProto", "-debug"} {
		if !strings.Contains(args, want) {
			t.Fatalf("expanded args %q do not contain %q", args, want)
		}
	}
}
