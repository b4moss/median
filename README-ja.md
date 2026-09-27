# median

[![CI](https://github.com/b4moss/median/actions/workflows/ci.yml/badge.svg)](https://github.com/b4moss/median/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/median)](https://codecov.io/gh/b4moss/median)
[![npm](https://img.shields.io/npm/v/@b4moss/median)](https://www.npmjs.com/package/@b4moss/median)
[![Release](https://img.shields.io/github/v/release/b4moss/median)](https://github.com/b4moss/median/releases)
[![License](https://img.shields.io/github/license/b4moss/median)](https://github.com/b4moss/median/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/b4moss/median/badge)](https://scorecard.dev/viewer/?uri=github.com/b4moss/median)

DDD 向けメディアストア。バイト列をストレージへ出し入れする契約を、言語横断で共通化する。

- [English README](./README.md)

中心は **Store / Delete / Get**（対応時は **PresignGet**）。Local FS・S3 Adapter、任意のメディア DB、画像パイプライン（圧縮・リサイズ・サムネ、SVG サニタイズ、PDF 先頭ページサムネ）を提供する。

版番号は **言語ごとに独立**。計画マイルストーン名と各パッケージの SemVer は一致しなくてよい（例: Go `0.4.0` と npm `0.5.0`）。

## TypeScript / JavaScript

npm パッケージ: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)**（`packages/js`）

現行: `packages/js/package.json`（現在 **`0.5.0`**）。**Node.js 24+** が必要。

```bash
npm install @b4moss/median
```

```bash
cd packages/js
npm ci
npm test
npm run build
```

API: [`packages/js/README.md`](./packages/js/README.md)  
受け入れテスト: [`docs/tests/v0.5.0.md`](./docs/tests/v0.5.0.md) / 仕様: [`docs/specs/v0.5.0/`](./docs/specs/v0.5.0/)

**リリース:** ルートタグ `vX.Y.Z` は `package.json` の `version` と一致させる。そのツリーを `release` に push すると、registry と内容が異なるときだけ npm publish（[`.github/CI.md`](./.github/CI.md)）。GitHub Actions の Trusted Publishing を利用。

## Go

Go モジュール: **[github.com/b4moss/median/go](./packages/go)**（`packages/go`）

現行: `packages/go/VERSION` → **`0.4.0`**、タグ **`packages/go/v0.4.0`**。

```bash
go get github.com/b4moss/median/go@v0.4.0
```

```bash
cd packages/go
go test ./...
```

| Import path | 役割 |
|-------------|------|
| `github.com/b4moss/median/go/core` | `New` / Store / Delete / Get / PresignGet |
| `github.com/b4moss/median/go/storage/...` | Local / S3 Adapter |
| `github.com/b4moss/median/go/db` | マイグレーション補助 + MediaRepo |
| `github.com/b4moss/median/go/pipeline` | 画像 / SVG / PDF パイプライン |

利用: [`packages/go/README.md`](./packages/go/README.md)  
仕様: [`docs/specs/v0.2.0/`](./docs/specs/v0.2.0/)〜[`v0.4.0/`](./docs/specs/v0.4.0/) / テスト: [`docs/tests/`](./docs/tests/)

**リリース:** `VERSION` に合わせたタグ `packages/go/vX.Y.Z`。ルートタグ `vX.Y.Z` だけでは Go は公開されない。

## PHP

未着手（`v0.6.0`）。`packages/php` 予定 — [`docs/plans/v0.6.0/`](./docs/plans/v0.6.0/)

## ドキュメント

| 文書 | 内容 |
|-----|----------|
| [`docs/main.md`](./docs/main.md) | プロダクトハブ（目的・スコープ・技術方針） |
| [`docs/roadmap.md`](./docs/roadmap.md) | マイルストーン |
| [`docs/specs/`](./docs/specs/) | 出荷済仕様（v0.1〜v0.5） |
| [`docs/plans/`](./docs/plans/) | 今後の計画（v0.6 PHP） |
| [`docs/tests/`](./docs/tests/) | TDD 受け入れ仕様 |
| [`docs/charter/`](./docs/charter/) | 憲章（Git / SemVer / TDD など） |
| [`.github/CI.md`](./.github/CI.md) | CI/CD 方針 |

## ライセンス

MIT © Bicycle for Mind LLC. — [`LICENSE`](./LICENSE) を参照。
