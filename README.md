# median

ファイル操作を抽象化する DDD 向けライブラリ。バイト列の Store / Delete / Get を中心に、Local FS・S3、任意のメディア DB、画像パイプライン（圧縮・リサイズ・サムネ・SVG サニタイズ・PDF 先頭ページサムネ）を、言語横断で同じ契約語彙で提供する。

## 現状（区切り: v0.5.0）

| 言語 | パス | 版 | 配布 |
| --- | --- | --- | --- |
| Go | [`packages/go`](packages/go/) | `0.4.0` | モジュール `github.com/b4moss/median/go`（タグ `packages/go/v0.4.0`） |
| TypeScript | [`packages/js`](packages/js/) | `0.5.0` | npm [`@b4moss/median`](https://www.npmjs.com/package/@b4moss/median)（タグ `v0.5.0`） |
| PHP | `packages/php` | — | 未着手（`v0.6.0`） |

## クイックスタート

### Go

```bash
cd packages/go
go test ./...
```

```go
import "github.com/b4moss/median/go/core"
```

詳細: [packages/go/README.md](packages/go/README.md)

### TypeScript (Node.js 24+)

```bash
cd packages/js
npm ci
npm test
npm run build
```

```bash
npm install @b4moss/median
```

詳細: [packages/js/README.md](packages/js/README.md)

## ドキュメント

| 文書 | 内容 |
| --- | --- |
| [docs/main.md](docs/main.md) | 目的・スコープ・技術方針 |
| [docs/roadmap.md](docs/roadmap.md) | マイルストーン |
| [docs/specs/](docs/specs/) | **現行仕様**（v0.1〜v0.5） |
| [docs/plans/](docs/plans/) | これからやる計画（v0.6 PHP） |
| [docs/tests/](docs/tests/) | TDD 入力のテスト仕様 |
| [docs/charter/](docs/charter/) | 憲章（Git / SemVer / TDD 等） |
| [.github/CI.md](.github/CI.md) | CI / タグ / 公開のメモ |

## 開発メモ

- ブランチ運用: `feat` → `dev-vX.Y.Z` → `develop` → `main` →（パッケージ）`release`
- JS 公開: `release` への `packages/js/**` push + ルートタグ `vX.Y.Z`（Trusted Publisher）
- Go 公開: タグ `packages/go/vX.Y.Z`

## ライセンス

リポジトリおよび各パッケージの表記に従う。
