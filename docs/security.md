# Security

Treat every invite code like a password while it is valid.

The host is prompted before the guest can control the shell. Use `--mode view` when the guest should only watch, and use `--mode approve` when every input chunk should be approved by the host.

The guest controls the hosted shell after approval. Use a dedicated workspace and avoid sharing more access than needed.

Only one active guest is allowed per invite code. The same guest can reconnect after a transient disconnect without another approval prompt. A different guest requires fresh host approval before input is accepted. Additional guests are rejected while a guest is already connected.

Audit logs are enabled by default and record session lifecycle, resize events, byte counts, and permission decisions. Typed input is not recorded unless `--audit-input` is set.

Run a Worker and Cloudflare Tunnel you control. Rotate invite codes by stopping and starting the host session.

Tunnel mode keeps terminal traffic out of Durable Objects. Relay fallback uses Durable Objects for terminal traffic and should be reserved for short sessions.
