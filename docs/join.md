# Join

For tunnel sessions, the guest can join from a browser without installing Controli.

Open:

```text
https://controli-relay.rgcsekaraa.workers.dev/join
```

Enter the 7-digit code and join password from the host. The browser redirects to the hosted terminal.

The CLI join command is still available:

```bash
controli join 1234567
controli join 1234567 --password abcd-1234-wxyz
```

For short-code joins, the command prompts for the join password if `--password` is omitted. The command opens a local browser terminal and keeps a small local HTTP server alive for the duration of the session. Use this for relay fallback sessions or diagnostics.

Direct console rendering is available for diagnostics:

```bash
controli join 1234567 --console
```
