# Relay

The relay is a Cloudflare Worker. It now uses Workers KV for 7-digit invite lookup.

For tunnel mode it handles:

- Temporary short-code invite lookup.

For relay fallback it also handles:

- WebSocket rendezvous for host and guest through Durable Objects.
- Pairing messages by session id.
- Byte-bounded pending queues during reconnects.
- Closing both sides when a session is finished.

Tunnel mode is recommended for long sessions because terminal traffic does not keep a Durable Object active.

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

Session queues are capped by message count and byte size so a disconnected relay fallback peer cannot grow relay memory without limit.
