$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$exampleRoot = Join-Path $root "examples\service-template"
$sourceRoot = Join-Path $exampleRoot "source"
$distRoot = Join-Path $exampleRoot "dist"
$stagingRoot = Join-Path $distRoot "echo-service-win32"
$zipPath = Join-Path $distRoot "echo-service-win32.zip"

if (-not (Test-Path $sourceRoot)) {
    throw "Missing example source directory: $sourceRoot"
}

if (Test-Path $stagingRoot) {
    Remove-Item $stagingRoot -Recurse -Force
}

if (Test-Path $zipPath) {
    Remove-Item $zipPath -Force
}

New-Item -ItemType Directory -Force -Path $distRoot | Out-Null
New-Item -ItemType Directory -Force -Path $stagingRoot | Out-Null

Copy-Item (Join-Path $sourceRoot '*') $stagingRoot -Recurse -Force
Compress-Archive -Path (Join-Path $stagingRoot '*') -DestinationPath $zipPath -CompressionLevel Optimal

Write-Host "Created example artifact: $zipPath"
