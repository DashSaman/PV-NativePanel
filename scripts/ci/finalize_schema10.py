from pathlib import Path


def replace_once(path: str, old: str, new: str) -> None:
    p = Path(path)
    text = p.read_text()
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"{path}: expected one match, found {count}: {old!r}")
    p.write_text(text.replace(old, new, 1))


# Latest-schema migration regression.
replace_once(
    "tests/db/migration_test.sh",
    "for version in 0001 0002 0003 0004 0005 0006 0007 0008 0009; do",
    "for version in 0001 0002 0003 0004 0005 0006 0007 0008 0009 0010; do",
)
replace_once(
    "tests/db/migration_test.sh",
    "grep -Fqx 'PVNAIVE_SCHEMA_VERSION=9' <<< \"${reapply_output}\"",
    "grep -Fqx 'PVNAIVE_SCHEMA_VERSION=10' <<< \"${reapply_output}\"",
)
replace_once(
    "tests/db/migration_test.sh",
    "# Destructive SQL in the next contiguous migration (10) must fail closed.",
    "# Destructive SQL in the next contiguous migration (11) must fail closed.",
)
for old, new in [
    ("-- pvnaive:migration-version 0010' \\", "-- pvnaive:migration-version 0011' \\") ,
    ("0010_forbidden_drop.up.sql", "0011_forbidden_drop.up.sql"),
    ("0010_forbidden_drop.down.sql", "0011_forbidden_drop.down.sql"),
    ("# An unlisted next migration (10) must fail checksum-manifest validation.", "# An unlisted next migration (11) must fail checksum-manifest validation."),
    ("0010_unlisted_file.up.sql", "0011_unlisted_file.up.sql"),
    ("0010_unlisted_file.down.sql", "0011_unlisted_file.down.sql"),
    ("# A version gap from 9 to 11 must fail before executing SQL.", "# A version gap from 10 to 12 must fail before executing SQL."),
    ("-- pvnaive:migration-version 0011' \\", "-- pvnaive:migration-version 0012' \\") ,
    ("0011_version_gap.up.sql", "0012_version_gap.up.sql"),
    ("0011_version_gap.down.sql", "0012_version_gap.down.sql"),
    ('expected_checksum="$(sha256sum "${repo_root}/db/migrations/0009_direct_naive_exact_accounting.up.sql" | awk \'{print $1}\')"', 'expected_checksum="$(sha256sum "${repo_root}/db/migrations/0010_pending_reservation_completeness.up.sql" | awk \'{print $1}\')"'),
    ('WHERE version=9', 'WHERE version=10'),
    ('for expected in 8 7 6 5 4 3 2 1; do', 'for expected in 9 8 7 6 5 4 3 2 1; do'),
]:
    replace_once("tests/db/migration_test.sh", old, new)

# The two migration-version header replacements above occur in two separate
# fixture blocks. The first replacement consumes the forbidden-drop header;
# update the unlisted-file header explicitly after that.
replace_once(
    "tests/db/migration_test.sh",
    "'-- pvnaive:migration-version 0010' \\\n  '-- pvnaive:migration-name unlisted_file'",
    "'-- pvnaive:migration-version 0011' \\\n  '-- pvnaive:migration-name unlisted_file'",
)

# Customer lifecycle uses the newest repository schema, then proves every
# destructive down migration in sequence.
replace_once(
    "tests/db/customer_lifecycle_migration_test.sh",
    '[[ "${version}" == "9" ]] || { echo "ERROR: schema version=${version}, want=9" >&2; exit 1; }',
    '[[ "${version}" == "10" ]] || { echo "ERROR: schema version=${version}, want=10" >&2; exit 1; }',
)
replace_once(
    "tests/db/customer_lifecycle_migration_test.sh",
    "# Exercise each destructive rollback in order: v9 -> v8 -> v7 -> v6 -> v5 -> v4 -> v3.",
    "# Exercise each destructive rollback in order: v10 -> v9 -> v8 -> v7 -> v6 -> v5 -> v4 -> v3.",
)
replace_once(
    "tests/db/customer_lifecycle_migration_test.sh",
    "for want in 8 7 6 5 4 3; do",
    "for want in 9 8 7 6 5 4 3; do",
)

