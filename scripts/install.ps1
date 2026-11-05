# One-time Windows setup check. Run: powershell -ExecutionPolicy Bypass -File scripts/install.ps1
$ErrorActionPreference = 'Continue'
function Check($name, $cmd, $hint) {
  if (Get-Command $cmd -ErrorAction SilentlyContinue) { "[ok]   $name" } else { "[MISS] $name  -> $hint" }
}
Check 'Git for Windows (provides Git Bash; REQUIRED for Claude Code hooks)' 'git' 'https://git-scm.com/download/win'
Check 'Go 1.22+'            'go'      'winget install GoLang.Go'
Check 'Docker Desktop'      'docker'  'winget install Docker.DockerDesktop (enable WSL2)'
Check 'Node 20+'            'node'    'winget install OpenJS.NodeJS.LTS'
Check 'Claude Code'         'claude'  'npm install -g @anthropic-ai/claude-code'
Check 'gofmt (ships with Go)' 'gofmt' 'reinstall Go'
if (Get-Command git -ErrorAction SilentlyContinue) {
  git config core.longpaths true
  git config core.autocrlf false
  "[ok]   git core.longpaths=true, core.autocrlf=false"
}
"Unity: install Unity Hub + an LTS editor (2022.3 LTS or Unity 6) manually; open simulator/FactoryPulseSim."
"Next: powershell -File scripts/verify-hooks.ps1"
