# Security Policy

## Supported versions

Controli is pre-1.0 software. Security fixes are made on the default branch until a stable release policy is published.

## Reporting a vulnerability

Please report security issues privately before opening a public issue.

Email: security@rgcsekaraa.dev

If that address is not available, open a GitHub issue with only a minimal description and ask for a private contact path. Do not include exploit details or sensitive logs in a public issue.

## Security model

Controli shares a real shell over an outbound relay. It is not a sandbox.

Use it with:

- A machine you own or are authorized to administer.
- A dedicated workspace directory when possible.
- A relay you control.
- Short-lived invite codes.
- A clear trust boundary: the guest can execute commands in the hosted shell.

Do not use it to expose sensitive personal accounts or machines without isolation.

## Logging guarantees

The Go relay core does not currently provide tamper-evident session logging. Add external logging or run inside your own audited environment if you need forensic evidence.
