---
description: Cheap pre-commit review of the current diff
---
Delegate to the `code-reviewer` subagent. Then run `test-runner` for the touched Go packages only. Return the findings list and PASS/FAIL, nothing else.
