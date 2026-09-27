# median

[![CI](https://github.com/b4moss/median/actions/workflows/ci.yml/badge.svg)](https://github.com/b4moss/median/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/median)](https://codecov.io/gh/b4moss/median)
[![npm](https://img.shields.io/npm/v/@b4moss/median)](https://www.npmjs.com/package/@b4moss/median)
[![Release](https://img.shields.io/github/v/release/b4moss/median)](https://github.com/b4moss/median/releases)
[![License](https://img.shields.io/github/license/b4moss/median)](https://github.com/b4moss/median/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/b4moss/median/badge)](https://scorecard.dev/viewer/?uri=github.com/b4moss/median)

DDD 向けメディアストア。バイト列をストレージへ出し入れする契約を、**言語横断でひとつ**に揃える。

- [English README](./README.md)

## できること

中心は **Store / Delete / Get**（対応時は **PresignGet**）。あわせて:

- **Local FS** / **S3 互換** Adapter
- 任意の **メディア DB**（テーブル名・カラムマップ）
- 画像パイプライン（圧縮・リサイズ・サムネ）、**SVG サニタイズ**、**PDF** 先頭ページサムネ（レンダラ注入）

版番号は **言語ごとに独立**。現行ラインは各ポートの README / `VERSION` / `package.json` を見る。いま Go / npm とも **`0.9.0`**（DDL CLI）。

## 言語ポートを選ぶ

| ポート | パッケージ | 言語ハブ（詳細はここ） |
|--------|------------|------------------------|
| **Go** | `github.com/b4moss/median/packages/go` | [`packages/go/README.md`](./packages/go/README.md) |
| **TypeScript** | `@b4moss/median` | [`packages/js/README.md`](./packages/js/README.md) |
| **PHP** | — | 未割当 — [`docs/plans/unscheduled/packages-php.md`](./docs/plans/unscheduled/packages-php.md) |

インストール・import・Config 例・CLI フラグの詳細は **言語ハブ** に委譲。このルート README は共通のホスト組み込み手順。

---

## ホストでの使い方（ステップバイステップ）

本番で「ホストが migrate を持つ」場合の推奨手順です。

### ステップ 1 — ポートを選んでインストール

- Go: [`packages/go/README.md`](./packages/go/README.md)（`go get …/packages/go@v0.9.0`）
- JS: [`packages/js/README.md`](./packages/js/README.md)（`npm install @b4moss/median`）

### ステップ 2 — テーブル名と ID 戦略を決める

| 項目 | デフォルト |
|------|------------|
| テーブル | `media` |
| ID 戦略 | `auto_increment`（`uuid_v4` / `uuid_v7` / `ulid` 可） |
| 方言 | `mysql` \| `postgres` \| `sqlite3` のいずれか |

DDL dump と実行時 Config で **同じ値**を使います。

### ステップ 3 — DDL をホストの migrations に出す

median は本番の migrate を代行しません。SQL を生成し、ホスト側の migrate（goose / Prisma / Flyway / knex 等）にコミットします。

```bash
# Go
go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 \
  migrate dump --dialect postgres --id-strategy auto_increment --table media \
  > migrations/XXXX_median_media.sql

# JS
npx @b4moss/median migrate dump \
  --dialect postgres --id-strategy auto_increment --table media \
  > migrations/XXXX_median_media.sql
```

- **stdout** = SQL のみ（リダイレクト向き）
- **stderr** = エラー / usage
- 仕様: [`docs/specs/ddl-cli/`](./docs/specs/ddl-cli/)

その後、いつもどおり `migrate up`。

**既存テーブル**で列名が違う場合は dump せず、Config のカラムマップを使う（[`docs/specs/db-media/`](./docs/specs/db-media/)）。

### ステップ 4 — migrate 適用後にアプリを起動

順番:

1. ホスト migrate が media テーブルを作る  
2. アプリが同じ `table` / `id-strategy` で median を初期化  
3. Store / Get / Delete を呼ぶ  

`New` / `createMedian` の配線は各言語ハブへ。

### ステップ 5 — Store → Get → Delete

概念（疑似）:

```text
Store(bytes, size, { mime, filename, … })
  → ストレージへ書き込み
  → 必要なら media 行を挿入（file_path / file_name / …）
  → { path, hash, id?, variants?, … } を返す

Get(path) → オブジェクト取得（DB 連携時は補強あり）
Delete(path) → オブジェクト削除（設定により DB・派生も）
```

公開 API のフィールド名 `path` はストレージ相対キーのまま。DB 物理列は `file_path`（v0.7+）。

### ステップ 6 — 任意: サムネ / SVG / PDF / S3

パイプラインとストレージは言語側 Config で設定。制約（例: PDF レンダラは注入必須）は言語ハブと [`docs/specs/`](./docs/specs/) を参照。

### ステップ 7 — リリース前の確認

- 単体: `packages/go` → `go test ./…` · `packages/js` → `npm test`
- E2E（RustFS + LocalFS、Go+JS）: `./e2e/run.sh` または Actions [`e2e.yml`](./.github/workflows/e2e.yml) の `workflow_dispatch` — [`docs/specs/e2e/`](./docs/specs/e2e/)

---

## CLI 早見表

```text
median migrate dump --dialect <mysql|postgres|sqlite3> [--id-strategy …] [--table …]
```

| | Go | JS |
|--|----|----|
| 実行 | `go run …/cmd/median@v0.9.0 migrate dump …` | `npx @b4moss/median migrate dump …` |
| ライブラリ同等 | `db.CreateMediaSQL` | `createMediaSQL` |

詳細は言語ハブ · [`migrations/README.md`](./migrations/README.md)。

---

## ドキュメント地図

| 文書 | 内容 |
|------|------|
| [`docs/README.md`](./docs/README.md) | 製品 pillar（目的・スコープ・技術方針） |
| [`docs/roadmap.md`](./docs/roadmap.md) | マイルストーン |
| [`docs/specs/`](./docs/specs/) | 現行仕様（ドメイン切り） |
| [`docs/plans/`](./docs/plans/) | 未実装計画（PHP unscheduled） |
| [`docs/tests/`](./docs/tests/) | TDD 受け入れ |
| [`docs/charter/`](./docs/charter/) | 憲章・OKF |
| [`migrations/`](./migrations/) | スキーマ正本の説明 |
| [`e2e/`](./e2e/) | E2E ハーネス |
| [`.github/CI.md`](./.github/CI.md) | CI/CD・公開タグ |

## License

MIT © Bicycle for Mind LLC. — [`LICENSE`](./LICENSE)
