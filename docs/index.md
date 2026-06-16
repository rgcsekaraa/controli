# Controli

Controli is a native Go CLI sharing tool for support sessions. A host starts a local shell, receives a 7-digit invite code, and a guest joins through an outbound Cloudflare WebSocket relay.

## Current Scope

| Area | Status |
| --- | --- |
| Host support | macOS and Linux with a real PTY |
| Guest support | Windows, macOS, and Linux |
| Relay | Cloudflare Worker with Durable Objects |
| Guest terminal | Browser terminal using xterm.js |
| Windows hosting | Planned via ConPTY |

## Quick Start

Configure the relay URL once:

```bash
controli relay configure --url wss://controli-relay.example.workers.dev
```

Start a host session and print an invite code:

```bash
controli host share --workspace main --minutes 480
```

Join from the guest machine:

```bash
controli join 1234567
```

## Documentation

- [Install](install.md): download releases and build from source.
- [Compatibility](compatibility.md): choose the right binary for each OS and CPU.
- [Host](host.md): configure a workspace and start sharing.
- [Join](join.md): connect from Windows, macOS, or Linux.
- [Relay](relay.md): deploy and operate the Cloudflare relay.
- [Protocol](protocol.md): understand session routing and terminal transport.
- [Security](security.md): review the permission model and operational limits.

## Operational Notes

- The invite code is temporary and should be treated like a password.
- The guest controls the hosted shell for the lifetime of the session.
- The relay does not require inbound ports on the host machine.
- Use a relay deployed in your own Cloudflare account for production use.
