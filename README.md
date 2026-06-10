# ITSProto Windows Client

Minimal Windows desktop client for ITSProto.

Current scope: small native Win32 GUI with one Connect/Disconnect workflow, YAML profile loading, manual pinned server public key, and an external backend launcher. This is intentionally separated from the protocol/server repository.

The first Windows client milestone is not a full native Wintun tunnel yet. It is a GUI shell that can start `pvpn-client.exe` for transport smoke-tests today and later start a native `pvpn-client-wintun.exe` backend.

## Repository layout

```text
cmd/itsproto-windows-client/   Native Windows GUI entrypoint
internal/config/               Small dependency-free YAML-like profile parser
internal/runner/               Backend process runner
configs/client.yaml.example    User profile example
scripts/build-windows.ps1      Windows build script
scripts/install-config.ps1     Installs C:\ProgramData\ITSProto\client.yaml
```

## Build

From PowerShell on Windows with Go 1.22+:

```powershell
go test ./...
.\scripts\build-windows.ps1
```

The GUI executable will be created at:

```text
dist\itsproto-windows-client.exe
```

Cross-build from Linux/WSL is also possible:

```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui -s -w" \
  -o dist/itsproto-windows-client.exe ./cmd/itsproto-windows-client
```

## Configure

Install example config:

```powershell
.\scripts\install-config.ps1
notepad C:\ProgramData\ITSProto\client.yaml
```

Get the server public key on the VPS:

```bash
sudo pvpn-server -config /etc/pvpn/server.yaml -print-pub-key
```

Paste it into:

```yaml
server_pub: "..."
```

For the first smoke-test, the GUI needs a backend executable: `pvpn-client.exe` from the protocol repository. The launcher searches for the backend in this order: next to `itsproto-windows-client.exe`, next to `client.yaml`, `C:\ProgramData\ITSProto`, and then `%PATH%`. You can also set an absolute path in `command`.

Build and install the temporary transport-check backend from the protocol repository:

```powershell
.\scripts\build-backend-from-protocol.ps1 -ProtocolRepo C:\path\to\ITSProto
```

This creates:

```text
C:\ProgramData\ITSProto\pvpn-client.exe
```

Example transport-check profile:

```yaml
server: "72.56.68.200:9443"
token: "dev-token"
device_id: "windows-client-1"
server_pub: "PASTE_SERVER_PUBLIC_KEY_HERE"
mode: "transport-check"
route: "1.1.1.1/32"
command: "pvpn-client.exe"
arguments: "-server {server} -token {token} -device {device_id} -server-pub {server_pub} -ping-only"
```

## Run

Start the GUI:

```powershell
.\dist\itsproto-windows-client.exe
```

Optional custom config path:

```powershell
.\dist\itsproto-windows-client.exe -config C:\path\client.yaml
```

Click **Connect**. The log panel should show that the backend started and then print the backend output.

## Current status

Done:

```text
[done] Small native Windows window
[done] Connect / Disconnect / Reload config buttons
[done] C:\ProgramData\ITSProto\client.yaml profile
[done] Manual pinned server public key
[done] External backend process management
[done] No runtime dependencies other than the backend executable
```

Next:

```text
[done] Helper script to build pvpn-client.exe from the protocol repo
[next] Native Wintun backend
[next] Windows route/DNS management
[next] tray icon
[next] signed installer
[next] automatic config delivery from control-plane
```

## Native Windows tunnel backend

For real tunneling the GUI starts:

```text
C:\ProgramData\ITSProto\pvpn-client-wintun.exe
```

Build both protocol backends from the `ITSProto` repository:

```powershell
.\scripts\build-backend-from-protocol.ps1 -ProtocolRepo C:\Users\pozzz\OneDrive\Desktop\itsecurity\vpn\ITSProto -InstallDeps
```

Copy `configs\client.yaml.example` to `C:\ProgramData\ITSProto\client.yaml`, paste the server public key, run the GUI as Administrator, then press `Reload config` and `Connect`.
