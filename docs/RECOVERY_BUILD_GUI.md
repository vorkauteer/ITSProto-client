# ITSProto-client recovery and GUI build

This repository contains the small Windows GUI launcher. The VPN backend is built from the ITSProto repository.

## Install GUI config

Open PowerShell as Administrator:

```powershell
cd C:\Users\pozzz\OneDrive\Desktop\itsecurity\vpn\ITSProto-client
New-Item -ItemType Directory -Force C:\ProgramData\ITSProto | Out-Null
Copy-Item .\configs\client.yaml.example C:\ProgramData\ITSProto\client.yaml -Force
notepad C:\ProgramData\ITSProto\client.yaml
```

Set `server_pub` to the public key from the server:

```bash
sudo pvpn-server -config /etc/pvpn/server.yaml -print-pub-key
```

## Build GUI

```powershell
go test ./...
Unblock-File .\scripts\build-windows.ps1
.\scripts\build-windows.ps1
```

Expected file:

```text
dist\itsproto-windows-client.exe
```

Run it as Administrator, then press:

```text
Reload config
Connect
```
