# Commands

This page lists every current Controli command and flag.

## Host a tunnel session

```bash
controli host tunnel --workspace main --public-url https://cli.example.com
```

Starts the configured workspace shell and serves the browser terminal locally for a named Cloudflare Tunnel. This is the recommended transport for long sessions.

| Flag | Required | Default | Purpose |
| --- | --- | --- | --- |
| `--workspace <name>` | Yes | | Workspace key from `~/.controli/state.json`. |
| `--public-url <url>` | Yes | | Public Cloudflare Tunnel hostname. |
| `--listen <addr>` | No | `127.0.0.1:8765` | Local HTTP service address for `cloudflared`. |
| `--room <name>` | No | Workspace name | Room label shown to the guest. |
| `--relay-url <url>` | No | Configured relay or default relay | Worker URL used only for 7-digit invite lookup. |
| `--name <name>` | No | `guest` | Guest label stored in the invite. |
| `--minutes <n>` | No | `1440` | Session lifetime in minutes. |
| `--shell <path>` | No | Workspace shell or default shell | Shell to start for this session. |
| `--print-only` | No | `false` | Print a code without starting the shell. |
| `--long-code` | No | `false` | Print the full self-contained code instead of a 7-digit code. |
| `--mode full` | No | `full` | Guest can type after host approval. |
| `--mode view` | No | | Guest can watch only. |
| `--mode approve` | No | | Host approves each input chunk. |
| `--approve=false` | No | `true` | Skip the first host approval prompt. |
| `--audit-log <path>` | No | `~/.controli/audit/<session>.jsonl` | Custom audit log path. |
| `--audit-log off` | No | | Disable audit logging. |
| `--audit-input` | No | `false` | Store typed input text in audit records. |
| `--status-interval 30s` | No | disabled | Print session counters while hosting. |

Examples:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 1440
controli host tunnel --workspace main --public-url https://cli.example.com --mode view
controli host tunnel --workspace main --public-url https://cli.example.com --listen 127.0.0.1:9000
controli host tunnel --workspace main --public-url https://cli.example.com --approve=false
```

## Host a relay fallback session

```bash
controli host share --workspace main
```

Starts the configured workspace shell, registers a 7-digit invite code, and sends terminal traffic through the Durable Object relay. Use this for short sessions or fallback testing.

| Flag | Required | Default | Purpose |
| --- | --- | --- | --- |
| `--workspace <name>` | Yes | | Workspace key from `~/.controli/state.json`. |
| `--room <name>` | No | Workspace name | Room label shown to the guest. |
| `--relay-url <url>` | No | Configured relay | Override the relay for this share. |
| `--name <name>` | No | `guest` | Guest label stored in the invite. |
| `--minutes <n>` | No | `120` | Session lifetime in minutes. |
| `--shell <path>` | No | Workspace shell or default shell | Shell to start for this session. |
| `--print-only` | No | `false` | Print a code without starting the shell. |
| `--long-code` | No | `false` | Print the full self-contained code instead of a 7-digit code. |
| `--mode full` | No | `full` | Guest can type after host approval. |
| `--mode view` | No | | Guest can watch only. |
| `--mode approve` | No | | Host approves each input chunk. |
| `--approve=false` | No | `true` | Skip the first host approval prompt. |
| `--audit-log <path>` | No | `~/.controli/audit/<session>.jsonl` | Custom audit log path. |
| `--audit-log off` | No | | Disable audit logging. |
| `--audit-input` | No | `false` | Store typed input text in audit records. |
| `--status-interval 30s` | No | disabled | Print session counters while hosting. |

Examples:

```bash
controli host share --workspace main --minutes 480 --mode full
controli host share --workspace main --mode view
controli host share --workspace main --mode approve
controli host share --workspace main --room support-a --status-interval 30s
controli host share --workspace main --long-code
controli host share --workspace main --print-only
controli host share --workspace main --audit-log off
```

## Join a session

```bash
controli join 1234567
```

Resolves the code and opens the local browser terminal by default on Windows, macOS, and Linux.

| Flag | Default | Purpose |
| --- | --- | --- |
| `--relay-url <url>` | Default relay | Relay to use when resolving a 7-digit code. |
| `--web-terminal` | automatic | Force the local browser terminal. |
| `--console` | `false` | Render directly in the current console for debugging. |

Examples:

```bash
controli join 1234567
controli join 1234567 --console
controli join 1234567 --web-terminal
controli join 1234567 --relay-url wss://controli-relay.example.workers.dev
```

If no code is passed, Controli prompts for one:

```bash
controli join
```

## Configure the relay

```bash
controli relay configure --url wss://controli-relay.example.workers.dev
```

Stores the relay URL in `~/.controli/state.json`.

## Check relay status

```bash
controli relay status
```

Prints the configured relay URL and checks the relay `/health` endpoint.

## Deploy the bundled relay

```bash
controli relay deploy
```

Runs the Cloudflare Worker deploy command from a Controli source checkout.

## Update Controli

```bash
controli update
```

Downloads the latest release asset for the current OS and CPU.

| Flag | Default | Purpose |
| --- | --- | --- |
| `--repo <owner/name>` | `rgcsekaraa/controli` | Download from a different GitHub repository. |

Example:

```bash
controli update --repo rgcsekaraa/controli
```

On Windows, the updater downloads a `.new.exe` file because Windows locks the running executable.

## Common workflows

Mac host, Windows guest with tunnel mode:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 1440 --mode full
```

The Windows guest runs:

```powershell
& "$env:LOCALAPPDATA\Controli\controli.exe" join 1234567
```

View-only support session:

```bash
controli host share --workspace main --mode view
```

Sensitive session with input approval:

```bash
controli host share --workspace main --mode approve --audit-input
```

Known relay issue with short code lookup:

```bash
controli host share --workspace main --long-code
```
