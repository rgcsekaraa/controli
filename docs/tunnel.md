# Tunnel Mode

Tunnel mode is the recommended transport for long sessions.

In tunnel mode:

- The host runs a local Controli terminal server.
- `cloudflared` exposes that local server through a named Cloudflare Tunnel.
- The guest enters the same 7-digit Controli code.
- Terminal traffic goes through Cloudflare Tunnel, not the Durable Object relay.
- The 7-digit invite lookup uses Workers KV.

This avoids Durable Objects duration usage for long terminal sessions.

## Cloudflare Setup

Create a named Cloudflare Tunnel in the Cloudflare dashboard or with `cloudflared`.

Cloudflare references:

- [Cloudflare Tunnel setup](https://developers.cloudflare.com/tunnel/setup/)
- [Cloudflare Tunnel configuration file](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/do-more-with-tunnels/local-management/configuration-file/)
- [Cloudflare WebSockets](https://developers.cloudflare.com/network/websockets/)

Add a public hostname route:

| Field | Value |
| --- | --- |
| Hostname | `cli.example.com` |
| Service URL | `http://localhost:8765` |

Run the connector on your Mac:

```bash
cloudflared tunnel run <tunnel-name>
```

Keep `cloudflared` running while the session is active.

## Start a Host Session

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --minutes 1440 --mode full
```

Send the printed 7-digit code to the guest.

The host must keep both processes running:

- `cloudflared tunnel run <tunnel-name>`
- `controli host tunnel ...`

## Guest Join

```bash
controli join 1234567
```

The guest browser opens the tunnel URL with a session token.

## Local Port

The default local service is:

```text
127.0.0.1:8765
```

Use a different local address if your named tunnel route points somewhere else:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com --listen 127.0.0.1:9000
```

The Cloudflare Tunnel route must match the same local port.

## When To Use Relay Mode

Use relay mode only for short sessions or fallback testing:

```bash
controli host share --workspace main
```

Relay mode keeps a Durable Object active for terminal traffic. Long relay sessions can exhaust the Durable Objects free tier.
