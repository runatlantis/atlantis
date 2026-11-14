# 6. Sign in to the web UI with SSO and roles

Date: 2026-11-13

## Status

Proposed

Related issue: to be filed before publication, as [ADR 0003](0003-govern-architecture-proposals.md) requires.

## Context

The Atlantis web UI is protected only by optional basic auth: one shared username and password, off by default, with no record of who did what. Anyone who reaches the UI can discard plans, unlock projects, and lock applies for every repository. Operators who want single sign-on put an authenticating proxy in front of Atlantis, but Atlantis cannot tell users apart behind it, so it cannot limit actions by person or name them in pull request comments.

The experimental web UI behind the `web-ui` experiment adds more actions and more visible data, which makes this gap larger. Users have asked for OAuth and SSO in [#596](https://github.com/runatlantis/atlantis/issues/596) and [#4452](https://github.com/runatlantis/atlantis/issues/4452).

Constraints:

- Atlantis may run as several replicas behind a load balancer. Today there is no shared session store, and [ADR 0005](0005-use-durable-storage-with-one-data-model.md) is still proposed.
- Organizations use different identity providers: Microsoft Entra ID, Google Workspace, Okta, Keycloak, Auth0, GitLab, GitHub, and AWS IAM Identity Center are the most common. IAM Identity Center supports only SAML for application sign-in.
- Atlantis is a single binary configured through flags, environment variables, and YAML. A required extra service would be a large change for small installs.

## Decision

Atlantis will sign people in to the web UI natively and give each person a role. This applies to the pages at `/ui`. The existing UI and the API keep their current authentication.

### Identity sources

Operators choose one of two modes:

- **Sign-in pages.** Atlantis redirects to an OpenID Connect provider, GitHub, or both. OIDC covers Entra ID, Google, Okta, Keycloak, Auth0, GitLab, Amazon Cognito, and Dex through one generic implementation using `coreos/go-oidc` and `golang.org/x/oauth2`. GitHub is not an OIDC provider, so it has its own OAuth implementation and requires an organization allowlist, because anyone can create a GitHub account. Organizations and teams become groups prefixed with `github:`, so a GitHub organization can never match another provider's group of the same name, and child-team membership counts for the parent team, as the existing team allowlist does.
- **An authenticating proxy.** Atlantis takes the identity from a proxy in front of it and verifies it on every request: the signed `x-amzn-oidc-data` header from an AWS Application Load Balancer, checked against the load balancer's ARN; the signed assertion from Google Cloud Identity-Aware Proxy, checked for audience and issuer; or plain headers, accepted only from configured networks. This mode covers IAM Identity Center through Cognito behind a load balancer.

Atlantis will not implement SAML. Organizations that need it can use a provider that bridges SAML to OIDC, such as Cognito, Keycloak, or Dex.

### Roles

There are three roles. Viewers can read pull requests, plans, locks, and logs. Operators can also discard plans and unlock projects. Admins can also lock and unlock applies. Every route checks the role on the server.

Roles come from rules that match groups, verified email addresses, email domains, or everyone who signed in. The highest matching role wins. People who match no rule are turned away unless operators set a default role. Atlantis refuses to start with sign-in enabled and no role rules.

### Sessions

Sessions are cookies encrypted with AES-256-GCM, with keys derived by HKDF. The cookie's purpose is bound as additional data, and every value carries an expiry. Sessions are short by default, four hours, because the role is fixed for the life of a session. Operators configure a list of secrets: the first encrypts and all decrypt, so secrets rotate without signing everyone out. Every replica uses the same secrets, so any replica can serve any request without sticky sessions or a shared store.

The sign-in state, nonce, and PKCE verifier travel in a separate short-lived encrypted cookie, so the provider's callback can reach any replica. The flow uses the authorization code grant with PKCE, a random state compared in constant time, and a nonce checked against the ID token.

### Security requirements

- Cookies are `HttpOnly`, `SameSite=Lax`, and `Secure` except on a loopback host over plain http.
- Every state-changing request, including sign-out, rejects cross-origin requests.
- Return paths after sign-in must stay under `/ui`.
- Signed-out visitors see only the sign-in page, which reveals nothing about the server.
- Entra ID's group overage, where groups are left out of the token, fails closed with an explanation.
- Unverified email addresses never match role rules, including emails from proxies that omit `email_verified`.
- Sign-in errors shown to users never include provider responses or internal hosts; details go to the log with a reference.
- Calls to identity providers and key endpoints have timeouts and size limits, and unknown key IDs cannot cause repeated outbound requests.
- Sign-ins, denials, and actions are logged with the user, provider, and role, and pull request comments name the user.

### Rollout

Sign-in ships behind the `web-ui` experiment and graduates with it under [ADR 0004](0004-govern-feature-lifecycles.md). Without sign-in settings, the UI keeps basic auth, or no auth, as today.

### Alternatives considered

- **Require an external proxy only.** Simplest for Atlantis, but every install needs another component, and plain header trust is unsafe unless the network is locked down. The proxy mode keeps this option for those who already run one.
- **Embed Dex.** Covers SAML, LDAP, and Google groups, but adds a large dependency set, its own storage, and another service surface to the binary.
- **Server-side sessions.** Allow instant revocation, but need a store shared by every replica. Revisit when durable storage is accepted.
- **Role expressions over claims**, such as JMESPath. More flexible, but harder to review than a fixed list of rules. Revisit if the rules prove too limited.

## Consequences

Operators can use their existing identity provider, limit what each person can do, and see who discarded a plan or locked applies. Small installs need no new services.

Atlantis takes on `coreos/go-oidc`, `go-jose`, and a direct dependency on `golang.org/x/oauth2`, and must keep its OIDC and OAuth handling current.

Ending one person's session before it expires is not possible without a shared store; operators can only rotate the secrets, which signs everyone out. Durable storage can add per-session revocation and a stored audit log later.

Google Workspace groups are not available natively, because Google does not include them in tokens. Entra ID users in more than 200 groups need app roles.

The existing UI and the API keep their current authentication. Extending sign-in to them needs its own decision.
