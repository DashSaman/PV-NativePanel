# PVNaive Deployment Progress — testAmir5-3

آخرین به‌روزرسانی: 2026-08-27 23:50 UTC

این فایل ledger رسمی استقرار است. برای جزئیات کامل continuation، `AGENT_HANDOFF.md` را بخوان.

## قواعد

1. فقط Stage فعلی اجرا شود.
2. هر failure با علت و rollback ثبت شود.
3. هیچ Stage بدون اجرای واقعی + postflight مستقل `PASSED` نمی‌شود.
4. Caddy/SSH/Firewall تا gate مشخص‌شده دست‌نخورده بمانند.
5. هر artifact قبل از اجرا SHA-256 verify شود.
6. هر Chat جدید ابتدا `AGENT_HANDOFF.md` و این فایل را بخواند.

## Target

| مورد | مقدار |
|---|---|
| Host | `testAmir5-3` |
| IPv4 | `91.107.182.147` |
| Domain | `namir.softarg.ir` |
| OS | Ubuntu 26.04 LTS |
| PostgreSQL | `18/main`, loopback `5432` |
| Caddy service | `caddy-naive.service` |
| Caddy SHA-256 | `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1` |

## Stage ledger

| ID | Status | نتیجه/Blocker |
|---|---|---|
| S00-NAMING | PASSED | نام محصول `PVNaive` تثبیت شد؛ repo rename نشده |
| S01-PREFLIGHT | PASSED | DNS/TLS/Caddy/listeners/capacity سالم |
| S02-FOUNDATION | PASSED | user/directories/backup پایه ساخته شد |
| S03-DATABASE | PASSED | PostgreSQL 18 + schema v1 + RLS + encrypted backup/restore + health + postflight |
| S04-AUTH | IN PROGRESS | schema v2 روی سرور؛ blocker فعلی startup سرویس `pvnaive-api.service` |
| S05-USERS | BLOCKED | منتظر S04 |
| S06-RUNTIME | BLOCKED | منتظر S05 |
| S07-SUBS | BLOCKED | منتظر S06 |
| S08-NOTIFY | BLOCKED | منتظر S07 |
| S09-INSTALLER | BLOCKED | منتظر S08 |
| S10-PILOT | BLOCKED | منتظر gateهای قبلی |

## S02 evidence

- PASSED: `2026-08-26 20:18:57 UTC`
- backup: `/var/backups/pvnaive/20260826T201857Z`
- user: `pvnaive` UID 995 / GID 982
- `/opt/pvnaive`: `750 root:pvnaive`
- `/etc/pvnaive`: `750 root:pvnaive`
- Caddy/SSH/Firewall unchanged

## S03 FINAL evidence

Production S03 و postflight مستقل هر دو PASS شدند.

- marker: `/opt/pvnaive/S03_DATABASE.json`
- PostgreSQL: `18/main`, loopback-only port `5432`
- database: `pvnaive`
- roles: `pvnaive_owner`, `pvnaive_app`
- schema version: `1`
- application DB health: PASSED
- secret direct SELECT: DENIED
- encrypted backup: PASSED
- restore drill: PASSED
- DB health timer: enabled/active
- Caddy SHA unchanged
- SSH/Firewall unchanged

Final output:

```text
S03_POSTFLIGHT=PASSED
S03_DATABASE=PASSED
NEXT_STAGE=S04-AUTH
```

## S04 code / CI baseline

Active branch: `s04-auth`
PR: `#2`

آخرین production bundle کاملاً verified قبل از live deploy:

- source commit: `b4803e27af36bb35de33f7dcbe39750aeadc4146`
- CI run: `33126898878`
- Go: PASS
- Web test/build: PASS
- PostgreSQL 18: PASS
- end-to-end auth rehearsal: PASS
- bundle: PASS

Bundle:

- `PVNaive-S04-b4803e27af36.tar.gz`
- SHA-256: `c279c91f42ca7d0b91096c48875fa61be8e8c80c8effa826aa477314a9732147`
- extracted server path: `/root/pvnaive-s04-deploy-b4803e27af36/PVNaive-S04-b4803e27af36`

## S04 live attempt log

### Attempt 1 — 23:44 UTC

Failure:

```text
file: command not found
ERROR: pvnaive binary architecture mismatch
S04_RESULT=FAILED
ROLLBACK=COMPLETED
```

Action: package `file` installed. ELF x86-64 validation then passed. No Caddy/SSH/Firewall change.

### Attempt 2 — 23:46 UTC

