# ITSProto-client v0.2.0

## Added

- Default profile now starts `pvpn-client-wintun.exe` for real Windows route-mode tunneling.
- Added placeholders for `tun_name`, `tun_cidr`, `tunnel_ip`, `gateway`, `mtu`, `debug_arg`, and `full_tunnel_arg`.
- Updated backend build helper to build both `pvpn-client.exe` and `pvpn-client-wintun.exe` from the ITSProto repository.

## First test target

Use route mode for `1.1.1.1/32` before full-tunnel.
