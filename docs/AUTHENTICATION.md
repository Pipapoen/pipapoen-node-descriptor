# Authentication Scope and Maintenance

Review date: 2026-11-21. This records the current implementation and deployment tradeoffs, rather than a security audit.

## Current implementation

PND implements an optional OIDC relying party using `github.com/coreos/go-oidc/v3` and `golang.org/x/oauth2`. It performs discovery, Authorization Code with PKCE, state and nonce verification, and ID-token verification through the library. The application issues signed HttpOnly, SameSite=Lax session cookies; Secure follows the configured HTTPS public origin. Enabled but incomplete or unavailable OIDC configuration fails startup. Omitted configuration enables public access.

The application maintains no local user database, roles, password login, registration, or refresh-token lifecycle. Application APIs require a valid session when OIDC is configured. A simulated provider tests the complete flow and negative state/nonce cases. A live Authentik client has not yet been configured on the trial deployment.

The application owns its cookie/session glue, so library use does not remove responsibility for that code. Sessions use a configurable absolute lifetime, defaulting to 30 minutes through `PND_SESSION_TTL`. API activity does not renew them. Logout clears the local browser cookie. Immediate centralized revocation and IdP back-channel logout are not implemented; an already issued session can remain valid until expiry. Public deployments should use HTTPS and an accurately configured public origin.

## Observed dependency check

`govulncheck` v1.10.0 reported `No vulnerabilities found` for the initial MVP source tree on Windows and for the Linux amd64 target on 2026-11-21. The initial MVP scan used go-oidc v3.21.0, oauth2 v0.38.0, and go-jose/v4 v4.1.4. The tool was run in project-local tooling/cache directories. No dependency versions were changed as a consequence of the clean scan.

This result concerns the known vulnerability database and the code paths analyzed at that time. It does not establish absence of unknown vulnerabilities, correct production proxy settings, or a completed application security review. A vulnerability in another OIDC product needs a matching component, version, and reachable behavior before it can be treated as a PND finding.

## Candidate 26.1.0-rc.3

The candidate updates go-jose/v4 to v4.1.9 and builds with Go 1.27.4. Local Go tests, simulated OIDC login and rejection cases, and vulnerability checks passed. The updated JWK parser accepts RSA moduli up to 4096 bits by default; configure the identity provider with a signing key within that limit.

## Maintenance options

| Approach | PND maintenance | Operator responsibility |
| --- | --- | --- |
| Built-in OIDC | Protocol client dependencies, callback/session code, negative integration tests, security updates | Provider configuration, HTTPS, client secrets, provider updates |
| External authenticating proxy | A smaller application authentication surface if native OIDC is eventually removed | Proxy/outpost updates, policies, HTTPS, protection of every application/API route, prevention of backend bypass |

Authentik forward auth uses an existing reverse proxy for application traffic and an outpost for authentication/authorization checks. For PND's sign-in-only requirement, this is a reasonable way to centralize access control where that infrastructure already exists. It transfers responsibility rather than eliminating it. The PND container must be reachable only through the protected path in such a deployment; the current directly exposed public LAN configuration is a trial mode, not proof of proxy enforcement. Client-supplied identity headers should not independently grant access.

Keeping native OIDC retains convenient standalone deployment for other operators. The recommended bounded scope is the current login gate, with dependency scanning and callback/session regression tests in the later CI stage. Expanding into accounts, roles, token refresh, or centralized logout would add scope and maintenance.

The subsequent session update retains the stateless signed-cookie architecture and introduces configurable short absolute lifetimes. The trial service remains intentionally public until its operator configures sign-in.

## References

- [Go vulnerability management](https://go.dev/doc/security/vuln/)
- [OAuth 2.0 Security Best Current Practice, RFC 9700](https://www.rfc-editor.org/rfc/rfc9700.html)
- [Authentik forward auth](https://docs.goauthentik.io/add-secure-apps/providers/proxy/forward_auth/)
