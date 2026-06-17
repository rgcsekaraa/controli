# Controli

Controli is a cross-platform CLI sharing tool. The host starts a shell session, gets a 7-digit code, and the guest joins from another machine through an outbound Cloudflare WebSocket relay.

The current implementation is a native Go binary.

## Supported Platforms

- Host: macOS and Linux/Ubuntu using a real PTY.
- Guest: Windows, macOS, and Linux/Ubuntu.
- Windows hosting: planned via ConPTY; Windows joining works now.
- Relay: Cloudflare Worker with Durable Objects.

## Install

Download the binary that matches the guest or host machine. Current release: [v0.2.0](https://github.com/rgcsekaraa/controli/releases/tag/v0.2.0).

Release page:

```text
https://github.com/rgcsekaraa/controli/releases
```

Documentation site:

```text
https://rgcsekaraa.github.io/controli/
```

### Windows

| Machine | Download |
| --- | --- |
| Most Intel or AMD PCs | [controli-windows-amd64.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-windows-amd64.exe) |
| Older 32-bit PCs | [controli-windows-386.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-windows-386.exe) |
| Windows on ARM64 | [controli-windows-arm64.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-windows-arm64.exe) |
| Older Windows ARM devices | [controli-windows-arm.exe](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-windows-arm.exe) |

Quick download:

```powershell
Invoke-WebRequest -Uri "https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-windows-amd64.exe" -OutFile "$env:USERPROFILE\Downloads\controli.exe"
& "$env:USERPROFILE\Downloads\controli.exe" join 1234567
```

### macOS

| Machine | Download |
| --- | --- |
| Apple Silicon | [controli-darwin-arm64](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-darwin-arm64) |
| Intel Mac | [controli-darwin-amd64](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-darwin-amd64) |

Quick download for Apple Silicon:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-darwin-arm64
chmod +x controli
./controli join 1234567
```

### Linux and Ubuntu

| Machine | Download |
| --- | --- |
| Most Intel or AMD desktops and servers | [controli-linux-amd64](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-amd64) |
| Older 32-bit Intel or AMD systems | [controli-linux-386](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-386) |
| ARM64 servers and boards | [controli-linux-arm64](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-arm64) |
| ARMv7 boards | [controli-linux-armv7](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-armv7) |
| ARMv6 boards | [controli-linux-armv6](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-armv6) |
| PowerPC 64 little-endian servers | [controli-linux-ppc64le](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-ppc64le) |
| RISC-V 64 systems | [controli-linux-riscv64](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-riscv64) |
| IBM Z or LinuxONE | [controli-linux-s390x](https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-s390x) |

Quick download for most PCs and servers:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-amd64
chmod +x controli
./controli join 1234567
```

After downloading on Unix systems, mark the file executable with `chmod +x`.

## Relay Setup

Deploy the bundled Cloudflare Worker once:

```bash
cd infra/cloudflare-relay
npm install
npx wrangler login
npx wrangler deploy
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

Start sharing:

```bash
controli host share --workspace main --minutes 480
```

Send the printed 7-digit code to the guest.

## Guest Join

```bash
controli join 1234567
```

By default, joining opens a local browser terminal powered by embedded xterm.js. This avoids Windows console rendering freezes and gives the same terminal renderer on Windows, macOS, and Linux.

Direct console rendering is available for debugging:

```bash
controli join 1234567 --console
```

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
- The relay sees encrypted WebSocket transport metadata but does not need host inbound ports.
- The guest controls the hosted shell for the lifetime of the session.
- This is alpha software. Use it on machines you own or are authorized to administer.

## Documentation

- [Documentation site](https://rgcsekaraa.github.io/controli/)
