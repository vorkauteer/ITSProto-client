# v0.1.1 GUI backend lookup fix

This release fixes the first Windows smoke-test issue where the GUI reported:

```text
exec: "pvpn-client.exe": executable file not found in %PATH%
```

The GUI is still a launcher over a backend process. It now searches for `pvpn-client.exe` in:

1. the directory containing `itsproto-windows-client.exe`;
2. the directory containing `client.yaml`;
3. `C:\ProgramData\ITSProto`;
4. `%PATH%`.

A helper script was added:

```powershell
.\scripts\build-backend-from-protocol.ps1 -ProtocolRepo C:\path\to\ITSProto
```

It builds `pvpn-client.exe` from the protocol repository and copies it to:

```text
C:\ProgramData\ITSProto\pvpn-client.exe
```
