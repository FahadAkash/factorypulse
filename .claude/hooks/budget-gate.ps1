# UserPromptSubmit: warns (or optionally blocks) when daily/session budget is near/over.
. "$PSScriptRoot\lib.ps1"
$in = Read-HookInput; if (-not $in) { exit 0 }
$root = Get-ProjectRoot $in
$b = Read-JsonFile (Join-Path $root '.claude\cost\budget.json') $null
$l = Read-JsonFile (Join-Path $root '.claude\cost\ledger.json') $null
if (-not $b -or -not $l) { exit 0 }
$today = (Get-Date).ToString('yyyy-MM-dd', $script:Inv)
$dayCost = 0.0; $sesCost = 0.0
foreach ($d in (As-Array $l.days)) { if ($d -and [string]$d.date -eq $today) { $dayCost = [double]$d.cost_usd } }
foreach ($s in (As-Array $l.sessions)) { if ($s -and [string]$s.session_id -eq [string]$in.session_id) { $sesCost = [double]$s.cost_usd } }
$dayMax = [double]$b.daily_usd; $sesMax = [double]$b.session_usd; $w = [double]$b.warn_ratio
$over = ($dayCost -ge $dayMax) -or ($sesCost -ge $sesMax)
$warn = ($dayCost -ge $dayMax * $w) -or ($sesCost -ge $sesMax * $w)
$prompt = [string]$in.prompt
$status = "today `$$(Fmt $dayCost 2)/`$$(Fmt $dayMax 2), session `$$(Fmt $sesCost 2)/`$$(Fmt $sesMax 2)"
if ($over) {
  if ($b.hard_block -and ($prompt -notmatch '#budget-ok')) {
    $o = @{ decision = 'block'; reason = "Budget reached ($status). Add #budget-ok to your prompt to continue, or raise limits in .claude/cost/budget.json." } | ConvertTo-Json -Compress
    Write-Output $o; exit 0
  }
  Write-Output "[BUDGET] Limit reached ($status). TERSE MODE: no subagents, no exploration, replies <=3 lines; ask the user before any multi-step work."
} elseif ($warn) {
  Write-Output "[BUDGET] Near limit ($status). Prefer minimal reads, narrow tests, short replies."
}
exit 0
