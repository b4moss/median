---
type: Spec
title: DB メディアスキーマ（パス列・テーブル名）
description: file_path / file_name / original_file_name への再編とテーブル名／カラムマップ設定。
tags: [median, specs, v0.7.0, db, schema]
timestamp: 2026-09-27T03:23:29Z
---

# DB メディアスキーマ（パス列・テーブル名）

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.7.0`
- **前提**: [specs/db-media/db-media.md](../../../specs/db-media/db-media.md)（本版でパス列・設定を上書き）
- **関連**: [er.dbml](../../../er.dbml) / [migrations/README.md](../../../../migrations/README.md) / [tests/db-media/db-media.md](../../../tests/db-media/db-media.md) / [README.md](./README.md)

## 目的

メディア行のパス・ファイル名表現を明確にし、組み込み先でテーブル名とカラム名を変えられるようにする。

## パス／ファイル名カラム

`path` 単一カラムをやめ、次の構成にする（他カラム — `mime` / `size` / `width` / `height` / `hash` / `original_id` / `variant_key` / `created_by` / `owned_by` / `status` / `created_at` 等 — は維持）。

| カラム | 役割 |
| --- | --- |
| `file_path` | ストレージルートからの相対パス。Adapter の Put/Get/Delete キー。UNIQUE |
| `file_name` | 最終 basename（`ResolveFilename` 後）。`path.Base(file_path)` と一致 |
| `original_file_name` | アップロード時の元ファイル名（空可）。派生行は空または NULL |
| `hash` | SHA-256 hex。重複抑止の参照用 |

### `hash` の一意制約

- **UNIQUE にしない**（INDEX のみ）
- 重複抑止はアプリ層（`FindByHash` / `RejectDuplicate`）
- 部分 UNIQUE は方言差のため当面スコープ外

### Store 時の書き込み

- `file_path` = shardian 相対パス
- `file_name` = ResolveFilename 結果
- `original_file_name` = 入力 filename（派生は空）
- 公開 API の `StoreResult.path` 等は**リネームしない**（値は `file_path` と一致）

## テーブル名とカラムマップ

- Config でテーブル名を指定可能（デフォルト `"media"`）
- Repo・DDL 生成（Up/Down）は指定テーブル名を使う
- 必須カラムの論理名→物理列名マップを許容（未指定キーはデフォルト物理名）
- median 自前 DDL の列名は常にデフォルト物理名。マップは既存テーブル向け Repo 専用

## マイグレーション

- 物理正本は `migrations/`（`00001` を置き換え。ALTER `00002` は出さない）
- Go / JS の DDL 生成は同内容に同期
- ホスト migrate 連携の現行正本は [`specs/ddl-cli/ddl-cli.md`](../../../specs/ddl-cli/ddl-cli.md)（`median migrate dump`）。本版時点ではスコープ外だった

## パッケージ版

- Go: `0.7.0`（タグ `packages/go/v0.7.0`）— 現行 Go は `0.9.0` を参照
- npm `@b4moss/median`: `0.7.0`（ルートタグ `v0.7.0`）— 現行 npm は `0.9.0` を参照

----

以上
