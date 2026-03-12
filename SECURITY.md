# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| main    | ✅ Yes             |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**⚠️ Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report security issues by emailing the maintainers or by using [GitHub's private vulnerability reporting](https://github.com/zeldebro/k8s-resource-rebalancer-operator/security/advisories/new).

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 1 week
- **Fix & Disclosure**: Coordinated with reporter

## Security Best Practices for Users

When deploying this operator:

1. **RBAC**: The operator uses minimal RBAC permissions. Review `config/rbac/role.yaml` before deploying
2. **Namespace Scoping**: Configure `userNamespace` to limit the operator's monitoring scope
3. **Network Policies**: Apply the provided network policies from `config/network-policy/`
4. **Image Security**: Use specific image tags rather than `latest` in production
5. **Metrics Endpoint**: Secure the metrics endpoint using the provided TLS configuration

## Dependencies

We regularly update dependencies to patch known vulnerabilities. If you notice an outdated dependency with a known CVE, please open an issue.

