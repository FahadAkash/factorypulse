# PreToolUse (Read|Grep|Glob): block token-wasting paths. Exit 2 = block, stderr goes to Claude.
. "$PSScriptRoot\lib.ps1"
$in = Read-HookInput; if (-not $in -or -not $in.tool_input) { exit 0 }
$ti = $in.tool_input; $tool = [string]$in.tool_name
$cands = @()
foreach ($k in @('file_path', 'path', 'glob')) { if ($ti.$k) { $cands += [string]$ti.$k } }
if ($tool -eq 'Glob' -and $ti.pattern) { $cands += [string]$ti.pattern }

$denyCase = '(^|/)(Library|Temp|Logs|Builds|UserSettings|PackageCache)(/|$)'
$denyAny  = '(^|/)(node_modules|obj|bin|\.git|dist|coverage|\.next|pgdata|__pycache__|vendor)(/|$)'
$denyExt  = '\.(meta|dll|exe|pdb|png|jpe?g|gif|zip|7z|fbx|mp4|wav|asset|prefab|unity|mat|anim|controller|lock|sum)$'
$denyEnv  = '(^|/)\.env(\.(?!example$).+)?$'

$rootN = ((Get-ProjectRoot $in) -replace '\\', '/').TrimEnd('/')
foreach ($c in $cands) {
  $p = $c -replace '\\', '/'
  if ($rootN -and $p.StartsWith($rootN, [System.StringComparison]::OrdinalIgnoreCase)) { $p = $p.Substring($rootN.Length) }
  if (($p -cmatch $denyCase) -or ($p -imatch $denyAny) -or ($p -imatch $denyExt) -or ($p -imatch $denyEnv)) {
    Block "Blocked by token guard: '$c' is generated/binary/secret/large. Use Grep/Glob on source dirs (backend/, web/, contracts/, simulator/**/Assets/**/*.cs). For Unity scenes/prefabs ask the user."
  }
}
# Whole-file reads of big files are expensive: require offset/limit.
if ($tool -eq 'Read' -and $ti.file_path -and -not $ti.limit -and -not $ti.offset) {
  $fp = [string]$ti.file_path
  if (Test-Path -LiteralPath $fp -PathType Leaf) {
    $len = (Get-Item -LiteralPath $fp).Length
    if ($len -gt 200KB) { Block "Blocked: '$fp' is $([int]($len/1KB)) KB. Use Grep -n to locate, then Read with offset/limit (<=150 lines)." }
  }
}
exit 0
