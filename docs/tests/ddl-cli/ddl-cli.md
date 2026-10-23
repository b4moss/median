---
type: Spec
title: テスト仕様 ddl-cli
description: DDL 生成 CLI（median migrate dump）の TDD 入力。Go / JS。
tags: [median, tests, ddl-cli, cli, migrate, ddl, go, js]
timestamp: 2026-09-27T04:40:00Z
---

# テスト仕様 — DDL 生成 CLI

対象マイルストーン: `v0.9.0`（DDL 生成 CLI）  
製品: [`../../README.md`](../../README.md)（pillar）  
仕様: [`../../specs/ddl-cli/`](../../specs/ddl-cli/)（[ddl-cli.md](../../specs/ddl-cli/ddl-cli.md)）  
前提（出荷契約）: [`../db-media/db-media.md`](../db-media/db-media.md)（スキーマ・`CreateMediaSQL` / `MigrateUp`）/ [`../go-module-path/go-module-path.md`](../go-module-path/go-module-path.md)  
ロードマップ: [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`../../charter/tdd.md`](../../charter/tdd.md)（氷山パターン）

## 共通前提

| 項目 | 値 |
| --- | --- |
| 対象 | Go `packages/go/cmd/median` と JS `packages/js`（`bin`: `median`）の**両方**。DDL 本体は既存 `CreateMediaSQL` / `createMediaSQL` |
| パッケージ SemVer | Go **`0.9.0`**（タグ `packages/go/v0.9.0`）/ npm **`0.9.0`**（ルートタグ `v0.9.0`） |
| 入口（CLI） | `median migrate dump` |
| 入口（ライブラリ） | 変更なし（`MigrateUp` / `migrateUp` は残置・本版で新規要件なし） |
| 破壊的変更 | なし（additive）。ホスト migrate の公式経路を CLI に定める |

### CLI 契約（実装固定）

| 項目 | 値 |
| --- | --- |
| サブコマンド | `migrate dump`（これ以外の `migrate` サブコマンドは本版スコープ外＝拒否） |
| `--dialect` | **必須**。`mysql` \| `postgres` \| `sqlite3` |
| `--id-strategy` | 省略時 `auto_increment`。`uuid_v4` \| `uuid_v7` \| `ulid` 可 |
| `--table` | 省略時 `media`。識別子規則は `NormalizeTableName` と同じ |
| stdout | **生成 SQL のみ**（リダイレクト前提。バナー・ログ禁止） |
| stderr | usage・エラーメッセージ |
| 終了コード | 成功 `0`。不正引数・未知サブコマンド・DDL 生成失敗は非ゼロ |
| SQL 内容 | 同一引数で `CreateMediaSQL` / `createMediaSQL` と**一致**（CLI は薄いラッパ） |

### 配布・起動（受け入れに含める）

| 言語 | 起動例 |
| --- | --- |
| Go | `go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 migrate dump ...`（または同バイナリの `go install` 成果物） |
| JS | `npx @b4moss/median migrate dump ...`（`package.json` の `bin.median` → build 済み CLI） |

本版で E2E ハーネス新規要件は追加しない（Compose / Workflow は v0.6.0 契約のまま）。

---

## CLI — `migrate dump`

### C1 — 正常系（方言 × デフォルト）

- **Given** 有効な `--dialect`（`sqlite3` / `postgres` / `mysql` 各々）
- **And** `--id-strategy` / `--table` を省略
- **When** `median migrate dump --dialect <d>` を実行する
- **Then** 終了コード `0`
- **And** stdout に `CREATE TABLE` とデフォルトテーブル名 `media` が含まれる
- **And** stdout に `file_path` / `file_name` / `original_file_name` が含まれる（旧列 `path` をテーブル列として出さない）
- **And** stdout の内容が、同引数の `CreateMediaSQL` / `createMediaSQL` の戻り値と一致する（末尾改行の有無は実装で固定し、テストで明示）
- **And** stderr に SQL 本文を流さない

### C2 — ID 戦略

- **Given** `--dialect sqlite3`
- **When** `--id-strategy auto_increment`（明示）および省略
- **Then** いずれも整数系 PK の DDL（SQLite なら `INTEGER PRIMARY KEY` 系）になる
- **When** `--id-strategy uuid_v4`（または `uuid_v7` / `ulid`）
- **Then** text 系 PK の DDL になる
- **And** いずれも `CreateMediaSQL` と同内容

### C3 — テーブル名

- **Given** `--dialect sqlite3 --table medias`
- **When** dump する
- **Then** stdout の `CREATE TABLE` 対象が `medias`（デフォルト `media` テーブルは作る SQL にならない）
- **When** `--table` に不正識別子（例: `bad-name`）
- **Then** 非ゼロ終了。stdout は空（または SQL を出さない）。stderr に理由

### C4 — 異常系（引数）

- **When** `--dialect` を省略する
- **Then** 非ゼロ。stderr に usage または必須である旨
- **When** `--dialect oracle` 等の未対応値
- **Then** 非ゼロ（ライブラリと同じ拒否）
- **When** `--id-strategy nope`
- **Then** 非ゼロ
- **When** 未知トップレベル（例: `median foo`）または `median migrate` のみ / `median migrate other`
- **Then** 非ゼロ

### C5 — stdout / stderr 分離（リダイレクト）

- **Given** 正常な dump
- **When** stdout のみをファイルへリダイレクトする（stderr は捨てる／別キャプチャ）
- **Then** ファイル内容がそのままホスト migration に貼れる SQL である（プロンプトや INFO 行が混ざらない）

---

## パッケージ配線

### P1 — Go

- **Given** `packages/go/cmd/median`
- **Then** `go test`（cmd 配下または同等）で C1–C4 相当が通る
- **And** `packages/go/VERSION` が `0.9.0`
- **And** module path は `github.com/b4moss/median/packages/go` のまま（v0.8.0 契約維持）

### P2 — JS

- **Given** `packages/js/package.json`
- **Then** `"version": "0.9.0"`
- **And** `"bin": { "median": "./dist/cli.js" }`（または計画どおりのパス）がある
- **And** build 後の CLI エントリに Node shebang がある（または `bin` 実行時に解釈される）
- **And** `test/cli.test.ts`（または同等）で C1–C4 相当が通る

### P3 — ライブラリ回帰（スモーク）

- **Given** 既存 `db` の `CreateMediaSQL` / `MigrateUp` テスト
- **Then** v0.7.0 契約の単体が落ちない（CLI 追加で DDL 本体を壊さない）

---

## ホスト組み込み（文書・手動受け入れ）

自動 CI 必須ではないが、出荷前チェックリストとする。

### H1 — 新規表ホスト

- **Given** CLI で `--dialect` / `--id-strategy` / `--table` をホスト Config と一致させて dump した SQL
- **When** ホストの migrate ツールに 1 ファイルとして追加し up する
- **Then** 指定テーブルが作成され、median を同じテーブル名・ID 戦略で初期化して Store できる

### H2 — 既存表＋カラムマップ

- **Then** CLI を使わず、従来のテーブル名／カラムマップで組み込む（本版の CLI 対象外である旨が docs にある）

---

## 版外（本ファイルで検証しない）

- goose embed / ホストへの自動 migrate プラグイン
- ALTER 既存 DB（`path` → `file_path` 等）の自動生成
- Prisma / Flyway 等のホスト固有ラッパの配布
- E2E ハーネスの変更
- DDL 三重メンテ（`migrations/00001` / Go / JS）の単一ソース化

----

以上
