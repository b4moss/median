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
| `path` varchar(2048) | `VARCHAR(2048)` | `VARCHAR(2048)` | `TEXT` |
| `mime` varchar(255) | `VARCHAR(255)` | `VARCHAR(255)` | `TEXT` |
| `size` bigint | `BIGINT` | `BIGINT` | `INTEGER` |
| `hash` char(64) | `CHAR(64)` | `CHAR(64)` | `TEXT` |
| `created_by` / `owned_by` | `VARCHAR(255)` | `VARCHAR(255)` | `TEXT` |
| `status` text | `TEXT` | `TEXT` | `TEXT` |
| `created_at` timestamptz | `DATETIME(6)` | `TIMESTAMPTZ` | `TEXT`（ISO-8601） |

UUID / ULID / UUID 採番を選ぶ場合も **別バージョンにせず**、同じ `00001` 定義内の条件分岐で `id` / `original_id` を **TEXT** にする（方言共通）。

実行時は環境変数 `MEDIAN_ID_STRATEGY` で切替（未指定 = `auto_increment`）:

| 値 | PK 型 |
| --- | --- |
| `auto_increment`（デフォルト） | integer / bigint AI |
| `uuid_v4` / `uuid_v7` / `ulid` | text |

`v0.3.0` で `packages/go` から embed / Provider で読み込む（現行ファイルは配線前のため `//go:build ignore`）。

## ファイル

| ファイル | 内容 |
| --- | --- |
| [`00001_create_media.go`](./00001_create_media.go) | media テーブル作成（dialect × id strategy の同一定義内分岐） |

## 実行（実装後）

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
