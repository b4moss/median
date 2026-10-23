# E2E (v0.6.0)

End-to-end tests against **RustFS** (S3-compatible) and a **Docker-shared POSIX** directory for Local FS. Scenarios: L1 / S1 / C-L / C-S / T-L / T-S — see [`docs/tests/e2e/e2e.md`](../docs/tests/e2e/e2e.md).

**Not** part of regular PR/push CI. Run locally before version bumps; GitHub Actions via **workflow_dispatch** (`.github/workflows/e2e.yml`).

## Requirements

- Docker Compose v2
- Go 1.26+ (`GOTOOLCHAIN=auto` is fine)
- Node.js 24+

## Quick start

```bash
./e2e/run.sh          # all (Go then JS)
./e2e/run.sh go
./e2e/run.sh js
```

`run.sh` will:

1. Start `rustfs` + `localfs` (bind-mounts `e2e/.local-data` for Local adapter)
2. Wait for RustFS `/health`
3. Run Go tests in `e2e/go`
4. Build `packages/js` and run `e2e/js` tests
5. Tear down compose (set `MEDIAN_E2E_KEEP=1` to keep containers)

## Environment

See [`.env.example`](./.env.example). Defaults:

| Variable | Default |
| --- | --- |
| `MEDIAN_E2E_S3_ENDPOINT` | `http://127.0.0.1:9000` |
| `MEDIAN_E2E_S3_ACCESS_KEY` | `median_e2e` |
| `MEDIAN_E2E_S3_SECRET_KEY` | `median_e2e_secret` |
| `MEDIAN_E2E_S3_BUCKET` | `median-e2e` |
| `MEDIAN_E2E_LOCAL_ROOT` | `e2e/.local-data` (created by `run.sh`) |

## Layout

```
e2e/
  docker-compose.yml   # RustFS 1.0.0 (pinned digest) + localfs
  run.sh
  go/                  # go test scenarios
  js/                  # node:test scenarios
```

----

以上
