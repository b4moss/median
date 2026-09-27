# CI / CD notes (median)

Aligned with `docs/charter/git-rule.md` and peer `-an` libraries (shardian / crudian).

## Current publish targets (checkpoint)

| Artifact | Version | How |
| --- | --- | --- |
| `@b4moss/median` (npm) | `0.7.0` | Push `packages/js/**` to `release` + root tag `vX.Y.Z`（Trusted Publisher） |
| `github.com/b4moss/median/packages/go` | `0.8.0` | Tag `packages/go/vX.Y.Z`（module path = ディレクトリ／タグ接頭辞） |
| E2E harness (`e2e/`) | `v0.6.0` | Tag `v0.6.0`（milestone release。Go/npm の SemVer は上げない。`release` ブランチ不要） |
| DB media schema | `v0.7.0` | Tags `v0.7.0` + `packages/go/v0.7.0`（出荷済） |
| Go module path | `v0.8.0` | Tag `packages/go/v0.8.0`（`github.com/b4moss/median/packages/go`） |

## Workflows

| Workflow | Role |
| --- | --- |
| `ci.yml` | PR/push tests for `packages/{go,js,php}` with path filters (docs-only does not start CI). Needs `pull-requests: read` for dorny/paths-filter on PRs |
| `e2e.yml` | **Manual only** (`workflow_dispatch`). RustFS + LocalFS E2E for Go and JS (`e2e/run.sh`). Not a PR gate — see `docs/tests/v0.6.0.md` / `docs/specs/v0.6.0/` |
| `codeql.yml` | CodeQL on **PR → main only** (crudian と同型). **Private 現状**: `upload: never` + 言語は `actions` のみ。**Public / GHAS 後**: workflow 内のコメントアウトを外す（go / js-ts matrix と SARIF upload） |
| `scorecard.yml` | OpenSSF Scorecard on default branch |
| `release-on-tag.yml` | GitHub Release for `v*`, `packages/go/v*`, `packages/php/v*` |
| `publish-npm.yml` | Publish `@b4moss/median` from `release` when JS changed |
| `publish-go.yml` | GitHub Release + proxy ping for `github.com/b4moss/median/packages/go` |

## Runtimes (CI)

- Go `1.26.x`
- Node `24`
- PHP `8.2`

## Tags

- JS: `vX.Y.Z` (matches `packages/js/package.json`)
- Go: `packages/go/vX.Y.Z` (matches `packages/go/VERSION`)
- PHP: `packages/php/vX.Y.Z`
