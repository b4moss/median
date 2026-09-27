---
type: Spec
title: テスト仕様 v0.3.0
description: Go DB・画像パイプライン・S3 の TDD 入力。
tags: [median, tests, v0.3.0, go]
timestamp: 2026-09-25T07:20:00Z
---

# テスト仕様 v0.3.0

対象マイルストーン: `v0.3.0`（Go DB・画像パイプライン・S3）  
製品: [`../main.md`](../../../README.md)  
仕様: [`../specs/db-media/`](../../../specs/db-media/)（[db-media.md](../../../specs/db-media/db-media.md) / [media-pipeline.md](../../../specs/media-pipeline/media-pipeline.md) / [s3-adapter.md](../../../specs/s3-adapter/s3-adapter.md)）  
前提（コア）: [`./v0.2.0.md`](../../../tests/core-api/core-api.md)  
ロードマップ: [`../roadmap.md`](../roadmap.md)  
書き方: charter [`tdd.md`](../../../charter/tdd.md)（氷山パターン）

> **現行 DB 列・テーブル設定の受け入れは [`./v0.7.0.md`](../../../specs/db-media/db-media.md) を正とする。** 本ファイルの `path` 必須カラム等は v0.3.0 出荷時点の契約として残す。

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/go` |
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

## migrations / db

### MigrateUp / MigrateDown

- goose で `media` テーブルを作成・削除する
- 1 定義を dialect で解釈する（方言別 SQL の並列メンテはしない）
- `MEDIAN_ID_STRATEGY` / Config の ID 戦略で PK 型が切り替わる

#### テスト：正常系

- SQLite に Up すると `media` が存在し、必須カラム（id/path/mime/size/width/height/hash/created_at/original_id/variant_key/created_by/owned_by/status）が揃う  
  - **v0.7.0 以降の現行列は `file_path` / `file_name` / `original_file_name`（`path` 廃止）。受け入れは [`./v0.7.0.md`](../../../specs/db-media/db-media.md)**
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

- path 欠落など必須欠落は拒否する（**v0.7.0 以降は `file_path` / `file_name` / `hash` 等。[`./v0.7.0.md`](../../../specs/db-media/db-media.md)**）
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

## pipeline

### Resize

- 本体の長辺 / 短辺 / 固定サイズ制約を適用する
- `fixed` のとき fit は `cover` / `contain` / `stretch`
- `contain` の余白塗りは `containBackground`: `black`（デフォルト） / `white` / `transparent` / `#RRGGBB` / `#RRGGBBAA`（大小文字不問）

#### テスト：正常系

- JPEG/PNG で longEdge 指定後、長辺が指定 px 以下（または一致）になる
- shortEdge 指定後、短辺が期待どおりになる
- fixed+cover で出力幅・高さが指定どおりになる
- fixed+contain（背景省略）で指定枠に収まり、余白が黒になる（サンプリングで確認してよい）
- containBackground=`white` で余白が白になる
- containBackground=`#RRGGBB` で余白がその RGB になる
- containBackground=`#RRGGBBAA` で余白の RGB と alpha が反映される（PNG/WebP 等）
- containBackground=`transparent` で PNG/WebP/GIF の余白が透明になる（GIF / indexed PNG は AA なしで可）

#### テスト: 異常系

- 非画像 MIME / 壊れたバイナリは拒否する
- fixed で fit 未指定は拒否する
- サイズ 0 以下の制約は拒否する
- JPEG + `transparent` は拒否する
- 不正な HEX（桁数・文字）は拒否する
- `transparent` と HEX を同時に渡す契約が無いこと（単一フィールドなら該当なし）

### Compress

- 圧縮オプション ON 時に品質を下げたバイナリを出す（フォーマット維持を基本）

#### テスト：正常系

- JPEG で quality を下げるとサイズが原寸未満になりうる（厳密なバイト減は必須としないが、再エンコードされる）
- 圧縮 OFF なら入力（リサイズ後）をそのまま次段へ渡せる

#### テスト: 異常系

- quality が 0〜1 の外なら拒否する

### Thumbnails

- config の presets / defaultKeys に従い派生画像を生成する
- 呼び出しの `ThumbnailKeys` で上書きできる

#### テスト：正常系

- defaultKeys のみ指定相当で、列挙 key の variants が生成される
- 明示 key（辞書にあって default 外）だけ生成できる
- 各 variant に key・バイト・width/height が付く

#### テスト: 異常系

- 辞書に無い key は拒否する
- presets 空なのに key 指定したら拒否する

### Animated GIF / APNG（最小）

- リサイズ後もデコーダで読める（再生可能な出力）

#### テスト：正常系

- アニメ GIF をリサイズしても gif としてデコードできる
- APNG（または対応デコード経路）をリサイズしても画像としてデコードできる

#### テスト: 異常系

- 壊れたアニメ入力は拒否する

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

## core（Pipeline 接続）

### Store（リサイズ・圧縮・サムネ、DB なし可）

- オプションに応じて本体を加工し、variants をストレージへ保存して返却する

#### テスト：正常系

- Compress ON かつ KeepOriginal 省略で、保存本体が圧縮後バイナリになる（原寸ファイルを残さない）
- 本体 longEdge 制約時は原寸を残さない
- defaultKeys の variants が返り、各 path のファイルが存在する
- width/height が本体・variant に載る（分かるフォーマット）

#### テスト: 異常系

- 未知の ThumbnailKeys は Store 全体が失敗し、部分書き込みを残さない（補償）
- 非画像に Resize を付けたら拒否する

---

## core（DB × Pipeline）

### Store variants 永続化

- 派生は同一テーブルの子行（`original_id` → 親、`variant_key` = プリセット key）

#### テスト：正常系

- Store 後、親行の original_id/variant_key が NULL、子行が key ごとに存在する
- 返却 variants[].id が子行と一致する（付与する場合）

#### テスト: 異常系

- 子の DB 挿入失敗時は親・子ファイルと親行を可能な範囲で巻き戻す

### Delete カスケード（ファイル + 行）

- 親 id 削除で子行と子ファイルも消える

#### テスト：正常系

- 親 Delete 後、子 path の Exists が false、子行も無い
- 親ファイルも消える

#### テスト: 異常系

- 子ファイル削除の一部失敗でも、可能な削除を試しエラーを返す（振る舞いを実装固定）

---

## 対象外（本版）

- SVG サニタイズ、PDF サムネ（`v0.4.0`）
- 署名付き PUT / DELETE URL
- ストレージ間の自動移行
- MySQL / Postgres を CI 必須とする実 DB 結合（任意）
- プロセスクラッシュ後の補償再開

----

以上
