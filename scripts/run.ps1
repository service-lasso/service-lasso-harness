param(
    [Parameter(Mandatory = $true)]
    [string]$Contract,

    [string]$OutputDir = "output"
)

$ErrorActionPreference = "Stop"

go run ./cmd/service-lasso-harness run --contract $Contract --output-dir $OutputDir
