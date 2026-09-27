---
type: Spec
title: テスト仕様 v0.5.0
description: TypeScript（packages/js）移植の TDD 入力。Go v0.2〜v0.4 契約の同等振る舞い。
tags: [median, tests, v0.5.0, js, typescript]
timestamp: 2026-09-27T00:40:00Z
---

# テスト仕様 v0.5.0

対象マイルストーン: `v0.5.0`（TypeScript / `packages/js` 移植）  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/packages-js/`](../../specs/packages-js/)（[packages-js.md](../../specs/packages-js/packages-js.md)）  
Go 契約（参照）: [`./v0.2.0.md`](../core-api/core-api.md) / [`./v0.3.0.md`](../db-media/db-media.md) / [`./v0.4.0.md`](../svg-sanitize/svg-sanitize.md)  
ロードマップ: [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`tdd.md`](../../charter/tdd.md)（氷山パターン）

> **現行 DB 列・テーブル設定の受け入れは [`./v0.7.0.md`](../db-media/db-media.md) を正とする。** 本ファイルの DB 必須カラム記述は v0.5.0 出荷時点の契約として残す。

Go で固めた語彙・振る舞いを TypeScript で同等に満たす。細部の期待値は上記 Go テスト仕様と食い違わないこと。本ファイルは JS 実装・CI（`npm test`）の入力とする。

## 共通前提

| 項目 | 値 |
| --- | --- |
| パッケージ名 | `@b4moss/median` |
| 配置 | `packages/js` |
| 対象モジュール | `core` / `storage`（local・s3）/ `db` / `pipeline` / internal 相当 |
| ランタイム | Node.js **24+**（bun でも開発可） |
| テスト | `npm test`（カバレッジは `npm run test:coverage`）。CI は [`.github/workflows/ci.yml`](../../../.github/workflows/ci.yml) |
| 型チェック | `npx tsc -p tsconfig.build.json --noEmit` |
| テスト DB | **SQLite**（実ファイルまたは `:memory:`）。他 dialect は SQL 生成の単体まで |
| S3 テスト | Adapter 単体 + HTTP モック（実 MinIO は必須にしない） |
| 入口 | `createMedian(config)` または同等 → `store` / `delete` / `get` / `presignGet` |
| 依存方針 | shardian（JS）、crudian（JS）、画像処理、AWS SDK v3（S3）、**svgo**（SVG）、PDF は注入レンダラ |

### 実装固定デフォルト（Go と同値）

| 項目 | 値 |
| --- | --- |
| 同時アップロード | インスタンスあたり in-flight `store` 上限。デフォルト **20** |
| サイズ上限 | デフォルト **20 MiB** |
| MIME | 宣言 MIME のみ。`allow` 非空なら許可リスト、`deny` は常に拒否 |
| `hash` | SHA-256 hex |
| ファイル名 | デフォルト `random`（hex **32** 文字） |
| shardian | `dirLetterCount` / `dirNestDepth` **必須**（未設定は create 失敗） |
| ID 戦略 | デフォルト `auto_increment` |
| 重複 hash | デフォルト **既存返却**。`rejectDuplicate: true` で拒否 |
| actor | DB 使用時必須（options または `defaultActor`） |
| Delete/Get（DB あり） | 正キーは **`id`** |
| contain 余白 | デフォルト **black**。`white` / `transparent` / `#RRGGBB` / `#RRGGBBAA` |
| Presign GET TTL | デフォルト **1 時間** |
| SVG | `image/svg+xml` は **常に** サニタイズ後を保存・ハッシュ（手法は **svgo**） |
| PDF | `pdfThumbnail: true` のときのみ先頭ページ→既存サムネ契約。偽レンダラで CI 可 |

Store 返却: `{ id?, path, mime, size, hash, storageKey, width?, height?, variants? }`（DB/パイプライン無し時は id・寸法・variants 省略可）

---

## Config / createMedian

### createMedian（New 相当）

- Config を検証しインスタンスを生成する

#### テスト：正常系

- 有効な Config（defaultKey・shardian・local ルート）でインスタンスが得られる
- MIME / maxSize / maxConcurrentStores / filename 省略時に固定デフォルトが使われる
- 複数 storage key、SQLite DB 注入、S3 key、Thumbnails presets を登録できる

#### テスト: 異常系

- defaultKey 欠落・storages 空・shardian 未設定は拒否
- local root 空 / 未知 driver / S3 bucket 欠落は拒否
- Thumbnails の defaultKeys が presets に無いとき拒否

---

## storage（local）

### put / get / delete / exists

- ストレージルート配下の相対 path で出し入れする（親ディレクトリ自動作成）

#### テスト：正常系

- put した内容が get で一致する。ネスト path・上書き可
- delete 後 exists が false。未作成 path の exists は false

#### テスト: 異常系

- `..` / 絶対 path / 空 path は拒否
- 存在しない get は not found 相当
- 存在しない delete は not found または冪等成功で一貫

---

## ファイル名 / path（shardian）

### resolveFilename / buildStoragePath

#### テスト：正常系

- preserve: Unicode 残存、危険文字除去、先頭末尾 `.` / 空白除去
- random: hex（デフォルト長 32）+ 元拡張子
- 既知の幅・深さで shardian 期待 path と一致（決定的）

#### テスト: 異常系

- サニタイズ結果が空・長さ 0 以下・空ファイル名は拒否
- path に `..` が残らない

---

## MIME / サイズ / 同時アップロード

#### テスト：正常系

