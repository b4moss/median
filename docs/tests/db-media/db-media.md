---
type: Spec
title: テスト仕様 — db-media
description: DB・メディアメタデータ連携の TDD 入力（パス列の現行は本ファイル＋v0.7 統合分）。
tags: [median, tests, db-media, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — db-media

対象ドメイン: `db-media`  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/db-media/`](../../specs/db-media/)  
関連: [`../media-pipeline/`](../media-pipeline/) / [`../s3-adapter/`](../s3-adapter/) / [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`tdd.md`](../../charter/tdd.md)

> パス列・テーブル設定の現行受け入れは本ファイル末尾の「パス列（現行）」を正とする（旧 `v0.7.0`）。

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/packages/go` |
| 対象パッケージ | `core` / `db` / `pipeline` / `storage` / `storage/local` / `storage/s3` |
| ランタイム / テスト | Go `1.26.x` + `go test` |
| テスト DB | **SQLite**（gorm + modernc 系）。他 dialect はマイグレーション SQL 生成の単体確認まで |
| S3 テスト | Adapter 単体 + `httptest` モック（実 MinIO は必須にしない） |
| 入口 | `core.New(Config)` → `Store` / `Delete` / `Get` / `PresignGet` |
| 依存 | Go: crudian（gorm）+ shardian + imaging + AWS SDK v2。JS 側は別ドメイン（packages-js） |

### 実装固定デフォルト

| 項目 | 値 |
| --- | --- |
| DB（Go） | crudian 注入優先（`*gorm.DB` / Crud）。未注入時は Config DSN から内部生成可 |
| ID 戦略 | デフォルト `auto_increment`（Config / `MEDIAN_ID_STRATEGY` で切替） |
| 重複 hash | デフォルト **既存レコード返却**。`RejectDuplicate=true` で拒否 |
| actor | DB 使用時必須（`StoreOptions.Actor` または Config `DefaultActor`） |
| Delete/Get（DB あり） | 正キーは **`id`** |
| contain 余白 | デフォルト **black**。`white` / `transparent` / `#RRGGBB` / `#RRGGBBAA` 可（詳細は Resize） |
| Presign GET TTL | デフォルト **1 時間**（呼び出しで上書き可） |

v0.2.0 の Local FS・MIME・サイズ・同時アップロード・ファイル名・shardian 契約は回帰として壊さない（本版で触る経路のスモークでよい）。

Store 返却（本版・DB/パイプライン有効時に増えうる）: `{ id?, path, mime, size, hash, storageKey, width?, height?, variants? }`

---


## migrations / db

### MigrateUp / MigrateDown

- goose で `media` テーブルを作成・削除する
- 1 定義を dialect で解釈する（方言別 SQL の並列メンテはしない）
- `MEDIAN_ID_STRATEGY` / Config の ID 戦略で PK 型が切り替わる

#### テスト：正常系

- SQLite に Up すると `media` が存在し、必須カラム（id/path/mime/size/width/height/hash/created_at/original_id/variant_key/created_by/owned_by/status）が揃う  
  - **v0.7.0 以降の現行列は `file_path` / `file_name` / `original_file_name`（`path` 廃止）。受け入れは [`./v0.7.0.md`](../db-media/db-media.md)**
- Down すると `media` が消える
- `auto_increment` 戦略で整数系 PK になる
- `uuid_v4`（または text 系戦略）で text PK の DDL が選ばれる（生成 SQL または適用結果で確認）

#### テスト: 異常系

- 未対応 dialect 指定は拒否する
- 未知の ID 戦略は拒否する
- Up 済みへの二重適用は goose の版管理に従い破綻しない（エラーまたは no-op で一貫）

### MediaRepo.Create

- メタデータ行を crudian 経由で挿入する
- Store 時 `created_by` と `owned_by` は同値

#### テスト：正常系

- 必須フィールドを渡すと行が返り、`id` が採番される
- actor 指定が両カラムに入る
- `original_id` / `variant_key` 付きの派生行を挿入できる

#### テスト: 異常系

- path 欠落など必須欠落は拒否する（**v0.7.0 以降は `file_path` / `file_name` / `hash` 等。[`./v0.7.0.md`](../db-media/db-media.md)**）
- actor 未指定かつ DefaultActor なしは拒否する（Store 経路）
- 一意制約（path、または original_id+variant_key）違反はエラーになる（**v0.7.0 以降の UNIQUE は `file_path`**）

### MediaRepo.FindByID / FindByHash

- ID または hash で行を取得する

#### テスト：正常系

- 存在する ID で行が取れる
- 同一 hash の行が取れる

#### テスト: 異常系

- 存在しない ID / hash は not found 相当
- 空 ID / 空 hash は拒否する

### MediaRepo.DeleteByID（カスケード）

- 親削除時、`original_id` 参照の子行も削除する（物理削除）

#### テスト：正常系

- 親のみなら親行が消える
- 親+子があるとき子行も消える
- 子だけの削除は親を残す

#### テスト: 異常系

- 存在しない ID は not found 相当
- 空 ID は拒否する

---


## core（DB 接続）

### New（DB / S3 Config）

- DB 注入または DSN、S3 storage 定義を受け付ける

#### テスト：正常系

- SQLite DB を注入した Config で New できる
- driver `s3` の storage key を登録できる
- Thumbnails presets / defaultKeys を持てる

#### テスト: 異常系

- S3 で bucket 欠落は拒否する
- Thumbnails の defaultKeys が presets に無いとき拒否する
- DB 有効なのに DefaultActor も無しで、Store 時 actor も無いと後段で失敗する（New 時検証してもよい）

### Store（DB あり）

- ストレージ保存後にメタを永続化し、返却に `id` を含める
- actor を created_by / owned_by に同値挿入
- 同一 hash はデフォルトで既存行を返す（ストレージへ二重書きしない）

#### テスト：正常系

- 新規 Store で id が返り、FindByID で同じ path/mime/size/hash が読める
- Actor または DefaultActor が両カラムに入る
- 同一内容の再 Store で既存 id が返り成功する
- RejectDuplicate=true の再 Store は拒否する

#### テスト: 異常系

- actor 未指定かつ DefaultActor なしはエラー（ファイルが残らないこと＝補償）
- DB 挿入失敗時は書いたストレージオブジェクトを削除する（補償）
- 未知の storage key は拒否する

### Get / Delete（DB あり・id）

- 正キーは id。path は派生情報

#### テスト：正常系

- Store した id で Get（メタ）でき、WithBody で実体が取れる
- Delete(id) で行とファイルが消える

#### テスト: 異常系

- 存在しない id は not found
- DB 有効時に id 無し Delete/Get は拒否する
- path のみの Delete は DB 有効時は使わない（拒否または未サポートで一貫）

---


## 対象外（本版）

- SVG サニタイズ、PDF サムネ（`v0.4.0`）
- 署名付き PUT / DELETE URL
- ストレージ間の自動移行
- MySQL / Postgres を CI 必須とする実 DB 結合（任意）
- プロセスクラッシュ後の補償再開

## パス列・テーブル設定（現行）

旧マイルストーン `v0.7.0` 由来。物理列 `path` は廃止。`file_path` / `file_name` / `original_file_name` が正。

## 共通前提

| 項目 | 値 |
| --- | --- |
| 対象 | Go `packages/go`（`db` / `core`）と JS `packages/js`（`src/db` / `src/core`）の**両方** |
| テスト DB | **SQLite**（Go: gorm + 既存系、JS: better-sqlite3）。他 dialect は DDL 生成の単体確認まで |
| 入口 | Go: `core.New` → `Store` / `Delete` / `Get`。JS: `createMedian` → `store` / `delete` / `get` |
| パッケージ SemVer | Go **`0.7.0`**（タグ `packages/go/v0.7.0`）/ npm **`0.7.0`**（ルートタグ `v0.7.0`） |
| 破壊的変更 | 物理列 `path` を廃止。既存 DB への ALTER マイグレーションは提供しない（`00001` 置き換え） |

### 実装固定デフォルト

| 項目 | 値 |
| --- | --- |
| デフォルトテーブル名 | `media` |
| パス系物理列 | `file_path` / `file_name` / `original_file_name` |
| `file_path` | UNIQUE。Adapter の Put/Get/Delete キー（現行 `path` の役割） |
| `file_name` | `ResolveFilename` 後の basename。`path.Base(file_path)` と一致 |
| `original_file_name` | オリジナル行: Store 入力の filename（空可）。派生行: 空または NULL |
| `hash` | SHA-256 hex。**UNIQUE にしない**（INDEX のみ）。重複抑止はアプリ層 |
| カラムマップ未指定 | デフォルト物理名を使用 |
| 公開 API | `StoreResult.path` 等は**リネームしない**（値は `file_path` と一致） |
| 重複 hash | デフォルト既存レコード返却。`RejectDuplicate=true` で拒否（現行契約維持） |
| actor / ID 戦略 | v0.3 / v0.5 と同じ |

本版で触らない経路（Local/S3 Adapter・画像パイプライン・SVG/PDF・Presign）は回帰スモークで壊さないこと。

Store 返却（公開）: `{ id?, path, mime, size, hash, storageKey, width?, height?, variants? }` — `path` はストレージ相対パス。

---

## migrations / DDL

### CreateMediaSQL / MigrateUp / MigrateDown

- 1 定義を dialect で解釈する（方言別 SQL の並列メンテはしない）
- 物理正本は `migrations/`。Go / JS の DDL 生成は同内容に同期する
- **テーブル名**を引数（またはオプション）で指定できる。未指定は `media`
- DDL 内の列名は常に**デフォルト物理名**（カラムマップは Repo 専用。Migrate の列リネームはしない）
- `MEDIAN_ID_STRATEGY` / Config の ID 戦略で PK 型が切り替わる（現行どおり）

#### テスト：正常系

- SQLite に Up すると指定テーブル（デフォルト `media`）が存在し、必須カラムが揃う:
  - `id` / `file_path` / `file_name` / `original_file_name` / `mime` / `size` / `width` / `height` / `hash` / `created_at` / `original_id` / `variant_key` / `created_by` / `owned_by` / `status`
- 旧列名 `path` は**存在しない**
- `file_path` に UNIQUE 制約がある
- `hash` に UNIQUE 制約が**無い**（INDEX のみ、または制約一覧に含まれない）
- Down すると当該テーブルが消える
- テーブル名を `medias` 等に指定して Up/Down できる（デフォルト名のテーブルは作られない）
- `auto_increment` で整数系 PK、`uuid_v4`（または text 系）で text PK の DDL が選ばれる

#### テスト: 異常系

- 未対応 dialect 指定は拒否する
- 未知の ID 戦略は拒否する

---

## MediaRepo

### Create

- メタデータ行を挿入する（Go: crudian、JS: 既存 SQLite 実装）
- 必須: `file_path` / `file_name` / `hash` / `mime` / actor 系（現行どおり）
- `original_file_name` は空を許容する

#### テスト：正常系

- 必須フィールドを渡すと行が返り、`id` が採番される
- `file_path` / `file_name` / `original_file_name` が永続化され、読み戻せる
- actor 指定が `created_by` / `owned_by` 両カラムに入る
- `original_id` / `variant_key` 付きの派生行を挿入できる（派生の `original_file_name` は空または NULL でよい）

#### テスト: 異常系

- `file_path` / `file_name` / `hash` / `mime` 欠落は拒否する
- `file_path` 一意制約違反はエラーになる
- `original_id` + `variant_key` 一意制約違反はエラーになる

### FindByID / FindByHash / ListChildren / DeleteByID

- 現行契約を維持（hash 検索はオリジナル優先 → フォールバック）
- 親削除で子行も物理削除

#### テスト：正常系

- 存在する ID / hash で行が取れ、パス系3列が読める
- 同一 `file_path` 値を持つ行の hash 重複抑止用 FindByHash が動く
- ListChildren / DeleteByID カスケードが現行どおり

#### テスト: 異常系

- 存在しない ID / hash は not found 相当
- 空 ID / 空 hash は拒否する

### テーブル名 / カラムマップ

- Config（または Repo コンストラクタ）でテーブル名と論理名→物理列名マップを渡せる
- マップ未指定キーはデフォルト物理名
- マップは少なくともパス系（`file_path` / `file_name` / `original_file_name`）および現行必須一式をカバーできる

#### テスト：正常系

- 別名テーブル（例: `medias`）+ デフォルト列名で Create / FindByID できる
- 物理列を別名にしたテーブル（例: `file_path` → `storage_key`）へマップを渡すと CRUD できる
- マップ一部省略時、省略分はデフォルト名でアクセスする

#### テスト: 異常系

- マップが指す列が実テーブルに無い場合、操作は失敗する（実装のエラー種別に従う）

---

## core Store / Delete / Get（DB 有効時）

### Store

- ストレージへ Put した相対パスを DB の `file_path` に書く
- `file_name` は `ResolveFilename` 結果（basename）
- `original_file_name` は入力 `filename`（preserve / random とも、入力値を保持。空入力は空）
- 公開 `StoreResult.path` は `file_path` と同じ値（フィールド名は `path` のまま）
- 派生行: 各 variant の `file_path` / `file_name` を格納。`original_file_name` は空または NULL
- hash 重複・RejectDuplicate・actor 必須は現行どおり

#### テスト：正常系

- filename 付き Store 後、DB 行の `original_file_name` が入力と一致する
- `file_name` がストレージ path の basename と一致する
- `StoreResult.path` === DB `file_path` === Adapter 上のオブジェクトキー
- random モードでも `original_file_name` に入力 filename が残る（入力が空なら空）
- サムネ等の派生行が親 `id` を `original_id` に持ち、パス系列が埋まる
- 同一内容の再 Store はデフォルトで既存 id を返し、行が増えない
- `RejectDuplicate=true` で拒否する

#### テスト: 異常系

- DB 有効かつ actor 未指定・DefaultActor なしは拒否する
- 必須欠落・MIME/サイズ拒否は現行どおり（成功メタと孤立ファイルを残さない）

### Delete / Get（DB あり）

- 正キーは **`id`**
- Adapter 操作は行の `file_path` を使う

#### テスト：正常系

- id 指定 Delete で親・子のストレージ実体と DB 行が消える
- Get（メタ / withBody）が `file_path` 経由で読める

#### テスト: 異常系

- id 未指定・not found は現行どおり

---

## E2E 追随（v0.6.0 シナリオ）

v0.6.0 のシナリオ ID（L1/S1・C-L/C-S・T-L/T-S）は維持する。変更点は DB メタ参照のみ。

#### テスト：正常系

- CRUD+DB シナリオで行のストレージキー照合が `file_path`（または Repo が返す同等フィールド）で成立する
- 生 SQL を使う場合は `path` 列を参照しない
- サムネ子行の path 照合も同様

本マイルストーンで E2E ハーネス新規要件は追加しない（Compose / Workflow は v0.6.0 契約のまま）。

---

## 版外（本ファイルで検証しない）

- マイグレーションコマンド / ホスト migrate 連携（DDL 出力 CLI・goose embed 等）
- `hash` の UNIQUE 化・部分 UNIQUE
- 公開 API フィールド名 `path` のリネーム
- 既存 DB からの自動 ALTER アップグレード

----

以上


----

以上

