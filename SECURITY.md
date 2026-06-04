# Security Policy

## Supported Versions

The latest minor release of `ionoscloud-mcp` receives security updates. Older
releases may receive fixes for critical vulnerabilities on a best-effort basis.

## Reporting a Vulnerability

**Please do not open public GitHub issues for security vulnerabilities.**

Report security issues privately via GitHub's
[private vulnerability reporting](https://github.com/ionos-cloud/ionoscloud-mcp/security/advisories/new),
or by email to `sdk-tooling@ionos.com`.

Include:
- A description of the issue and its potential impact
- Steps to reproduce
- Any proof-of-concept code or configuration

We aim to acknowledge reports within 3 business days and to provide an initial
assessment within 10 business days. Coordinated disclosure timelines will be
agreed with the reporter.

## Scope

In scope:
- The `ionoscloud-mcp` server binary and Docker image
- Tool implementations under `tools/`
- The release pipeline and published artifacts

Out of scope:
- Vulnerabilities in the IONOS CLOUD APIs themselves (report to IONOS via the
  IONOS CLOUD support channels)
- Vulnerabilities in dependencies — please report upstream and notify us
