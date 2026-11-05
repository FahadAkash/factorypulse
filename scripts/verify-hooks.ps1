# Smoke-tests every hook with fake input. Run from the repo root.
$ErrorActionPreference = 'Continue'
$root = (Get-Location).Path
$env:CLAUDE_PROJECT_DIR = $root
$h = Join-Path $root '.claude\hooks'
function Run($script, $json) {
  $json | powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $h $script)
  return $LASTEXITCODE
}
# 1. path guard must BLOCK (exit 2) and ALLOW (exit 0)
$bad  = '{"tool_name":"Read","tool_input":{"file_path":"simulator/FactoryPulseSim/Library/x.dll"}}'
$good = '{"tool_name":"Read","tool_input":{"file_path":"CLAUDE.md","limit":20}}'
"guard-paths block : exit " + (Run 'guard-paths.ps1' $bad)  + " (expect 2)"
"guard-paths allow : exit " + (Run 'guard-paths.ps1' $good) + " (expect 0)"
# 2. bash guard
"guard-bash block  : exit " + (Run 'guard-bash.ps1' '{"tool_input":{"command":"go test ./... -v"}}') + " (expect 2)"
"guard-bash allow  : exit " + (Run 'guard-bash.ps1' '{"tool_input":{"command":"go test ./internal/ingest -run TestDup"}}') + " (expect 0)"
# 3. cost tracker with a synthetic transcript
$tmp = Join-Path $env:TEMP 'fp-verify-transcript.jsonl'
$l1 = '{"type":"assistant","message":{"id":"m1","model":"claude-sonnet-4-5","usage":{"input_tokens":1000,"output_tokens":200,"cache_read_input_tokens":5000,"cache_creation_input_tokens":800}}}'
$l2 = '{"type":"assistant","message":{"id":"m2","model":"claude-haiku-4-5","usage":{"input_tokens":400,"output_tokens":100,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}'
Set-Content -Path $tmp -Value @($l1, $l2) -Encoding UTF8
$j = (@{ session_id = 'verify-0001'; transcript_path = $tmp; hook_event_name = 'Stop'; cwd = $root } | ConvertTo-Json -Compress)
"cost-tracker      : exit " + (Run 'cost-tracker.ps1' $j) + " (expect 0)"
if (Test-Path (Join-Path $root 'TOKEN_COST_REPORT.md')) { "report created: TOKEN_COST_REPORT.md" } else { "REPORT MISSING - see .claude\cost\tracker-errors.log" }
"NOTE: verify-0001 is a fake session; delete its entry from .claude\cost\ledger.json and reset docs\token-usage\usage-log.csv + TOKEN_COST_REPORT.md afterwards (run scripts/reset-cost.ps1)."
