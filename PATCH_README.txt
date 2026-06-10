ITSProto-client patch: v0.2.0 Windows GUI tunnel mode

Copy/expand this archive into the root of the ITSProto-client repository.

Changes:
- client.yaml.example now uses pvpn-client-wintun.exe by default
- config parser supports tun_name/tun_cidr/tunnel_ip/gateway/mtu/debug
- backend build helper builds both protocol backends

After applying:
  cd C:\Users\pozzz\OneDrive\Desktop\itsecurity\vpn\ITSProto-client
  go test ./...
  .\scripts\build-windows.ps1
