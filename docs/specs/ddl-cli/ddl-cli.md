---
type: Spec
title: DDL 生成 CLI（median migrate dump）
description: ホスト migrate 向けに media テーブル DDL を stdout へ出す CLI。
tags: [median, specs, ddl-cli, cli, migrate, ddl]
timestamp: 2026-09-27T05:10:00Z
---

# DDL 生成 CLI（`median migrate dump`）

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.9.0`
- **前提**: [../db-media/db-media.md](../db-media/db-media.md) / [../../tests/ddl-cli/ddl-cli.md](../../tests/ddl-cli/ddl-cli.md)
- **関連**: [../../../migrations/README.md](../../../migrations/README.md) / [../go-module-path/go-module-path.md](../go-module-path/go-module-path.md)

## 目的

ホストアプリの migrate パイプラインに載せる SQL を、median の正本 DDL（`CreateMediaSQL`）から生成する。goose embed は採用しない。

## CLI 契約

```bash
median migrate dump --dialect postgres --id-strategy auto_increment --table media
```

| 引数 | 必須 | 既定 | 値 |
| --- | --- | --- | --- |
| `--dialect` | はい | — | `mysql` \| `postgres` \| `sqlite3` |
| `--id-strategy` | いいえ | `auto_increment` | `auto_increment` \| `uuid_v4` \| `uuid_v7` \| `ulid` |
| `--table` | いいえ | `media` | 識別子（`NormalizeTableName` と同じ） |

- stdout: SQL のみ
- stderr: usage / エラー
- 終了コード: 成功 0、不正時非ゼロ
- SQL は同引数の `CreateMediaSQL` / `createMediaSQL` と一致

## 配布

| 言語 | 入口 |
| --- | --- |
| Go | `packages/go/cmd/median` — `go run` / `go install` |
| JS | `@b4moss/median` の `bin.median` — `npx @b4moss/median` |

## ホスト組み込み

1. CLI で dump（引数は実行時 Config と一致させる）
2. ホスト `migrations/` にコミット
3. ホスト流儀で migrate up
4. median を初期化（同じ table / id strategy）

既存表＋カラムマップ経路は CLI 対象外。

## 非目標

- goose embed
- ALTER 自動生成
- Prisma / Flyway 固有ラッパ
- DDL 三重メンテの単一ソース化
- E2E ハーネス変更

## パッケージ版

- Go: `0.9.0`（タグ `packages/go/v0.9.0`）
- npm `@b4moss/median`: `0.9.0`（ルートタグ `v0.9.0`）

----

以上
