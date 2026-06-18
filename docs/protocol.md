# Protocol

Controli has two transports:

- Tunnel mode for long sessions.
- Relay fallback for short sessions and fallback testing.

## Tunnel Mode

The host serves the browser terminal from a local HTTP/WebSocket server. A named Cloudflare Tunnel exposes that local server through a public hostname. The guest receives a 7-digit code, claims a token from Workers KV, and opens the tunnel URL in the browser.

Terminal bytes in tunnel mode do not pass through Durable Objects.

One active guest WebSocket is allowed per host session. A guest can reconnect with the same 7-digit code after disconnect while the invite is still valid. Every new guest connection resets host approval before input is accepted.

## Relay Fallback

The relay protocol has two sides:

- `host`
- `client`

Terminal bytes are sent as WebSocket binary messages. Resize events and guest connection events are encoded as control messages with a reserved prefix.

Relay fallback allows one active client socket per session. Additional clients are rejected while a guest is connected. When a guest reconnects after disconnect, the relay notifies the host so approval is reset before any input reaches the shell.

Short invites store the full token in Workers KV for a limited time. The guest sends the 7-digit code to claim that token.

Invite metadata can include a room name and permission mode. The host enforces permissions locally.

Relay pending data is capped by message count and total bytes. If a peer stays disconnected while output continues, the oldest queued messages are dropped first to keep the relay stable.
