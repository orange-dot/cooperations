# Security Policy

## Supported Versions

Only the latest `main` branch is supported for security fixes.

## Reporting a Vulnerability

Please report security issues privately to the maintainers.

Preferred options:
1. Use a private security advisory channel (if available in your hosting platform).
2. Email the maintainers at a private address (add your org email here, e.g. `security@yourdomain`).

If no private channel is available, open a public issue with minimal details and request a private follow-up.

## Scope

This project is a local-first CLI orchestration tool. Security concerns include:
- secret handling and redaction
- prompt-injection or untrusted input handling
- local file access and logging behavior
