# Token + cost tracker. Runs on Stop, SubagentStop, PreCompact, SessionEnd.
# Re-reads the session transcript, dedupes by message id, computes cost per model,
# upserts the ledger, appends a delta row to usage-log.csv and REWRITES TOKEN_COST_REPORT.md.
# Never blocks Claude: any error is logged and the hook exits 0.
. "$PSScriptRoot\lib.ps1"
$in = Read-HookInput
if (-not $in -or -not $in.transcript_path -or -not (Test-Path -LiteralPath $in.transcript_path)) { exit 0 }
$root   = Get-ProjectRoot $in
$sid    = [string]$in.session_id
$evt  = [string]$in.hook_event_name
$costDir = Join-Path $root '.claude\cost'
$errLog  = Join-Path $costDir 'tracker-errors.log'
$pricing = Read-JsonFile (Join-Path $costDir 'pricing.json') $null
if (-not $pricing) { exit 0 }
$budget  = Read-JsonFile (Join-Path $costDir 'budget.json') $null

function Get-Rate($model) {
  $m = $model.ToLowerInvariant()
  foreach ($k in @('opus', 'sonnet', 'haiku')) { if ($m.Contains($k)) { return $pricing.per_million.$k } }
  return $pricing.per_million.($pricing.default)
}

$mutex = New-Object System.Threading.Mutex($false, 'Local\FactoryPulseCostTracker')
$locked = $false
try {
  $locked = $mutex.WaitOne(15000)
  if (-not $locked) { exit 0 }

  # ---- 1. collect usage from main transcript (+ subagent transcripts if present)
  $files = @([string]$in.transcript_path)
  $sub = Join-Path (Join-Path (Split-Path -Parent $in.transcript_path) $sid) 'subagents'
  if (Test-Path -LiteralPath $sub) {
    $files += @(Get-ChildItem -LiteralPath $sub -Filter '*.jsonl' -File -Recurse | ForEach-Object { $_.FullName })
  }
  $msgs = @{}; $n = 0
  foreach ($f in $files) {
    foreach ($line in (Read-LinesShared $f)) {
      if ($line.IndexOf('"usage"') -lt 0) { continue }
      try { $o = $line | ConvertFrom-Json } catch { continue }
      $m = $o.message
      if (-not $m -or -not $m.usage) { continue }
      $model = [string]$m.model
      if (-not $model -or $model.StartsWith('<')) { continue }
      $id = [string]$m.id
      if (-not $id) { $n++; $id = "line-$n-$f" }
      $u = $m.usage
      $rec = @{ model = $model; i = [long]$u.input_tokens; o = [long]$u.output_tokens; cr = [long]$u.cache_read_input_tokens; cw = [long]$u.cache_creation_input_tokens }
      if (-not $msgs.ContainsKey($id) -or $rec.o -ge $msgs[$id].o) { $msgs[$id] = $rec }
    }
  }

  # ---- 2. aggregate per model and price
  $agg = @{}
  foreach ($r in $msgs.Values) {
    if (-not $agg.ContainsKey($r.model)) { $agg[$r.model] = @{ i = 0L; o = 0L; cr = 0L; cw = 0L } }
    $a = $agg[$r.model]; $a.i += $r.i; $a.o += $r.o; $a.cr += $r.cr; $a.cw += $r.cw
  }
  $modelRows = @(); $tot = @{ i = 0L; o = 0L; cr = 0L; cw = 0L; cost = 0.0 }
  foreach ($k in ($agg.Keys | Sort-Object)) {
    $a = $agg[$k]; $p = Get-Rate $k
    $c = ($a.i * [double]$p.input + $a.o * [double]$p.output + $a.cr * [double]$p.cache_read + $a.cw * [double]$p.cache_write) / 1000000.0
    $modelRows += [pscustomobject]@{ model = $k; input = $a.i; output = $a.o; cache_read = $a.cr; cache_write = $a.cw; cost_usd = [math]::Round($c, 6) }
    $tot.i += $a.i; $tot.o += $a.o; $tot.cr += $a.cr; $tot.cw += $a.cw; $tot.cost += $c
  }

  # ---- 3. upsert ledger
  $ledgerPath = Join-Path $costDir 'ledger.json'
  $ledger = Read-JsonFile $ledgerPath $null
  if (-not $ledger) { $ledger = [pscustomobject]@{ sessions = @(); days = @() } }
  $sessions = As-Array $ledger.sessions
  $days = As-Array $ledger.days
  $now = Get-Date
  $nowStr = $now.ToString('yyyy-MM-dd HH:mm:ss', $script:Inv)
  $today = $now.ToString('yyyy-MM-dd', $script:Inv)

  $idx = -1
  for ($k = 0; $k -lt $sessions.Count; $k++) { if ([string]$sessions[$k].session_id -eq $sid) { $idx = $k; break } }
  $prev = $null; if ($idx -ge 0) { $prev = $sessions[$idx] }
  $pI = 0L; $pO = 0L; $pCr = 0L; $pCw = 0L; $pC = 0.0; $started = $nowStr
  if ($prev) { $pI = [long]$prev.input; $pO = [long]$prev.output; $pCr = [long]$prev.cache_read; $pCw = [long]$prev.cache_write; $pC = [double]$prev.cost_usd; $started = [string]$prev.started }

  $dI = [math]::Max(0L, $tot.i - $pI); $dO = [math]::Max(0L, $tot.o - $pO)
  $dCr = [math]::Max(0L, $tot.cr - $pCr); $dCw = [math]::Max(0L, $tot.cw - $pCw)
  $dC = [math]::Max(0.0, $tot.cost - $pC)

  $rec = [pscustomobject]@{
    session_id = $sid; started = $started; updated = $nowStr; last_event = $evt
    input = $tot.i; output = $tot.o; cache_read = $tot.cr; cache_write = $tot.cw
    cost_usd = [math]::Round($tot.cost, 6); models = $modelRows
  }
  if ($idx -ge 0) { $sessions[$idx] = $rec } else { $sessions += $rec }

  if ($dC -gt 0 -or $dI -gt 0 -or $dO -gt 0) {
    $dayIdx = -1
    for ($k = 0; $k -lt $days.Count; $k++) { if ([string]$days[$k].date -eq $today) { $dayIdx = $k; break } }
    if ($dayIdx -ge 0) {
      $days[$dayIdx].cost_usd = [math]::Round([double]$days[$dayIdx].cost_usd + $dC, 6)
      $days[$dayIdx].tokens = [long]$days[$dayIdx].tokens + $dI + $dO + $dCr + $dCw
    } else {
      $days += [pscustomobject]@{ date = $today; cost_usd = [math]::Round($dC, 6); tokens = ($dI + $dO + $dCr + $dCw) }
    }
  }
  $ledger = [pscustomobject]@{ sessions = $sessions; days = $days }
  Write-Utf8 $ledgerPath ($ledger | ConvertTo-Json -Depth 8)

  # ---- 4. project totals
  $gI = 0L; $gO = 0L; $gCr = 0L; $gCw = 0L; $gC = 0.0; $byModel = @{}
  foreach ($s in $sessions) {
    $gI += [long]$s.input; $gO += [long]$s.output; $gCr += [long]$s.cache_read; $gCw += [long]$s.cache_write; $gC += [double]$s.cost_usd
    foreach ($mr in (As-Array $s.models)) {
      $mk = [string]$mr.model
      if (-not $byModel.ContainsKey($mk)) { $byModel[$mk] = @{ i = 0L; o = 0L; cr = 0L; cw = 0L; c = 0.0 } }
      $b = $byModel[$mk]; $b.i += [long]$mr.input; $b.o += [long]$mr.output; $b.cr += [long]$mr.cache_read; $b.cw += [long]$mr.cache_write; $b.c += [double]$mr.cost_usd
    }
  }

  # ---- 5. append delta row to CSV (only when something changed)
  if ($dC -gt 0 -or $dI -gt 0 -or $dO -gt 0) {
    $csv = Join-Path $root 'docs\token-usage\usage-log.csv'
    if (-not (Test-Path -LiteralPath $csv)) {
      Write-Utf8 $csv ("timestamp,session_id,event,delta_input,delta_output,delta_cache_read,delta_cache_write,delta_cost_usd,session_cost_usd,project_total_cost_usd`r`n")
    }
    $row = ($nowStr, $sid, $evt, $dI, $dO, $dCr, $dCw, (Fmt $dC 6), (Fmt $tot.cost 6), (Fmt $gC 6)) -join ','
    [System.IO.File]::AppendAllText($csv, $row + "`r`n", $script:Utf8NoBom)
  }

  # ---- 6. rewrite the human-readable report
  $L = New-Object System.Collections.Generic.List[string]
  $L.Add('# FactoryPulse - Token & Cost Report')
  $L.Add('')
  $L.Add("> Auto-generated by ``.claude/hooks/cost-tracker.ps1`` after every turn. Do not edit. Last updated: $nowStr (event: $evt)")
  $L.Add('')
  $L.Add('## Project total')
  $L.Add('')
  $L.Add('| Metric | Value |'); $L.Add('|---|---:|')
  $L.Add("| **Estimated cost (USD)** | **`$$(Fmt $gC 4)** |")
  $L.Add("| Input tokens | $(FmtInt $gI) |"); $L.Add("| Output tokens | $(FmtInt $gO) |")
  $L.Add("| Cache-read tokens | $(FmtInt $gCr) |"); $L.Add("| Cache-write tokens | $(FmtInt $gCw) |")
  $L.Add("| Total tokens | $(FmtInt ($gI + $gO + $gCr + $gCw)) |"); $L.Add("| Sessions | $($sessions.Count) |")
  $L.Add('')
  $L.Add('## By model'); $L.Add('')
  $L.Add('| Model | Input | Output | Cache read | Cache write | Cost (USD) |'); $L.Add('|---|---:|---:|---:|---:|---:|')
  foreach ($k in ($byModel.Keys | Sort-Object)) { $b = $byModel[$k]; $L.Add("| $k | $(FmtInt $b.i) | $(FmtInt $b.o) | $(FmtInt $b.cr) | $(FmtInt $b.cw) | `$$(Fmt $b.c 4) |") }
  $L.Add('')
  $L.Add('## By day (last 14)'); $L.Add('')
  $L.Add('| Date | Tokens | Cost (USD) |'); $L.Add('|---|---:|---:|')
  foreach ($d in ($days | Sort-Object { [string]$_.date } -Descending | Select-Object -First 14)) { $L.Add("| $($d.date) | $(FmtInt ([double]$d.tokens)) | `$$(Fmt ([double]$d.cost_usd) 4) |") }
  $L.Add('')
  $L.Add('## Recent sessions (last 15)'); $L.Add('')
  $L.Add('| Session | Started | Updated | Total tokens | Cost (USD) |'); $L.Add('|---|---|---|---:|---:|')
  foreach ($s in ($sessions | Sort-Object { [string]$_.updated } -Descending | Select-Object -First 15)) {
    $short = ([string]$s.session_id); if ($short.Length -gt 8) { $short = $short.Substring(0, 8) }
    $L.Add("| $short | $($s.started) | $($s.updated) | $(FmtInt ([double]$s.input + [double]$s.output + [double]$s.cache_read + [double]$s.cache_write)) | `$$(Fmt ([double]$s.cost_usd) 4) |")
  }
  if ($budget) {
    $dayCost = 0.0; foreach ($d in $days) { if ([string]$d.date -eq $today) { $dayCost = [double]$d.cost_usd } }
    $L.Add(''); $L.Add('## Budget status'); $L.Add('')
    $L.Add("- Today: `$$(Fmt $dayCost 4) of `$$(Fmt ([double]$budget.daily_usd) 2) daily budget")
    $L.Add("- This session: `$$(Fmt $tot.cost 4) of `$$(Fmt ([double]$budget.session_usd) 2) session budget")
  }
  $L.Add(''); $L.Add('## Pricing assumptions'); $L.Add('')
  $L.Add([string]$pricing.note)
  $L.Add(''); $L.Add('Per-turn history: `docs/token-usage/usage-log.csv`')
  Write-Utf8 (Join-Path $root 'TOKEN_COST_REPORT.md') (($L -join "`r`n") + "`r`n")
}
catch {
  try { [System.IO.File]::AppendAllText($errLog, ((Get-Date).ToString('s') + ' ' + $_.Exception.Message + "`r`n"), $script:Utf8NoBom) } catch {}
}
finally {
  if ($locked) { try { $mutex.ReleaseMutex() } catch {} }
}
exit 0
