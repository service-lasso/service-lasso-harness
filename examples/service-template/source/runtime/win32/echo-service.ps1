param(
  [string]$Message = $env:ECHO_MESSAGE
)

if ([string]::IsNullOrWhiteSpace($Message)) {
  $Message = "hello from service-lasso-harness"
}

Write-Host $Message
