# Validation: shorturl-org-list

## Verdict

`validated`

Independent rerun failed for the claimed reason: `GetAuthorizer` for `verb=list` on `shorturls` returns `authorizer.DecisionAllow` for an authenticated resource request.

## Candidate

Authenticated Viewer can LIST all org short URLs via the Kubernetes API because `GetAuthorizer` returns `DecisionAllow` for any resource request, and SQL `List` is `WHERE org_id = ?` only.

## Attack chain (reconstructed from code, not the author's summary)

1. `ShortURLAppInstaller.GetAuthorizer` returns `apps/shorturl/pkg/app.GetAuthorizer` (`pkg/registry/apps/shorturl/register.go:59-60`).
2. That authorizer is registered for `shorturl.grafana.app/v1beta1` via `appinstaller.RegisterAuthorizers` (`pkg/services/apiserver/appinstaller/installer.go:123-136`) into `GrafanaAuthorizer` (`pkg/services/apiserver/service.go:293`).
3. Union order (`pkg/services/apiserver/auth/authorizer/authorizer.go:44-61`): impersonation, `system:masters`, namespace authorizer, then the API authorizer. Namespace check returns `DecisionNoOpinion` when `ns.OrgID == ident.GetOrgID()` (`pkg/services/apiserver/auth/authorizer/namespace.go:48-56`).
4. `GetAuthorizer` (`apps/shorturl/pkg/app/authorizer.go:13-18`): if `attr.IsResourceRequest()`, return `DecisionAllow` with empty reason. No verb, user, or role check.
5. Union stops at `DecisionAllow`; `NewRoleAuthorizer` is not consulted.
6. `legacyStorage.List` (`pkg/registry/apps/shorturl/legacy_storage.go:60-79`) calls `service.List(ctx, orgID)`.
7. SQL (`pkg/services/shorturls/shorturlimpl/store.go:92-95`): `dbSession.Where("org_id = ?", orgID).Find(&shortURLs)` — no `created_by` filter.

Required edges 1–5 are the claimed authorizer path. Edge 4 is what the reproduction executes. Edges 6–7 make org-wide enumeration the storage result of an allowed list.

## Command (unchanged)

```
export PATH="/home/ubuntu/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.linux-amd64/bin:$PATH"
export GOTOOLCHAIN=local
cd /tmp/sec-shorturl
go run .
```

## Assertion

Secure expectation: `list` must not be `authorizer.DecisionAllow` for an arbitrary authenticated user (`viewer@example.com`).

- Expected (secure): exit `0`, stdout `PASS: list is not unconditionally allowed`
- Received (current revision `6ae568f9389b851e15cd1ce849d93ebae22b578a`): exit `1`, `list decision=Allow reason=""`

## Verbatim output

### stdout

```
list decision=Allow reason=""
```

### stderr

```
FAIL: shorturl list is DecisionAllow for any authenticated resource request; org-wide short URL enumeration is permitted
exit status 1
```

(`exit status 1` is printed by `go run` when the program calls `os.Exit(1)`.)

### exit code

`1`

Failed for the claimed reason: `d == authorizer.DecisionAllow` (`authorizer.DecisionAllow` stringifies as `Allow`). Not a setup or import error.

## Failed for claimed reason?

Yes. The program printed `list decision=Allow reason=""` and exited `1` with the FAIL line that names DecisionAllow on list. The harness imported and ran against `/workspace/apps/shorturl` via the replace directive.

## Bounds / unknowns

- Reproduction is a local `GetAuthorizer().Authorize(...)` call, not an HTTP LIST against a running Grafana process.
- Namespace authorizer still denies cross-org namespaces; impact is org-scoped, not tenant-cross.
- Legacy HTTP API (`POST /api/short-urls`, `GET /api/short-urls/:uid`) has no list route; enumeration is the Kubernetes `list` verb.
- Authorizer does not inspect authentication; any resource request that reaches it is allowed. The test supplies an authenticated-looking user, matching the claim.

## Evidence paths

- `docs/security/validations/shorturl-org-list.md` (this record)
- `docs/security/validations/shorturl-org-list.log`
- `docs/security/validations/sec-shorturl-stdout.txt`
- `docs/security/validations/sec-shorturl-stderr.txt`
- `docs/security/validations/sec-shorturl-exitcode.txt`
- Harness (unchanged): `/tmp/sec-shorturl/main.go`
