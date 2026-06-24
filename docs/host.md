# Host

Hosting is supported on macOS, Linux, and Windows.

The host opens a real PTY and starts the configured shell inside the selected workspace. PTY resize events come from the guest browser terminal.

macOS and Linux use a real PTY. When `tmux` is installed, Controli starts the hosted shell inside a persistent `tmux` session by default. If the Controli host process exits or the network path drops, start the same workspace again to reattach to the same shell.

Windows hosting uses a stdio backend. Windows guests can reconnect, but a Windows-hosted shell does not yet survive the host process exiting.

Start a long tunnel session:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 0
```

Use a custom persistent shell name:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 0 --persist-name main
```

Start a relay fallback session:

```bash
controli host share --workspace main --minutes 480
```

Relay fallback sessions use Durable Objects for live terminal traffic. Keep them short, or use tunnel mode for long sessions to avoid Cloudflare free-tier duration limits.

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
| `--password <value>` | Set the join password printed with the 7-digit invite code. |
| `--downloads` | Allow browser-terminal downloads from `<workspace>/controli-drive`. |
| `--download-code <value>` | Let guests authorize downloads with this secret code. Can also be set with `CONTROLI_DOWNLOAD_CODE`. |
| `--s4d-code <value>` | Alias for `--download-code`. Can also be set with `CONTROLI_S4D_CODE`. |
| `--download-approve=false` | Deprecated; use `--download-code` or host approval. |
| `--persist=false` | Disable the persistent `tmux` backend on macOS and Linux. |
| `--persist-name <name>` | Choose the stable `tmux` session name used for reattach. |

By default audit logs are written under `~/.controli/audit/`.

To share files, start the host with `--downloads` and place files inside the workspace's `controli-drive` folder. Guests can download only paths inside that folder. If `--download-code`, `--s4d-code`, `CONTROLI_DOWNLOAD_CODE`, or `CONTROLI_S4D_CODE` is set, guests can enter that code in the browser to authorize the download; blank or wrong codes fall back to host approval.

Tunnel mode requires a named Cloudflare Tunnel route pointing to the local Controli service, usually `http://localhost:8765`.

Persistent host sessions require the host machine to stay awake and powered on. A reboot, shutdown, user logout that kills `tmux`, or the shell process exiting ends the live shell.
