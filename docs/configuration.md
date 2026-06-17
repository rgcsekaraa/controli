# Configuration

Configuration lives in the Controli state file under the user home directory.

The host needs a relay URL and at least one workspace.

Keep workspace paths narrow. A shared shell has the same file access as the host user who started it.

Common host options:

```bash
controli host share --workspace main --room pair-session --mode full --status-interval 30s
```

Use `--audit-log off` only for local trusted testing. The default audit location is `~/.controli/audit/`.
