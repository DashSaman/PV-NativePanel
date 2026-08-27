# S03 hotfix bundle build — package candidate failure

UTC: 2026-08-27T20:39:40Z
Host: `testAmir5-3`

The first S03 SIGPIPE hotfix was applied successfully to the verified 13-file bundle source:

- old S03 Git blob: `f14cc7ecab7aa07c56b0b4f90ccd582c7e039248`
- patched S03 Git blob: `5cbcf71d1302c6d757653b83a1dbf06877e26493`
- patch match count: `1`
- other 12 bundle files matched their expected Git blob SHAs
- migration checksums passed
- shell syntax passed

The build intentionally failed before producing a new bundle because the APT candidate regression gate found:

- `postgresql=18+290ubuntu1`
- `postgresql-client=18+290ubuntu1`
- `postgresql-contrib=(none)`

Result:

- `ERROR=NO_APT_CANDIDATE:postgresql-contrib`
- `HOTFIX_BUNDLE_BUILD=FAILED`
- no package install was performed by this build step
- the temporary work directory and candidate bundle were removed by the build cleanup trap

Follow-up verification against Ubuntu 26.04 package metadata shows that `postgresql-contrib-18` is virtual and is provided by `postgresql-18`; the S03 package list must therefore not require a separate `postgresql-contrib` candidate on Ubuntu 26.04.

S03 remains `NEXT`; do not advance to S04.
