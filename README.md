# Controli

Controli is a cross-platform CLI sharing tool. The host starts a shell session, gets a 7-digit code, and the guest joins from another machine through a browser terminal.

The current implementation is a native Go binary.

## Supported Platforms

- Host: macOS and Linux/Ubuntu using a real PTY. Windows hosting uses a stdio backend.
- Guest: Windows, macOS, and Linux/Ubuntu.
- Long sessions: Cloudflare Tunnel transport.
- Short-code lookup: Cloudflare Worker with Workers KV.
- Relay fallback: Cloudflare Worker with Durable Objects.

## Install

Download the binary that matches the guest or host machine. Current release: [v0.4.5](https://github.com/rgcsekaraa/controli/releases/tag/v0.4.5).

Release page:

```text
https://github.com/rgcsekaraa/controli/releases
```

Documentation site:

```text
https://rgcsekaraa.github.io/controli/
```

Command reference:

```text
https://rgcsekaraa.github.io/controli/commands
```

### Windows

| Machine | Download |
| --- | --- |
| Most Intel or AMD PCs | [controli-windows-amd64.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-windows-amd64.exe) |
| Older 32-bit PCs | [controli-windows-386.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-windows-386.exe) |
| Windows on ARM64 | [controli-windows-arm64.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-windows-arm64.exe) |
| Older Windows ARM devices | [controli-windows-arm.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-windows-arm.exe) |

Quick download:

```powershell
Invoke-WebRequest -Uri "https://github.com/rgcsekaraa/controli/releases/latest/download/controli-windows-amd64.exe" -OutFile "$env:USERPROFILE\Downloads\controli.exe"
& "$env:USERPROFILE\Downloads\controli.exe" join 1234567
```

Windows release binaries are currently unsigned. If Device Guard or App Control blocks the file on a managed Windows machine, the organization's policy must allow the release hash or Controli must ship a trusted signed binary.

### macOS

| Machine | Download |
| --- | --- |
| Apple Silicon | [controli-darwin-arm64](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-darwin-arm64) |
| Intel Mac | [controli-darwin-amd64](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-darwin-amd64) |

Quick download for Apple Silicon:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/latest/download/controli-darwin-arm64
chmod +x controli
./controli join 1234567
```

### Linux and Ubuntu

| Machine | Download |
| --- | --- |
| Most Intel or AMD desktops and servers | [controli-linux-amd64](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-amd64) |
| Older 32-bit Intel or AMD systems | [controli-linux-386](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-386) |
| ARM64 servers and boards | [controli-linux-arm64](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-arm64) |
| ARMv7 boards | [controli-linux-armv7](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-armv7) |
| ARMv6 boards | [controli-linux-armv6](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-armv6) |
| PowerPC 64 little-endian servers | [controli-linux-ppc64le](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-ppc64le) |
| RISC-V 64 systems | [controli-linux-riscv64](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-riscv64) |
| IBM Z or LinuxONE | [controli-linux-s390x](https://github.com/rgcsekaraa/controli/releases/download/v0.4.5/controli-linux-s390x) |

Quick download for most PCs and servers:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/latest/download/controli-linux-amd64
chmod +x controli
./controli join 1234567
```

After downloading on Unix systems, mark the file executable with `chmod +x`.

One-command install:

```bash
curl -fsSL https://raw.githubusercontent.com/rgcsekaraa/controli/main/scripts/install.sh | sh
```

Windows one-command install:

```powershell
iwr https://raw.githubusercontent.com/rgcsekaraa/controli/main/scripts/install.ps1 -UseB | iex
```

Update later:

```bash
controli update
```

## Tunnel Setup

Tunnel mode is recommended for long sessions because terminal traffic does not use Durable Objects.

Create a named Cloudflare Tunnel and add a public hostname route:

| Cloudflare field | Value |
| --- | --- |
| Hostname | `cli.example.com` |
| Service URL | `http://localhost:8765` |

Run the tunnel connector:

```bash
cloudflared tunnel run <tunnel-name>
```

Then start Controli:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 0 --mode full
```

Send the printed 7-digit code to the guest.

Only one guest can be connected to a live session at a time. The same 7-digit code can be used again while the invite has not expired. A reconnect from the same guest keeps the existing approval; a different guest requires fresh host approval before input reaches the shell.

On macOS and Linux, install `tmux` for persistent hosted shells. Controli uses it by default when available, so the shell keeps running if the Controli host process detaches. Start the same workspace again to reattach.

## Relay Setup

Relay mode is available as a fallback for short sessions. Long relay sessions can exhaust Durable Objects free-tier duration.

Deploy the bundled Cloudflare Worker once:

```bash
cd infra/cloudflare-relay
npm install
npx wrangler login
controli relay deploy
```

Configure the relay URL on the host:

```bash
controli relay configure --url wss://controli-relay.<your-subdomain>.workers.dev
controli relay status
```

## Host Setup

Controli reads workspaces from `~/.controli/state.json`. Minimal example:

```json
{
  "relay": {
    "url": "wss://controli-relay.<your-subdomain>.workers.dev"
  },
  "workspaces": {
    "main": {
      "path": "/Users/you/work",
      "shell": "/bin/zsh"
    }
  }
}
```

Start a long tunnel session:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 0 --mode full
```

Start a relay fallback session:

```bash
controli host share --workspace main --minutes 480 --mode full
```

Send the printed 7-digit code to the guest.

Permission modes:

- `full`: guest can type after host approval.
- `view`: guest can watch but input is blocked.
- `approve`: host approves each input chunk before it reaches the shell.

Security behavior:

- One active guest is allowed per invite code.
- A reconnect with the same valid code is allowed after disconnect.
- Same-guest reconnects keep approval so temporary WebSocket drops do not interrupt control.
- A different guest requires fresh host approval before input reaches the shell.
- Additional guests are rejected while a guest is already connected.

Audit logs are written to `~/.controli/audit/<session>.jsonl` by default. Add `--audit-input` only when recording typed input is acceptable.

## Guest Join

```bash
controli join 1234567
```

By default, joining opens a local browser terminal powered by embedded xterm.js. This avoids Windows console rendering freezes and gives the same terminal renderer on Windows, macOS, and Linux.

Direct console rendering is available for debugging:

```bash
controli join 1234567 --console
```

## Command Reference

Common host commands:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 0
controli host tunnel --workspace main --public-url https://cli.example.com --persist-name main
controli host share --workspace main --mode full
controli host share --workspace main --mode view
controli host share --workspace main --mode approve
controli host share --workspace main --room support-a --status-interval 30s
controli host share --workspace main --long-code
controli host share --workspace main --print-only
```

Common guest commands:

```bash
controli join 1234567
controli join 1234567 --console
controli join 1234567 --web-terminal
```

Relay and update commands:

```bash
controli relay configure --url wss://controli-relay.<your-subdomain>.workers.dev
controli relay status
controli relay deploy
controli update
```

Full command and flag documentation is in the [command reference](https://rgcsekaraa.github.io/controli/commands).

## Build From Source

```bash
go test ./...
make build
```

This builds:

```text
dist/controli-darwin-arm64
dist/controli-darwin-amd64
dist/controli-linux-386
dist/controli-linux-amd64
dist/controli-linux-armv6
dist/controli-linux-armv7
dist/controli-linux-arm64
dist/controli-linux-ppc64le
dist/controli-linux-riscv64
dist/controli-linux-s390x
dist/controli-windows-386.exe
dist/controli-windows-amd64.exe
dist/controli-windows-arm.exe
dist/controli-windows-arm64.exe
```

## Security Notes

- Treat the 7-digit code like a password while it is valid.
- Use a Cloudflare Worker account you control.
- Tunnel mode avoids Durable Objects duration for terminal traffic.
- Relay fallback uses Durable Objects for terminal traffic and should be used for short sessions.
- The host is prompted before guest control starts unless `--approve=false` is used.
- Use `--mode view` when the guest should only watch.
- This is alpha software. Use it on machines you own or are authorized to administer.

## Documentation

- [Documentation site](https://rgcsekaraa.github.io/controli/)
