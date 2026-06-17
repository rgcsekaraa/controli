# Security

Treat every invite code like a password while it is valid.

The host is prompted before the guest can control the shell. Use `--mode view` when the guest should only watch, and use `--mode approve` when every input chunk should be approved by the host.

The guest controls the hosted shell after approval. Use a dedicated workspace and avoid sharing more access than needed.

Audit logs are enabled by default and record session lifecycle, resize events, byte counts, and permission decisions. Typed input is not recorded unless `--audit-input` is set.

Run a relay you control. Rotate invite codes by stopping and starting the host share.
