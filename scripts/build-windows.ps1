$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path dist | Out-Null

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-H=windowsgui -s -w" -o dist\itsproto-windows-client.exe .\cmd\itsproto-windows-client

Write-Host "Built: dist\itsproto-windows-client.exe"
