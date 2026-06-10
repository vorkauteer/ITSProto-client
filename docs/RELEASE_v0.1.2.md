# v0.1.2 backend ping-only launcher

## Why

The server is now usually run with `tun.enabled: true`. In that mode encrypted `FrameData` carries raw IPv4 packets and is not a synthetic echo channel.

The previous Windows GUI example launched:

```text
pvpn-client.exe ... -echo windows-gui-test
```

That command is only valid against a non-TUN test server. Against a TUN-enabled production-style server it can complete handshake/auth/ping and then time out on data echo.

## Change

The default GUI config now launches:

```text
pvpn-client.exe ... -ping-only
```

This verifies server reachability, pinned identity, auth and encrypted ping/pong without sending synthetic `FrameData`.
