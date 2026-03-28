# Security Policy

## Supported Versions

| Version | Status | Supported Until |
|---------|--------|-----------------|
| 1.0.x   | Current | TBD |

## Reporting a Vulnerability

If you discover a security vulnerability in Centipede, please **do not** open a public GitHub issue.

Instead, please email **security@centipede.dev** with:

1. **Description** — What is the vulnerability?
2. **Location** — Where in the code is it?
3. **Impact** — What could be exploited?
4. **Reproduction** — How can we verify it?

We will respond within 48 hours and work with you on a patch.

## Security Considerations

### Production Deployment

- Always run with proper authentication (Azure managed identity or service principal)
- Use TLS/HTTPS for all API calls
- Rotate credentials regularly
- Enable audit logging for all tenant blocking actions
- Monitor exit codes for anomalies in CI/CD pipelines
- Use network policies to restrict APIM API access

### Credential Management

- Store `AZURE_CLIENT_SECRET` in secure secret managers (GitHub Secrets, Azure KeyVault, etc.)
- Never commit credentials to version control
- Use managed identities in Kubernetes when possible
- Implement credential rotation policies

### Data Protection

- Baseline files contain traffic patterns — protect with appropriate permissions
- Detection logs may contain sensitive tenant information
- Implement log retention policies
- Use encryption at rest for storage

## Known Limitations

1. **Baseline Dependency** — Detection accuracy depends on clean baseline data
2. **Subtle Attacks** — Attackers using stolen credentials at normal rates won't be detected
3. **Zero-Day Exploits** — Logic bugs are not detected by behavioral analysis
4. **Silent Exfiltration** — Slow data theft at normal request rates won't trigger alerts

See [Production Readiness Assessment](PRODUCTION_READINESS_ASSESSMENT.md) for complete threat model.
