---
type: Reference
title: migrations
description: 方言解釈型gooseマイグレーションの正本説明。
tags: [median, migrations, goose]
timestamp: 2026-09-27T03:23:29Z
---

# migrations

スキーママイグレーションの正本ディレクトリ。論理 ER は [`docs/er.dbml`](../docs/er.dbml)。

## 対応 Dialects（crudian に合わせる）

| 対象 | goose dialect / driver |
| --- | --- |
| MySQL | `mysql` |
| MariaDB | `mysql`（MySQL 互換ドライバ） |
| Postgres | `postgres` |
| SQLite | `sqlite3` |

libSQL を使う場合は SQLite 系として扱う（crudian の libSQL アダプタに合わせる）。

## 書き方（必須）

- **方言別 SQL を手で増やして並べない**
- マイグレーションは **1 バージョン = 1 定義**とし、goose 実行時の dialect に応じてツール／コードが解釈・展開する
- Go では **goose の Go migration** を正とする（`SetDialect` / `Provider` の dialect を見て DDL を切り替える）
- 型の対応（デフォルト採番 = auto increment）:

| 論理（er.dbml） | MySQL / MariaDB | Postgres | SQLite |
| --- | --- | --- | --- |
| `id` bigint PK AI | `BIGINT AUTO_INCREMENT` | `BIGSERIAL` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `original_id` | `BIGINT NULL` | `BIGINT NULL` | `INTEGER NULL` |
| `file_path` varchar(2048) | `VARCHAR(2048)` | `VARCHAR(2048)` | `TEXT` |
| `file_name` varchar(512) | `VARCHAR(512)` | `VARCHAR(512)` | `TEXT` |
| `original_file_name` varchar(512) | `VARCHAR(512) NULL` | `VARCHAR(512) NULL` | `TEXT NULL` |
| `mime` varchar(255) | `VARCHAR(255)` | `VARCHAR(255)` | `TEXT` |
| `size` bigint | `BIGINT` | `BIGINT` | `INTEGER` |
| `width` / `height` | `INT NULL` | `INT NULL` | `INTEGER NULL` |
| `hash` char(64) | `CHAR(64)` | `CHAR(64)` | `TEXT` |
| `created_by` / `owned_by` | `VARCHAR(255)` | `VARCHAR(255)` | `TEXT` |
| `variant_key` | `VARCHAR(64) NULL` | `VARCHAR(64) NULL` | `TEXT NULL` |
| `status` text | `TEXT` | `TEXT` | `TEXT` |
| `created_at` timestamptz | `DATETIME(6)` | `TIMESTAMPTZ` | `TEXT`（ISO-8601） |

制約: `file_path` は UNIQUE。`hash` は INDEX のみ（UNIQUE にしない）。

UUID / ULID / UUID 採番を選ぶ場合も **別バージョンにせず**、同じ `00001` 定義内の条件分岐で `id` / `original_id` を **TEXT** にする（方言共通）。

実行時は環境変数で切替:

| 変数 | 値 | 意味 |
| --- | --- | --- |
| `MEDIAN_ID_STRATEGY` | `auto_increment`（デフォルト） / `uuid_v4` / `uuid_v7` / `ulid` | PK 型 |
| `MEDIAN_TABLE_NAME` | 識別子（デフォルト `media`） | テーブル名（goose 正本） |

- Go / JS の DDL 生成は同内容に同期
- **ホストへの適用**: [`median migrate dump`](../docs/specs/ddl-cli/ddl-cli.md) で SQL を出し、ホストの migrate に載せる（公式）
- ルートでの `goose -dir migrations ...` 直叩きは参考（開発・検証用）。ライブラリ配布物には goose 正本は含まれない
- ランタイムの `MigrateUp` / `migrateUp` はテスト・使い捨て DB 向け

## ファイル

| ファイル | 内容 |
| --- | --- |
| [`00001_create_media.go`](./00001_create_media.go) | media テーブル作成（dialect × id strategy × table name） |

## ホスト向け（推奨）

```bash
# Go
go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 \
  migrate dump --dialect postgres --id-strategy auto_increment --table media \
  > migrations/XXXX_median_media.sql

# JS
npx @b4moss/median@0.9.0 migrate dump --dialect postgres --table media \
  > migrations/XXXX_median_media.sql
```

引数はアプリ側 median Config（テーブル名・ID 戦略）と一致させる。

## 参考: ルート goose（クローン／開発用）

```bash
# 例: Postgres
goose -dir migrations postgres "$DATABASE_URL" up

# 例: MySQL / MariaDB（parseTime / multiStatements を有効化）
goose -dir migrations mysql "$DSN" up

# 例: SQLite
goose -dir migrations sqlite3 "$DB_PATH" up
```

----

以上
