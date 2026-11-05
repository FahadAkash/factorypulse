# Runs a command and prints ONLY what matters (saves tokens).
# Usage: powershell -File scripts/run-quiet.ps1 "cd backend; go test ./internal/ingest/..."
param([Parameter(Mandatory = $true, Position = 0)][string]$Command, [int]$Tail = 40)
$ErrorActionPreference = 'Continue'
$global:LASTEXITCODE = 0
$out = Invoke-Expression $Command 2>&1 | Out-String
$code = $LASTEXITCODE
$lines = @($out -split "`r?`n" | Where-Object { $_ -ne '' })
if ($code -eq 0) {
  "OK exit=0 (output lines suppressed: $($lines.Count))"
  $lines | Select-Object -Last 3
} else {
  "FAILED exit=$code (total output lines: $($lines.Count))"
  $lines | Select-String -Pattern '^(--- FAIL|FAIL|panic:|.*\.go:\d+)' | Select-Object -First 20 | ForEach-Object { $_.Line }
  '--- tail ---'
  $lines | Select-Object -Last $Tail
}
exit $code