- pre-S04 schema-v1 encrypted backup created:
  `/var/backups/pvnaive/database/20260827T234607Z/pvnaive.dump.age`
- migration `0002_auth_foundation` applied successfully
- schema reached `2`
- second backup attempted in same second and collided with same destination

Failure:

```text
ERROR: backup destination already exists
S04_RESULT=FAILED
FAILED_LINE=276
ROLLBACK=INCOMPLETE
ROLLBACK_FAILED_STEP=rollback-migration-0002-no-schema2-backup
```

Root cause: backup directory timestamp had second-only precision, so two backups in one second collided.

DB migration intentionally remained v2 because no verified schema-v2 rollback backup existed yet. S04 artifacts were cleaned; infrastructure invariants unchanged.

### Recovery preflight — 23:49:59 UTC

Verified:

```text
SCHEMA_VERSION=2
STORED_0002=0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886
MANIFEST_0002=84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886
SCHEMA_RECOVERY_STATE=PASSED
S04_ARTIFACT_CLEANUP=PASSED
INFRASTRUCTURE_INVARIANTS=PASSED
RECOVERY_PREFLIGHT=PASSED
```

### Attempt 3 — recovery mode

Stage recognized:

```text
RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER
```

schema-v2 encrypted backup created successfully:

`/var/backups/pvnaive/database/20260827T234959Z/pvnaive.dump.age`

Then service startup failed:

```text
Created symlink '/etc/systemd/system/multi-user.target.wants/pvnaive-api.service' → '/etc/systemd/system/pvnaive-api.service'.
ERROR: pvnaive-api.service is not active
S04_RESULT=FAILED
FAILED_LINE=298
ROLLBACK=STARTED
ROLLBACK=COMPLETED
CADDY_ACTION=none
SSH_ACTION=none
FIREWALL_ACTION=none
```

This is the current live blocker.

## Expected current server state after Attempt 3 rollback

Must be re-verified before retry:

- schema `2`
- migration 0002 checksum matches bundle
- no `/opt/pvnaive/S04_AUTH.json`
- no `/etc/pvnaive/auth.key`
- no `/opt/pvnaive/bin/pvnaive`
- no `/opt/pvnaive/bin/pvnaive-password`
- no `/etc/systemd/system/pvnaive-api.service`
- `pvnaive-api.service` inactive
- no listener on `8080`
- Caddy SHA unchanged
- SSH/Caddy/PostgreSQL baseline listeners healthy

## NEXT exact action

Do not retry S04 blindly. First capture service startup evidence from the previous failed boot:

```bash
systemctl status pvnaive-api.service --no-pager -l || true
journalctl -u pvnaive-api.service -b --no-pager -n 200 || true
systemctl show pvnaive-api.service -p Result -p ExecMainStatus -p ExecMainCode -p ActiveState -p SubState || true
systemctl cat pvnaive-api.service || true
namei -l /etc/pvnaive/db.env /etc/pvnaive/auth.key /opt/pvnaive/bin/pvnaive 2>/dev/null || true
ss -lntp | grep -E ':(22|80|443|5432|8080)([[:space:]]|$)' || true
```

از journal root cause را مشخص کن. قبل از evidence هیچ permission/systemd/env fix حدسی نزن.

## backup collision regression status

بعد از live failure، regression test برای collision اضافه شد. Current branch head در زمان ثبت این ledger:

`ab025aebe96af2920f7e2a405b63b1bd9c965ad9`

CI run `33127608010`:

- Go: PASS
- Web: PASS
- Database: FAIL
- Rehearsal: SKIPPED
- Bundle: SKIPPED

پس fix branch هنوز final-green نیست. قبل از ساخت bundle جدید، failure تست collision باید debug و کل pipeline سبز شود.

## Gate خروج از S04

برای `S04=PASSED` همه موارد زیر لازم است:

1. service active
2. listener فقط `127.0.0.1:8080`
3. `/api/v1/health/live` PASS
4. `/api/v1/health/ready` PASS
5. schema v2 + checksum صحیح
6. S04 marker اتمیک ساخته شود
7. encrypted schema-v2 backup checksum PASS
8. Caddy SHA unchanged در localhost phase
9. SSH/Firewall unchanged
10. independent postflight PASS
11. owner bootstrap بعد از localhost stage
12. login/session/logout واقعی PASS
13. سپس Caddy integration جداگانه با backup/validate/reload/rollback

تا قبل از این موارد S05 باز نشود.
