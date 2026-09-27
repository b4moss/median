---
type: Spec
title: DB・メディアメタデータ
description: crudian・カラム・削除・原子性・採番の仕様詳細。
tags: [median, specs, db-media, db]
timestamp: 2026-09-25T06:30:00Z
---

# DB・メディアメタデータ

- **状態**: 出荷済（現行仕様）
- **導入**: `v0.3.0`（パス列・テーブル設定は `v0.7.0` で現行形）
- **関連**: [er.dbml](../../er.dbml) / [../core-api/core-api.md](../core-api/core-api.md) / [../storage/storage.md](../storage/storage.md) / [../media-pipeline/media-pipeline.md](../media-pipeline/media-pipeline.md) / [../../tests/db-media/](../../tests/db-media/)

## DB スタック（言語別）

### Go

- 行操作は **b4moss/crudian**（gorm）経由
- 対応 Dialects（DDL / goose）: **MySQL / MariaDB / Postgres / SQLite**（libSQL は SQLite 系）
- 呼び出し側が `*gorm.DB` / Crud を注入できる。未注入時は Config DSN から内部生成してもよい

### JS（`packages/js`）

- **crudian は使わない**。自前 `MediaRepo` + **better-sqlite3**（実行時）
- DDL 文字列生成は MySQL / Postgres / SQLite 向けを持つが、ランタイム Repo は **SQLite のみ**

## スキーマ方式

1. 独自テーブルをリポジトリ直下 `migrations/` でマイグレーション（論理 ER は [er.dbml](../../er.dbml)）
2. 既存メディアテーブルへ、必須カラムをマッピング

### マイグレーションの書き方

- Go は **goose** を用いる
- **方言別 SQL を並べてメンテしない**。1 バージョン = 1 定義とし、goose の dialect（`mysql` / `postgres` / `sqlite3`）に応じて解釈・展開する
  - MariaDB は goose / driver 上 `mysql` として扱う
- 正本の実装形: [`migrations/00001_create_media.go`](../../../migrations/00001_create_media.go)（詳細は [`migrations/README.md`](../../../migrations/README.md)）
- 採番方式も **同一定義内の条件分岐**（別マイグレーション版にしない）。切替は `MEDIAN_ID_STRATEGY`（値は下記ワイヤ形式）

論理 ER の正本: [er.dbml](../../er.dbml)（適用後の物理スキーマは migrations が正）

## 必須カラム（マッピングの正）

| カラム | 役割 |
| --- | --- |
| `id` | 主キー（DB 使用時の正） |
| `file_path` | ストレージルートからの相対パス。Adapter の Put/Get/Delete キー。UNIQUE（storage key は持たない） |
| `file_name` | 最終 basename（`ResolveFilename` 後）。`path.Base(file_path)` と一致 |
| `original_file_name` | アップロード時の元ファイル名（空可）。派生行は空または NULL |
| `mime` | MIME タイプ |
| `size` | バイトサイズ |
| `width` | 幅 px（寸法が分かるとき。不明なら NULL） |
| `height` | 高さ px（寸法が分かるとき。不明なら NULL） |
| `hash` | SHA-256 hex。**UNIQUE にしない**（INDEX のみ）。重複抑止はアプリ層 |
| `created_at` | 作成日時 |
| `original_id` | オリジナルへの自己参照。派生は子行 |
| `variant_key` | 派生のプリセット key（オリジナルは NULL。サムネイル key 名） |
| `created_by` | 作成者（Store 時は `owned_by` と同値） |
| `owned_by` | 所有者（Store 時は `created_by` と同値） |
| `status` | 状態（text。アプリ側定義。median の Delete では使わない） |

## `id` 採番

| ワイヤ値（`MEDIAN_ID_STRATEGY` / Config） | 生成 | 物理型 |
| --- | --- | --- |
| `auto_increment`（**デフォルト**） | DB 採番 | integer / bigint |
| `uuid_v4` | UUID v4 文字列 | text |
| `uuid_v7` | UUID v7 文字列 | text |
| `ulid` | **現行実装は UUID v7 文字列**（真の ULID ライブラリは未使用。DDL は text ID 分岐） | text |

- `original_id` の物理型は `id` に合わせる
- マイグレーションは方式ごとに版を増やさず、**同一 `00001` 内の分岐**で DDL を選ぶ

## actor（`created_by` / `owned_by`）

- Store 時は両カラムに**同値**を挿入
- 未指定時は config の default actor を使う
- default actor も無ければエラー

## SHA-256 重複

- DB 使用時、同一 hash の再 Store はデフォルトで**既存レコード返却**（成功）
- オプションで拒否に切替可

## Delete

- **物理削除**（行とストレージファイルを削除）
- 親削除時、`original_id` 参照の子レコードと子ファイルもカスケード削除
- 論理削除は当面スコープ外

## DB とストレージの原子性

- ライブラリが補償する（片方だけ成功した場合、可能な範囲で戻す）
- **プロセスクラッシュ等で補償処理自体が中断した場合は責務外**（呼び出し側の再実行・運用で扱う）

## パス／ファイル名カラム（現行）

`path` 単一カラムは廃止。上記 `file_path` / `file_name` / `original_file_name` を用いる。

### Store 時の書き込み

- `file_path` = shardian 相対パス
- `file_name` = ResolveFilename 結果
- `original_file_name` = 入力 filename（派生は空）
- 公開 API の `StoreResult.path` 等は**リネームしない**（値は `file_path` と一致）

### `hash` の一意制約

- **UNIQUE にしない**（INDEX のみ）
- 重複抑止はアプリ層（`FindByHash` / `RejectDuplicate`）
- 部分 UNIQUE は方言差のため当面スコープ外

## テーブル名とカラムマップ

- Config でテーブル名を指定可能（デフォルト `"media"`）
- Repo・DDL 生成（Up/Down）は指定テーブル名を使う
- 必須カラムの論理名→物理列名マップを許容（未指定キーはデフォルト物理名）
- median 自前 DDL の列名は常にデフォルト物理名。マップは既存テーブル向け Repo 専用

## マイグレーション（パス列再編）

- 物理正本は `migrations/`（`00001` を置き換え。ALTER `00002` は出さない）
- Go / JS の DDL 生成は同内容に同期
- ホスト migrate 連携（DDL 出力 CLI・goose embed 等）はスコープ外
- 既存 DB への ALTER マイグレーションは提供しない

## パッケージ版（パス列再編時点）

- Go: `0.7.0`（タグ `packages/go/v0.7.0`）
- npm `@b4moss/median`: `0.7.0`（ルートタグ `v0.7.0`）

----

以上
