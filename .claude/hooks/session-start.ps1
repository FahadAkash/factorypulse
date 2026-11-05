# SessionStart: inject a tiny (<=8 lines) context block. Stdout is added to Claude's context.
. "$PSScriptRoot\lib.ps1"
$in = Read-HookInput; $root = Get-ProjectRoot $in
$ErrorActionPreference = 'Continue'
$out = @('FactoryPulse session (auto-context, keep replies terse):')
if (Get-Command git -ErrorAction SilentlyContinue) {
  $br = (& git -C $root rev-parse --abbrev-ref HEAD 2>$null)
  if ($br) { $out += "- branch: $br; uncommitted files: $((& git -C $root status --porcelain 2>$null | Measure-Object).Count)" }
}
$rm = Join-Path $root 'docs\ROADMAP.md'
if (Test-Path -LiteralPath $rm) {
  $next = Select-String -LiteralPath $rm -Pattern '^\s*- \[ \]' | Select-Object -First 1
  if ($next) { $out += "- next milestone: " + ($next.Line.Trim() -replace '^- \[ \]\s*', '') }
}
$l = Read-JsonFile (Join-Path $root '.claude\cost\ledger.json') $null
if ($l) {
  $today = (Get-Date).ToString('yyyy-MM-dd', $script:Inv)
  foreach ($d in (As-Array $l.days)) { if ($d -and [string]$d.date -eq $today) { $out += "- spend today: `$$(Fmt ([double]$d.cost_usd) 2) (see TOKEN_COST_REPORT.md)" } }
}
$out += '- Rules: 1 milestone/session; Grep before Read; haiku subagents for search/tests; answer <=8 lines.'
Write-Output ($out -join "`n")
exit 0
