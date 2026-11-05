# PostToolUse (Write|Edit|MultiEdit): gofmt Go files immediately; surface syntax errors to Claude.
. "$PSScriptRoot\lib.ps1"
$ErrorActionPreference = 'Continue'
$in = Read-HookInput; if (-not $in -or -not $in.tool_input) { exit 0 }
$f = [string]$in.tool_input.file_path
if (-not $f -or ($f -notmatch '\.go$') -or -not (Test-Path -LiteralPath $f)) { exit 0 }
if (-not (Get-Command gofmt -ErrorAction SilentlyContinue)) { exit 0 }
$out = & gofmt -w $f 2>&1 | Out-String
if ($LASTEXITCODE -ne 0) {
  $short = (($out -split "`r?`n") | Select-Object -First 8) -join "`n"
  Block "gofmt failed in ${f}:`n$short"
}
exit 0
