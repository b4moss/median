# DB・メディアメタデータ

- **状態**: 仕様詳細
- **マイルストーン**: `v0.3.0`
- **前提**: `v0.2.0`
- **関連**: [er.dbml](../../er.dbml) / [v0.2.0/core-api.md](../v0.2.0/core-api.md) / [v0.2.0/storage.md](../v0.2.0/storage.md)

## crudian

- DB 操作は **b4moss/crudian** を用いる
- 対応 Dialects（crudian と同じ系統）: **MySQL / MariaDB / Postgres / SQLite**（libSQL は SQLite 系として扱う）
- 呼び出し側が CRUD / DB ハンドルを注入できる
- 注入がなければ config から DB / crudian を内部生成してもよい

## スキーマ方式

1. 独自テーブルをリポジトリ直下 `migrations/` でマイグレーション（論理 ER は [er.dbml](../../er.dbml)）
2. 既存メディアテーブルへ、必須カラムをマッピング

### マイグレーションの書き方

- Go は **goose** を用いる
- **方言別 SQL を並べてメンテしない**。1 バージョン = 1 定義とし、goose の dialect（`mysql` / `postgres` / `sqlite3`）に応じて解釈・展開する
  - MariaDB は goose / driver 上 `mysql` として扱う
- 正本の実装形: [`migrations/00001_create_media.go`](../../../migrations/00001_create_media.go)（詳細は [`migrations/README.md`](../../../migrations/README.md)）
- 採番方式（auto increment / UUID / ULID）も **同一定義内の条件分岐**（別マイグレーション版にしない）。切替は `MEDIAN_ID_STRATEGY`

論理 ER の正本: [er.dbml](../../er.dbml)（適用後の物理スキーマは migrations が正）

## 必須カラム（マッピングの正）

| カラム | 役割 |
| --- | --- |
| `id` | 主キー（DB 使用時の正） |
| `path` | ストレージルートからの相対パス（storage key は持たない） |
| `mime` | MIME タイプ |
| `size` | バイトサイズ |
| `hash` | SHA-256（重複抑止） |
| `created_at` | 作成日時 |
| `original_id` | オリジナルへの自己参照。派生は子行 |
| `created_by` | 作成者（Store 時は `owned_by` と同値） |
| `owned_by` | 所有者（Store 時は `created_by` と同値） |
| `status` | 状態（text。アプリ側定義。median の Delete では使わない） |

## `id` 採番

- 選択可: `auto increment` / `UUID v4` / `UUID v7` / `ULID`
- **デフォルトは auto increment**
- 物理型:
  - `auto increment` → **integer / bigint**（デフォルト ER はこちら。`docs/er.dbml`）
  - `UUID v4` / `UUID v7` / `ULID` → **text**
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

----

以上
