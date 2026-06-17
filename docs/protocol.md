# Protocol

The relay protocol has two sides:

- `host`
- `client`

Terminal bytes are sent as WebSocket binary messages. Resize events are encoded as control messages with a reserved prefix.

Short invites store the full relay token in the Worker for a limited time. The guest sends the 7-digit code to claim that token.

Invite metadata can include a room name and permission mode. The host enforces permissions locally.

Relay pending data is capped by message count and total bytes. If a peer stays disconnected while output continues, the oldest queued messages are dropped first to keep the relay stable.
