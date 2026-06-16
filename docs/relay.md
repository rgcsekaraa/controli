# Relay

The relay is a Cloudflare Worker with Durable Objects.

It handles:

- WebSocket rendezvous for host and guest.
- Temporary short-code invite lookup.
- Pairing messages by session id.
- Closing both sides when a session is finished.

The host does not open inbound ports. Both host and guest make outbound connections.

