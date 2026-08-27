# S03 private repository access check — testAmir5-3

Recorded from the operator-provided server output.

- UTC: `2026-08-27T20:12:54Z`
- Host: `testAmir5-3`
- Git: `2.53.0` present
- GitHub CLI (`gh`): not installed
- `GH_TOKEN`: not set
- `GITHUB_TOKEN`: not set
- `/root/.git-credentials`: absent
- `/root/.config/gh/hosts.yml`: absent
- Private repository read test: `GIT_TERMINAL_PROMPT=0 git ls-remote https://github.com/DashSaman/PV-NativePanel.git refs/heads/main`
- Result: failed as expected with `fatal: could not read Username for 'https://github.com': terminal prompts disabled`
- Exit code: `128`

## Decision

Do not place a long-lived GitHub credential or PAT on the pilot server merely to deliver S03. Transfer only the verified S03 bundle out-of-band, verify its SHA-256 and inventory on the server, and then execute the guarded launcher. The obsolete `pvnaive-s03-6d4e5ce.tar.gz` must not be executed.

No server state was changed by this check.
