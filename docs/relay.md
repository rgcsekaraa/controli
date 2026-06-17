# Relay

The relay is a Cloudflare Worker with Durable Objects.

It handles:

- WebSocket rendezvous for host and guest.
- Temporary short-code invite lookup.
- Pairing messages by session id.
- Byte-bounded pending queues during reconnects.
- Closing both sides when a session is finished.

The host does not open inbound ports. Both host and guest make outbound connections.

Deploy from a source checkout:

```bash
controli relay deploy
```

Check the configured relay:

```bash
controli relay status
```

The Worker also exposes:

```text
/health
```

Session queues are capped by message count and byte size so a disconnected peer cannot grow relay memory without limit.
