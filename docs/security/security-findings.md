# Security findings ledger

Revision: `6ae568f9389b851e15cd1ce849d93ebae22b578a`

| ID | Title | Domain | Status | Validation |
| --- | ----- | ------ | ------ | ---------- |
| shorturl-org-list | Authenticated user can list all org short URLs | Kubernetes app APIs / shorturl authorizer | validated | docs/security/validations/shorturl-org-list.md |

## shorturl-org-list

- **Title:** GetAuthorizer allows `list` on `shorturls` for any authenticated resource request
- **Domain:** Kubernetes short URL API (`shorturl.grafana.app/v1beta1`)
- **Evidence:** `apps/shorturl/pkg/app/authorizer.go:13-18` returns `DecisionAllow` for every resource request; `pkg/services/shorturls/shorturlimpl/store.go:92-95` lists by `org_id` only
- **Attack chain:** signed-in org user → `/apis/shorturl.grafana.app/v1beta1/namespaces/{org}/shorturls` LIST → API authorizer Allow → SQL `WHERE org_id = ?`
- **Attacker capability:** any authenticated org member (Viewer / `None` role included)
- **Deployment assumptions:** Grafana OSS with the short URL app installer registered (default)
- **Impact:** org-wide disclosure of short URL paths (dashboard/explore links created by other users)
- **Confidence:** high for the authorizer decision; storage filter confirmed in SQL
- **Unknowns:** no live HTTP LIST in this proof; cross-org blocked by namespace authorizer
- **Source specialist:** appsec / authorization review
- **Status:** validated
- **Validation artifact:** `docs/security/validations/shorturl-org-list.md`
