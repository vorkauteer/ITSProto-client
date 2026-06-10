$ErrorActionPreference = "Stop"

$targetDir = "C:\ProgramData\ITSProto"
$targetFile = Join-Path $targetDir "client.yaml"

New-Item -ItemType Directory -Force -Path $targetDir | Out-Null

if (Test-Path $targetFile) {
    Write-Host "Kept existing $targetFile"
} else {
    Copy-Item .\configs\client.yaml.example $targetFile
    Write-Host "Installed example config to $targetFile"
}

Write-Host "Edit server_pub before connecting."
