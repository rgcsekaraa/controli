# Troubleshooting

If a guest cannot join:

- Confirm the invite code has not expired.
- Confirm the Worker URL is configured on the host.
- For tunnel mode, confirm `cloudflared tunnel run <tunnel-name>` is running.
- For tunnel mode, confirm the public hostname route points to the same local address passed to `--listen`.
- Confirm both machines can make outbound HTTPS and WebSocket connections.
- Restart the host session to create a new invite.

If rendering is slow, use the default browser terminal instead of `--console`.

If the browser terminal does not scroll, upgrade both sides to the latest release and reopen the terminal page. Current builds forward mouse-wheel, trackpad, and touch gestures into xterm scrollback even when the download/status bar is visible.

If Cloudflare reports Durable Objects duration limits, stop relay fallback sessions and use tunnel mode:

```bash
controli host tunnel --workspace main --public-url https://cli.example.com
```