- allow/deny 空なら任意 MIME。上限未満サイズ・空ファイル（size 0）を通す
- 上限未満の並列 store は成功

#### テスト: 異常系

- deny / allow 外 MIME、サイズ超過、負サイズは拒否（put されない）
- 上限中の追加 store は待機。待機中キャンセルでスロットを漏らさない

---

## core（DB なし）

### store / delete / get

#### テスト：正常系

- `Uint8Array` / `Buffer` / stream 相当で `{ path, mime, size, hash, storageKey }` が返り実体がある
- hash が入力 SHA-256 と一致。未指定 key は defaultKey
- get（メタ / withBody）・delete が path+key で動く

#### テスト: 異常系

- MIME/サイズ拒否・未知 key・空 path・put 失敗時は成功メタを返さない

### 入力ヘルパ

#### テスト：正常系

- 素の base64 / data URL からバイトと MIME
- multipart part（filename + MIME + 本文）を store 入力に載せられる

#### テスト: 異常系

- 不正 base64、本文無し part、MIME 無し part は拒否（filename は random で補える）

---

## db / migrations

### migrateUp / migrateDown / MediaRepo

#### テスト：正常系

- SQLite Up で `media` と必須カラムが揃う。Down で消える  
  - **v0.7.0 以降の現行列は `file_path` / `file_name` / `original_file_name`（`path` 廃止）。受け入れは [`./v0.7.0.md`](../db-media/db-media.md)**
- `auto_increment` で整数 PK、text 系戦略で text PK
- Create / FindByID / FindByHash / ListChildren / DeleteByID（親削除で子行も消える）

#### テスト: 異常系

- 未対応 dialect・未知 ID 戦略・必須欠落・not found

---

## pipeline（画像）

### resize / compress / thumbnails

#### テスト：正常系

- longEdge / shortEdge / fixed（cover・contain・stretch）
- contain 余白: black デフォルト、white、transparent（非 JPEG）、HEX
- compress ON で再エンコード。defaultKeys / 明示 ThumbnailKeys で variants（key・バイト・width/height）
- アニメ GIF リサイズ後もデコード可能（最小）

#### テスト: 異常系

- 非画像・壊れた入力、fit 未指定、サイズ 0 以下、JPEG+transparent、不正 HEX、未知 thumbnail key

---

## storage/s3 / presignGet

#### テスト：正常系

- モック上で put/get/delete/exists。presignGet が TTL を反映（省略時 defaultKey・デフォルト 1h）

#### テスト: 異常系

- path 不正、not found、5xx 伝播
- local key への presignGet・未知 key は拒否

---

## core（DB・パイプライン接続）

### store（DB あり）

#### テスト：正常系

- id 返却、actor が created_by/owned_by に同値、同一 hash は既存返却、rejectDuplicate で拒否
- 圧縮・リサイズ・defaultKeys variants がファイルとして存在。width/height が載る（分かるフォーマット）
- variants は子行（`original_id` / `variant_key`）

#### テスト: 異常系

- actor 無しはエラーかつストレージ補償
- DB 挿入失敗・未知 ThumbnailKeys は部分書き込みを残さない
- 非画像への resize は拒否

### get / delete（DB あり・id）

#### テスト：正常系

- id で get（withBody 可）。親 delete で子ファイル・子行も消える

#### テスト: 異常系

- 存在しない id、DB 有効時の id 無し get/delete は拒否

---

## pipeline / core（SVG）

### sanitizeSvg（svgo）

#### テスト：正常系

- 基本図形・`style` が残る。危険要素混在でも許可部分は残る

#### テスト: 異常系

- `<script>` / `on*` / `javascript:`・http(s) 外部参照が出力に残らない
- 壊れた入力はエラー

### store（`image/svg+xml`）

#### テスト：正常系

- 保存ファイルに script 無し。hash/size はサニタイズ後と一致

#### テスト: 異常系

- パース不能はファイルを残さない。MIME deny はサニタイズ前に拒否

---

## pipeline / core（PDF）

### PDFRenderer / thumbnailsFromImage

#### テスト：正常系

- 偽レンダラに page **0** が渡る。指定 key の PNG 相当 variants が空でない

#### テスト: 異常系

- レンダエラー伝播、未知 key、不正画像

### store（`application/pdf` + `pdfThumbnail`）

#### テスト：正常系

- フラグ省略/false: 本体のみ、hash は原本
- true + keys/defaultKeys: variants が返りファイル存在。DB ありなら親子行

#### テスト: 異常系

- true かつレンダラ未注入は拒否（ファイル残さず）
- レンダ失敗・未知 keys は補償
- false なのに ThumbnailKeys のみは拒否（無効画像相当で一貫）

---

## 回帰スモーク

#### テスト：正常系

- 非画像バイトの store/get/delete（DB なし）
- JPEG/PNG の圧縮・サムネ
- SVG / PDF 経路を通したあとでも通常画像が壊れていない

---

## 対象外（本版）

- 公開 npm exports 条件の最終最適化（ビルドが通ればよい）
- 署名付き PUT / DELETE URL
- ストレージ間の自動移行
- MySQL / Postgres を CI 必須とする実 DB 結合（任意）
- PDF 全ページ・テキスト抽出・編集、暗号化・巨大 PDF の特別扱い
- 本番 PDF レンダラ実バイナリを CI 必須にする結合（偽レンダラで契約固定）
- PHP（[unscheduled](../../plans/unscheduled/packages-php.md)）
- Go パッケージ自体の変更

----

以上
