# Protocol

The relay protocol has two sides:

- `host`
- `client`

Terminal bytes are sent as WebSocket binary messages. Resize events are encoded as control messages with a reserved prefix.

Short invites store the full relay token in the Worker for a limited time. The guest sends the 7-digit code to claim that token.

