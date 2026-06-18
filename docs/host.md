# Host

Hosting is supported on macOS, Linux, and Windows.

The host opens a real PTY and starts the configured shell inside the selected workspace. PTY resize events come from the guest browser terminal.

macOS and Linux use a real PTY. Windows hosting uses a stdio backend.

Start a long tunnel session:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 1440
```

Start a relay fallback session:

```bash
controli host share --workspace main --minutes 480
```

Permission modes:

```bash
controli host share --workspace main --mode full
controli host share --workspace main --mode view
controli host share --workspace main --mode approve
```

- `full` lets the guest type after the host approves control.
- `view` blocks guest input and only streams output.
- `approve` asks the host before each input chunk is written to the shell.

Useful host flags:

| Flag | Purpose |
| --- | --- |
| `--approve=false` | Skip the initial host approval prompt. |
| `--audit-log <path>` | Write JSONL audit events to a custom path. |
| `--audit-log off` | Disable audit logging. |
| `--audit-input` | Include typed input text in audit records. |
| `--status-interval 30s` | Print byte counters and last activity while hosting. |
| `--room <name>` | Attach a room name to the invite. |

By default audit logs are written under `~/.controli/audit/`.

Tunnel mode requires a named Cloudflare Tunnel route pointing to the local Controli service, usually `http://localhost:8765`.
