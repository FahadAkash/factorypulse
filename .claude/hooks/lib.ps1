# Shared helpers. Compatible with Windows PowerShell 5.1 and PowerShell 7.
$ErrorActionPreference = 'Stop'
$script:Utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$script:Inv = [System.Globalization.CultureInfo]::InvariantCulture

function Read-HookInput {
  try {
    $raw = [Console]::In.ReadToEnd()
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return ($raw | ConvertFrom-Json)
  } catch { return $null }
}
function Get-ProjectRoot($in) {
  if ($env:CLAUDE_PROJECT_DIR) { return $env:CLAUDE_PROJECT_DIR }
  if ($in -and $in.cwd) { return [string]$in.cwd }
  return (Get-Location).Path
}
function Write-Utf8($path, $text) {
  $dir = Split-Path -Parent $path
  if ($dir -and -not (Test-Path -LiteralPath $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
  [System.IO.File]::WriteAllText($path, $text, $script:Utf8NoBom)
}
function Read-JsonFile($path, $default) {
  if (-not (Test-Path -LiteralPath $path)) { return $default }
  try { return ((Get-Content -Raw -LiteralPath $path -Encoding UTF8) | ConvertFrom-Json) } catch { return $default }
}
function As-Array($x) { if ($null -eq $x) { return ,@() } else { return ,@($x) } }
function Fmt([double]$n, [int]$d = 4) { return $n.ToString("F$d", $script:Inv) }
function FmtInt([double]$n) { return $n.ToString('N0', $script:Inv) }
function Block($msg) { [Console]::Error.WriteLine($msg); exit 2 }
function Read-LinesShared($path) {
  # Reads a file even while Claude Code is still appending to it.
  $fs = New-Object System.IO.FileStream($path, [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read, [System.IO.FileShare]::ReadWrite)
  $sr = New-Object System.IO.StreamReader($fs, [System.Text.Encoding]::UTF8)
  try { while (-not $sr.EndOfStream) { $sr.ReadLine() } } finally { $sr.Dispose(); $fs.Dispose() }
}
