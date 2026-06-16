# Overview

Controli shares a terminal session through outbound WebSocket connections.

The host starts a shell in a PTY. The guest joins with a short code. Both sides connect to the relay, and the relay forwards terminal bytes between them.

The default guest renderer is a local browser terminal backed by embedded xterm.js assets. This keeps rendering consistent across Windows, macOS, and Linux.

