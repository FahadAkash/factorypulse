# PreToolUse (Bash): block destructive commands and context-flooding output.
. "$PSScriptRoot\lib.ps1"
$in = Read-HookInput; if (-not $in -or -not $in.tool_input) { exit 0 }
$cmd = [string]$in.tool_input.command
if (-not $cmd) { exit 0 }
$rules = @(
  @{ p = '(^|[\s;|&])rm\s+-[a-zA-Z]*r[a-zA-Z]*\s+(/|~|\*|\.\s*$|[A-Za-z]:)';           m = 'Destructive rm blocked.' },
  @{ p = 'Remove-Item[^|;]*-Recurse[^|;]*(\s|^)([A-Za-z]:\\?(\s|$)|\\(\s|$)|/(\s|$)|\*|\.\s*$)'; m = 'Recursive Remove-Item on a root/wildcard is blocked.' },
  @{ p = '\b(rd|rmdir)\s+/s';                                                             m = 'Recursive rd/rmdir blocked.' },
  @{ p = '\bdel\s+[^|;]*/s';                                                              m = 'Recursive del blocked.' },
  @{ p = '\bformat\s+[A-Za-z]:';                                                          m = 'format blocked.' },
  @{ p = 'git\s+push\b[^|;]*(--force|\s-f(\s|$))';                                        m = 'Force push blocked.' },
  @{ p = 'git\s+reset\s+--hard';                                                          m = 'git reset --hard blocked; ask the user.' },
  @{ p = 'git\s+clean\s+-[a-z]*f';                                                        m = 'git clean -f blocked; ask the user.' },
  @{ p = '\bdocker\s+(system|volume)\s+prune';                                            m = 'docker prune blocked; ask the user.' },
  @{ p = '(docker(\s+compose)?|docker-compose)\s+logs(?![^|;]*--tail)';                   m = 'Add --tail 50 (and a service name) to docker logs.' },
  @{ p = '\bgo\s+test\b[^|;]*\s-v(\s|$)';                                                 m = 'go test -v floods context. Use: powershell -File scripts/run-quiet.ps1 "go test ./pkg -run Name".' },
  @{ p = '(^|[\s;|&])(ls\s+-[a-zA-Z]*R|tree(\s|$)|dir\s+/s)';                             m = 'Recursive listing floods context. Use Glob.' },
  @{ p = 'Get-ChildItem(?![^|;]*(-Filter|-Include|-Depth))[^|;]*-Recurse';                m = 'Unfiltered Get-ChildItem -Recurse floods context. Use Glob or add -Filter/-Depth.' },
  @{ p = '\b(cat|type|Get-Content)\s+\S+\.(log|jsonl)\b(?![^;]*(-Tail|tail|Select-Object|head))'; m = 'Never dump logs. Use Get-Content <file> -Tail 40.' },
  @{ p = '\bnpm\s+(install|i)\s+(-g|--global)';                                           m = 'Global npm installs blocked.' }
)
foreach ($r in $rules) {
  if ($cmd -imatch $r.p) { Block "Blocked by bash guard: $($r.m)" }
}
exit 0
