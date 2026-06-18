# Controli

Controli is a native Go CLI sharing tool for support sessions. A host starts a local shell, receives a 7-digit invite code, and a guest joins through a browser terminal.

## Current Scope

| Area | Status |
| --- | --- |
| Host support | macOS and Linux with a real PTY, Windows with stdio |
| Guest support | Browser join for tunnel sessions, CLI join for fallback |
| Long sessions | Named Cloudflare Tunnel |
| Short-code lookup | Cloudflare Worker with Workers KV |
| Relay fallback | Cloudflare Worker with Durable Objects |
| Guest terminal | Browser terminal using xterm.js |

## Quick Start

Configure the short-code Worker URL once:

```bash
controli relay configure --url wss://controli-relay.example.workers.dev
```

Start a long tunnel session and print an invite code:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 0
```

Join from the guest machine with no install:

```text
https://controli-relay.rgcsekaraa.workers.dev/join
```

Or join with the CLI:

```bash
controli join 1234567
```

## Documentation

- [Install](install.md): download releases and build from source.
- [Compatibility](compatibility.md): choose the right binary for each OS and CPU.
- [Tunnel Mode](tunnel.md): run long sessions without Durable Objects duration.
- [Host](host.md): configure a workspace and start sharing.
- [Join](join.md): connect from Windows, macOS, or Linux.
- [Relay](relay.md): deploy and operate the Cloudflare relay.
- [Protocol](protocol.md): understand session routing and terminal transport.
- [Security](security.md): review the permission model and operational limits.

## Operational Notes

- The invite code is temporary and should be treated like a password.
- The guest controls the hosted shell for the lifetime of the session.
- Tunnel mode does not require inbound ports on the host machine.
- Use a relay deployed in your own Cloudflare account for production use.
