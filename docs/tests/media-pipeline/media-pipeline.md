---
type: Spec
title: テスト仕様 — media-pipeline
description: 画像圧縮・リサイズ・サムネイルの TDD 入力。
tags: [median, tests, media-pipeline, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — media-pipeline

対象ドメイン: `media-pipeline`  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/media-pipeline/`](../../specs/media-pipeline/)  
関連: [`../db-media/`](../db-media/) / [`../../roadmap.md`](../../roadmap.md)  
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


> **WebP エンコード**: Go は std にエンコーダが無く、加工後は **PNG バイト**を返す場合がある（宣言 MIME が webp のまま残りうる）。

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

### GIF（先頭フレーム）

- GIF は先頭フレームのみデコード・加工する。アニメ再生維持・APNG は対象外

#### テスト：正常系

- GIF をリサイズしても画像としてデコードできる（先頭フレーム）

#### テスト: 異常系

- 壊れた GIF 入力は拒否する

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

