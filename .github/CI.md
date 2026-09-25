# CI / CD notes (median)

Aligned with `docs/charter/git-rule.md` and peer `-an` libraries (shardian / crudian).

## Workflows

| Workflow | Role |
| --- | --- |
| `ci.yml` | PR/push tests for `packages/{go,js,php}` with path filters (docs-only does not start CI) |
| `codeql.yml` | CodeQL on main / develop / `dev-v*` |
| `scorecard.yml` | OpenSSF Scorecard on default branch |
| `release-on-tag.yml` | GitHub Release for `v*`, `packages/go/v*`, `packages/php/v*` |
| `publish-npm.yml` | Publish `@b4moss/median` from `release` when JS changed |
| `publish-go.yml` | GitHub Release + proxy ping for `github.com/b4moss/median/go` |

## Runtimes (CI)

- Go `1.26.x`
- Node `24`
- PHP `8.2`

## Tags

- JS: `vX.Y.Z` (matches `packages/js/package.json`)
- Go: `packages/go/vX.Y.Z` (matches `packages/go/VERSION`)
- PHP: `packages/php/vX.Y.Z`
