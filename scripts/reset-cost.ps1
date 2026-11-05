# Wipes cost history (ledger, csv, report). Use after verify-hooks or to start a new billing period.
$ErrorActionPreference = 'Continue'
Remove-Item -LiteralPath '.claude\cost\ledger.json' -ErrorAction SilentlyContinue
Remove-Item -LiteralPath 'docs\token-usage\usage-log.csv' -ErrorAction SilentlyContinue
Remove-Item -LiteralPath 'TOKEN_COST_REPORT.md' -ErrorAction SilentlyContinue
"Cost history reset. It will be recreated automatically after the next Claude turn."
