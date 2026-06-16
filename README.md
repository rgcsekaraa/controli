# Controli

Controli is a cross-platform CLI sharing tool. The host starts a shell session, gets a 7-digit code, and the guest joins from another machine through an outbound Cloudflare WebSocket relay.

The current implementation is a native Go binary.

## Supported Platforms

- Host: macOS and Linux/Ubuntu using a real PTY.
- Guest: Windows, macOS, and Linux/Ubuntu.
- Windows hosting: planned via ConPTY; Windows joining works now.
- Relay: Cloudflare Worker with Durable Objects.

## Install

Download a binary from the latest release:

```text
controli-darwin-arm64
controli-darwin-amd64
controli-linux-amd64
controli-linux-arm64
controli-windows-amd64.exe
```

Release page:

```text
https://github.com/rgcsekaraa/controli/releases
```

Documentation site:

```text
https://rgcsekaraa.github.io/controli/
```

Windows example:

```powershell
Invoke-WebRequest -Uri "https://github.com/rgcsekaraa/controli/releases/download/v0.2.1-go-alpha/controli-windows-amd64.exe" -OutFile "$env:USERPROFILE\Downloads\controli.exe"
& "$env:USERPROFILE\Downloads\controli.exe" join 1234567
```

macOS arm64 example:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/download/v0.2.1-go-alpha/controli-darwin-arm64
chmod +x controli
./controli join 1234567
```

Ubuntu/Linux example:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/download/v0.2.1-go-alpha/controli-linux-amd64
chmod +x controli
./controli join 1234567
```

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
dist/controli-linux-amd64
dist/controli-linux-arm64
dist/controli-windows-amd64.exe
```

## Security Notes

- Treat the 7-digit code like a password while it is valid.
- Use a Cloudflare Worker account you control.
- The relay sees encrypted WebSocket transport metadata but does not need host inbound ports.
- The guest controls the hosted shell for the lifetime of the session.
- This is alpha software. Use it on machines you own or are authorized to administer.

## More Docs

- [Documentation site](https://rgcsekaraa.github.io/controli/)
- [Overview](docs/overview.md)
- [Install](docs/install.md)
- [Relay](docs/relay.md)
- [Protocol](docs/protocol.md)
- [Configuration](docs/configuration.md)
- [Host](docs/host.md)
- [Join](docs/join.md)
- [Windows](docs/windows.md)
- [Linux](docs/linux.md)
- [macOS](docs/macos.md)
- [Security](docs/security.md)
- [Build](docs/build.md)
- [Release](docs/release.md)
- [Troubleshooting](docs/troubleshooting.md)
- [State](docs/state.md)
- [Roadmap](docs/roadmap.md)
