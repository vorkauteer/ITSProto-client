# Windows Client Roadmap

## v0.1 GUI launcher

Small OpenVPN-sized window, manual config, pinned server key, backend process launcher.

## v0.2 Windows transport smoke-test

Bundle Windows `pvpn-client.exe` and expose handshake/auth/data echo status in UI.

## v0.3 Native Wintun MVP

Implement Wintun adapter creation, route to `1.1.1.1/32`, and encrypted Data frames to the ITSProto server.

## v0.4 Full tunnel

Default route capture, DNS setting, route restore, crash-safe cleanup.

## v0.5 Product shell

Tray icon, update channel, branded UI, support links, signed release artifact.
