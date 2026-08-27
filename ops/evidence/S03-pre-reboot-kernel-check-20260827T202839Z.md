# S03 pre-reboot kernel check

Observed on `testAmir5-3` at `2026-08-27T20:28:39Z`.

- Running kernel: `7.0.0-29-generic`.
- Target kernel package installed: `linux-image-7.0.0-30-generic` version `7.0.0-30.30`.
- `linux-base` installed: `4.15ubuntu5`.
- `/boot/vmlinuz-7.0.0-30-generic` present.
- `/boot/initrd.img-7.0.0-30-generic` present.
- GRUB contains normal and recovery entries for `7.0.0-30-generic`.
- `GRUB_DEFAULT=0`.
- `caddy-naive.service` is enabled and currently active.
- `ssh.service` is disabled, but `ssh.socket` is enabled; SSH is currently active.
- `/var/run/reboot-required` remains present.

Decision: perform a controlled reboot before the real `S03-DATABASE` execution. After reconnecting, verify kernel, SSH/Caddy, Caddyfile checksum, required listeners, reboot flag, and S03 pre-state before running the database stage.

Stage remains `S03-DATABASE=NEXT`; `S04-AUTH` remains blocked until a real `S03_RESULT=PASSED`.
