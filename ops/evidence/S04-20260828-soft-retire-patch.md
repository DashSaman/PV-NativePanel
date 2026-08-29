# S04 auth soft-retire patch evidence

- Stage: `S04-AUTH`
- Branch: `s04-auth`
- Auth migration patch commit: `7dfbc51c2e7f232838b9accf11f3ea21e96acf08`
- Migration checksum after patch: `e871db3b3cbfeececbd1b62986e44e312ca409746b50859d9220550c1c356cce`
- One-shot helper workflow removed itself after committing the patch.
- The patch replaces destructive MFA DELETE operations in the non-destructive up migration with auditable soft-retirement (`disabled_at` / `used_at`) while keeping the destructive-migration scanner unchanged.
- CI must pass on a normal repository-owner commit before this evidence is considered a release gate.
