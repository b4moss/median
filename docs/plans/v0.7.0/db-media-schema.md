---
type: Plan
title: DB メディアスキーマ（パス列・テーブル名）
description: file_path / file_name / original_file_name への再編とテーブル名設定可能化の方針。
tags: [median, plans, v0.7.0, db, schema]
timestamp: 2026-09-27T03:07:27Z
---

# DB メディアスキーマ（パス列・テーブル名）

- **状態**: 方針確定（マイグレーション連携は後日詳細・本マイルストーン外）
- **マイルストーン**: `v0.7.0`
- **前提**: 現行 [specs/v0.3.0/db-media.md](../../specs/v0.3.0/db-media.md) / [er.dbml](../../er.dbml)
- **関連**: [README.md](./README.md) / [roadmap.md](../../roadmap.md) / [migrations/README.md](../../../migrations/README.md) / [tests/v0.7.0.md](../../tests/v0.7.0.md)

## 目的

メディア行のパス・ファイル名表現を明確にし、組み込み先でテーブル名を変えられるようにする。

## 採用する方針（v0.7.0 スコープ）

### 1. パス／ファイル名カラムの再編

現行の `path` 単一カラムをやめ、次の構成にする（他カラム — `mime` / `size` / `width` / `height` / `hash` / `original_id` / `variant_key` / `created_by` / `owned_by` / `status` / `created_at` 等 — は維持）。

| カラム | 役割 |
| --- | --- |
| `file_path` | ストレージルートからの相対パス（現行 `path`）。Adapter の Put/Get/Delete キー。UNIQUE |
| `file_name` | 最終 basename（表示・派生命名用）。`file_path` と不整合にならないよう、書き込み時は導出またはセット検証する |
| `original_file_name` | アップロード時の元ファイル名（現行は Store 後に失われる） |
| `hash` | SHA-256 hex。重複抑止の参照用 |

#### `hash` の一意制約

- **UNIQUE にしない**（現行どおり INDEX のみ）
- 重複抑止はアプリ層（`FindByHash` / `RejectDuplicate`）のまま
- 部分 UNIQUE（オリジナル行のみ等）は方言差があるため、当面スコープ外

### 3. テーブル名 `media` の変更を許容する

- Config 等でテーブル名を指定可能にする（デフォルト `"media"`）
- Repo・DDL 生成・Down は指定名を使う
- 仕様上ある「既存メディアテーブルへ必須カラムをマッピング」を実装側で満たすため、**カラム名マッピングも併せて許容する**（最低限パス系のリネームに耐える）

## やらぬこと（本マイルストーンの範囲外）

- 必須メタデータ列（mime / size / 派生・actor 等）の削除や大幅削減
- `hash` のグローバル UNIQUE 化
- マイグレーションコマンドとホスト連携（下記「後日詳細」）

## 後日詳細（未採択・検討延期）

### 2. マイグレーションコマンドとホスト連携

他プロダクトのマイグレーションシステムへどう載せるか（DDL 出力・goose embed・ホスト任せ等）は **後日検討**。`v0.7.0` スコープ外。必要なら別マイルストーンまたは `unscheduled` で詳細化する。

## 実装時に触る正本

- `docs/er.dbml`
- `docs/specs/` の DB メディア仕様（現行は v0.3.0。完了後は本版の specs へ）
- `migrations/` および Go / JS の schema・repo
- `docs/tests/v0.7.0.md`（本マイルストーンの TDD 入力）

----

以上
