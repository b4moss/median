---
type: Spec
title: テスト仕様 — s3-adapter
description: S3 互換 Adapter・署名付き GET の TDD 入力。
tags: [median, tests, s3-adapter, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — s3-adapter

対象ドメイン: `s3-adapter`  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/s3-adapter/`](../../specs/s3-adapter/)  
関連: [`../storage/`](../storage/) / [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`tdd.md`](../../charter/tdd.md)

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/packages/go` |
| 対象パッケージ | `core` / `db` / `pipeline` / `storage` / `storage/local` / `storage/s3` |
| ランタイム / テスト | Go `1.26.x` + `go test` |
| テスト DB | **SQLite**（gorm + modernc 系）。他 dialect はマイグレーション SQL 生成の単体確認まで |
| S3 テスト | Adapter 単体 + `httptest` モック（実 MinIO は必須にしない） |
| 入口 | `core.New(Config)` → `Store` / `Delete` / `Get` / `PresignGet` |
| 依存 | crudian（gorm）、shardian、imaging、AWS SDK v2（S3） |

### 実装固定デフォルト

| 項目 | 値 |
| --- | --- |
| crudian | 注入優先（`*gorm.DB` / Crud）。未注入時は Config DSN から内部生成可 |
| ID 戦略 | デフォルト `auto_increment`（Config / `MEDIAN_ID_STRATEGY` で切替） |
| 重複 hash | デフォルト **既存レコード返却**。`RejectDuplicate=true` で拒否 |
| actor | DB 使用時必須（`StoreOptions.Actor` または Config `DefaultActor`） |
| Delete/Get（DB あり） | 正キーは **`id`** |
| contain 余白 | デフォルト **black**。`white` / `transparent` / `#RRGGBB` / `#RRGGBBAA` 可（詳細は Resize） |
| Presign GET TTL | デフォルト **1 時間**（呼び出しで上書き可） |

v0.2.0 の Local FS・MIME・サイズ・同時アップロード・ファイル名・shardian 契約は回帰として壊さない（本版で触る経路のスモークでよい）。

Store 返却（本版・DB/パイプライン有効時に増えうる）: `{ id?, path, mime, size, hash, storageKey, width?, height?, variants? }`

---


## storage/s3

### Put / Get / Delete / Exists

- S3 互換へ相対 path で出し入れする（bucket / prefix は Config）

#### テスト：正常系

- Put したオブジェクトを Get で同じ内容として読める（モック）
- Delete 後 Exists が false（または Get が not found）
- ネストした key（shardian path）を扱える

#### テスト: 異常系

- `..` や空 path は拒否する（local と同趣旨）
- 存在しない Get は not found 相当
- モックが 5xx を返したらエラーを伝播する

### PresignGet

- GET 用署名 URL を発行する。TTL は config デフォルト 1h、引数で上書き可

#### テスト：正常系

- URL が返り、期限パラメータまたは SDK 出力が期待 TTL を反映する
- storage key 省略時は defaultKey（S3）を使う

#### テスト: 異常系

- local driver の key に対する PresignGet は拒否する
- 存在しない storage key は拒否する
- TTL 0 以下は拒否する（またはデフォルトへフォールバックを実装固定しテスト一致）

---


## 対象外（本版）

- SVG サニタイズ、PDF サムネ（`v0.4.0`）
- 署名付き PUT / DELETE URL
- ストレージ間の自動移行
- MySQL / Postgres を CI 必須とする実 DB 結合（任意）
- プロセスクラッシュ後の補償再開

----

以上