# Direct accounting must land on schema 10 and explicitly prove the wrapper
# semantics independently of later sequence-conflict degradation.
replace_once(
    "tests/db/direct_naive_accounting_pg18_test.sh",
    '[[ "${schema}" == 9 ]] || { echo "ERROR: expected schema 9, got ${schema}" >&2; exit 1; }',
    '[[ "${schema}" == 10 ]] || { echo "ERROR: expected schema 10, got ${schema}" >&2; exit 1; }',
)
anchor = """[[ \"${not_started}\" == t ]] || { echo 'ERROR: authorize started first-use' >&2; exit 1; }\n\n# Sequence 1 is emitted only after authenticated CONNECT + successful target dial.\n"""
proof = """[[ \"${not_started}\" == t ]] || { echo 'ERROR: authorize started first-use' >&2; exit 1; }\n\n# Schema 10 truthfulness: any pending reservation is immediately incomplete.\npsql_admin --dbname \"${test_db}\" --command \"UPDATE pvnaive.direct_naive_accounting_terms SET reserved_bytes=1 WHERE service_term_id='cccccccc-cccc-cccc-cccc-cccccccccccc';\" >/dev/null\npending_auth=\"$(psql_admin --dbname \"${test_db}\" --tuples-only --no-align --command \"SET ROLE pvnaive_app; SELECT concat_ws('|',remaining_bytes::text,accounting_complete::text) FROM pvnaive.direct_naive_accounting_authorize('11111111-1111-1111-1111-111111111111','2026-08-29T18:00:00Z');\")\"\n[[ \"${pending_auth##*$'\\n'}\" == '99|false' ]] || { echo \"ERROR: schema10 pending authorize: ${pending_auth}\" >&2; exit 1; }\npending_read=\"$(psql_admin --dbname \"${test_db}\" --tuples-only --no-align --command \"SET ROLE pvnaive_app; SELECT concat_ws('|',remaining_bytes::text,accounting_complete::text) FROM pvnaive.direct_naive_accounting_read('cccccccc-cccc-cccc-cccc-cccccccccccc','2026-08-29T18:00:00Z',90);\")\"\n[[ \"${pending_read##*$'\\n'}\" == '99|false' ]] || { echo \"ERROR: schema10 pending read: ${pending_read}\" >&2; exit 1; }\npsql_admin --dbname \"${test_db}\" --command \"UPDATE pvnaive.direct_naive_accounting_terms SET reserved_bytes=0 WHERE service_term_id='cccccccc-cccc-cccc-cccc-cccccccccccc';\" >/dev/null\ncomplete_again=\"$(psql_admin --dbname \"${test_db}\" --tuples-only --no-align --command \"SET ROLE pvnaive_app; SELECT accounting_complete FROM pvnaive.direct_naive_accounting_authorize('11111111-1111-1111-1111-111111111111','2026-08-29T18:00:00Z');\")\"\n[[ \"${complete_again##*$'\\n'}\" == t ]] || { echo \"ERROR: schema10 settled completeness: ${complete_again}\" >&2; exit 1; }\n\n# Sequence 1 is emitted only after authenticated CONNECT + successful target dial.\n"""
replace_once("tests/db/direct_naive_accounting_pg18_test.sh", anchor, proof)

# The shared setter now follows the actual migration directory. Historical stage
# tests keep their frozen successful values but reject one version beyond latest.
replace_once(
    "tests/stages/S04_db_env_version_test.sh",
    "PVNAIVE_DB_ENV_FILE=\"${env_file}\" bash \"${repo_root}/scripts/db/set-expected-schema-version.sh\" 8 >/dev/null\ngrep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=8' \"${env_file}\"\n[[ \"$(grep -c '^PVNAIVE_EXPECTED_SCHEMA_VERSION=' \"${env_file}\")\" == \"1\" ]]\n\nif PVNAIVE_DB_ENV_FILE=\"${env_file}\" bash \"${repo_root}/scripts/db/set-expected-schema-version.sh\" 9 >/dev/null 2>&1; then\n  echo 'ERROR: unsupported schema version was accepted' >&2\n  exit 1\nfi\n\ngrep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=8' \"${env_file}\"",
    "PVNAIVE_DB_ENV_FILE=\"${env_file}\" bash \"${repo_root}/scripts/db/set-expected-schema-version.sh\" 10 >/dev/null\ngrep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=10' \"${env_file}\"\n[[ \"$(grep -c '^PVNAIVE_EXPECTED_SCHEMA_VERSION=' \"${env_file}\")\" == \"1\" ]]\n\nif PVNAIVE_DB_ENV_FILE=\"${env_file}\" bash \"${repo_root}/scripts/db/set-expected-schema-version.sh\" 11 >/dev/null 2>&1; then\n  echo 'ERROR: unsupported schema version was accepted' >&2\n  exit 1\nfi\n\ngrep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=10' \"${env_file}\"",
)
replace_once(
    "tests/stages/S05_upgrade_contract_test.sh",
    'if PVNAIVE_DB_ENV_FILE="${setter_env}" bash "${expected_schema_setter}" 9 >/dev/null 2>&1; then\n  echo \'ERROR: expected schema setter accepted unsupported schema version 9\' >&2',
    'if PVNAIVE_DB_ENV_FILE="${setter_env}" bash "${expected_schema_setter}" 11 >/dev/null 2>&1; then\n  echo \'ERROR: expected schema setter accepted unsupported schema version 11\' >&2',
)

print("WS1_SCHEMA10_PATCH=PASSED")
