param(
    [Parameter(Mandatory=$true)]
    [string]$ProtocolRepo,
    [string]$OutputDir = "C:\ProgramData\ITSProto",
    [switch]$InstallDeps
)

$ErrorActionPreference = "Stop"

if (!(Test-Path $ProtocolRepo)) {
    throw "Protocol repo not found: $ProtocolRepo"
}

Push-Location $ProtocolRepo
try {
    New-Item -ItemType Directory -Force $OutputDir | Out-Null

    if ($InstallDeps) {
        go get golang.zx2c4.com/wintun golang.org/x/sys/windows
        go mod tidy
    }

    go build -o (Join-Path $OutputDir "pvpn-client.exe") .\services\protocol-server\cmd\pvpn-client
    go build -o (Join-Path $OutputDir "pvpn-client-wintun.exe") .\services\protocol-server\cmd\pvpn-client-wintun

    Write-Host "Built protocol backends into $OutputDir"
} finally {
    Pop-Location
}
