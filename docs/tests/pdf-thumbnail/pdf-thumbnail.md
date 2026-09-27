---
type: Spec
title: テスト仕様 — pdf-thumbnail
description: PDF 先頭ページサムネの TDD 入力。
tags: [median, tests, pdf-thumbnail, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — pdf-thumbnail

対象ドメイン: `pdf-thumbnail`  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/pdf-thumbnail/`](../../specs/pdf-thumbnail/)  
関連: [`../svg-sanitize/`](../svg-sanitize/) / [`../media-pipeline/`](../media-pipeline/)  
書き方: charter [`tdd.md`](../../charter/tdd.md)

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/packages/go` |
| 対象パッケージ | `pipeline`（SVG / PDF）/ `core`（Store 配線） |
| ランタイム / テスト | Go `1.26.x` + `go test` |
| PDF レンダラ | `PDFRenderer` は **Config 注入必須**（サムネ要求時）。Go に CGO 付き `FitzPDFRenderer` はあるが New はデフォルト注入しない。単体・結合は **偽レンダラ**で足りる |
| 入口 | `core.New(Config)` → `Store` / 既存 Get・Delete |
| 依存追加 | 任意で go-fitz（CGO・`FitzPDFRenderer`）。テストはインターフェース越しに偽実装 |

### 実装固定デフォルト

| 項目 | 値 |
| --- | --- |
| SVG MIME | `image/svg+xml`。Store 時は **常に** サニタイズしてから保存・ハッシュ |
| SVG 手法 | `encoding/xml` ベースの許可リスト。外部依存なし |
| 除去対象 | `script`、イベント属性（`on*`）、外部参照（`javascript:`・http(s) の `href` / `xlink:href` 等） |
| 残すもの | 基本図形（`svg` / `g` / `path` / `rect` / `circle` / `ellipse` / `line` / `polyline` / `polygon` / `text` 等）と `style` |
| PDF MIME | `application/pdf` |
| PDF オプトイン | `StoreOptions.PDFThumbnail`（bool）。`true` のときだけ先頭 1 ページをラスタ化し、既存 `Thumbnails` / `ThumbnailKeys` へ接続 |
| PDF 本体 | 原本 PDF をそのまま保存・ハッシュ。サムネは **variant**（既存契約: key・サイズ・圧縮・`original_id`） |
| ハッシュ | SVG: **サニタイズ後**バイトの SHA-256。PDF: **原本**バイトの SHA-256 |
| 処理順 | MIME 検査 →（SVG なら）サニタイズ →（画像なら）既存 pipeline →（PDF+オプトインなら）先頭ページ→サムネ → 保存 |

v0.2.0 / v0.3.0 の Local FS・MIME・サイズ・DB・画像パイプライン・S3 契約は回帰として壊さない（本版で触る経路のスモークでよい）。

Store 返却は v0.3 と同じ形。SVG は通常 `variants` なし（サムネ指定しても SVG 自体はラスタ pipeline 対象外）。PDF + `PDFThumbnail` 時は既存と同様に `variants` を返しうる。

---


## pipeline（PDF）

### PDFRenderer（インターフェース）

- PDF バイトとページ番号（本版は 0 = 先頭）からラスタ画像を返す
- Store は Config に注入された実装を使う。未注入かつ `PDFThumbnail=true` なら拒否する

#### テスト：正常系

- 偽レンダラが先頭ページ呼び出しで固定 `image.Image`（または PNG バイト）を返す
- ページ番号 0 が渡される

#### テスト: 異常系

- レンダラがエラーを返したら呼び出し側へ伝播する
- （本番 go-fitz パスは任意・CGO 環境でのスモークに限定してよい。CI 必須にしない場合はその旨を実装で固定）

### ThumbnailsFromPDFPage（または同等）

- 先頭ページ画像に既存サムネプリセット（`Thumbnails` / keys）を適用し、`[]pipeline.Variant` を返す
- サイズ・圧縮・key 名は v0.3 の画像サムネと同じ契約

#### テスト：正常系

- 指定 key の variant が返り、各バイトが空でない
- 未知でない preset の mode/size に従った寸法メタが載る（分かる場合）

#### テスト: 異常系

- 未知の thumbnail key はエラー
- 入力画像が不正ならエラー

---


## core（PDF 接続）

### Store（`application/pdf` + `PDFThumbnail`）

- `PDFThumbnail=false`（省略時）: PDF をそのまま保存。サムネは作らない
- `PDFThumbnail=true`: 注入 `PDFRenderer` で先頭ページを取り、既存 Thumbnails 契約で variants を保存
- 非画像への `Resize` 単独指定は従来どおり拒否。サムネは本フラグ経由のみ許可

#### テスト：正常系

- `PDFThumbnail` 省略/false で PDF を Store → 本体のみ保存、`variants` 空（または無し）
- `PDFThumbnail=true` + ThumbnailKeys（または defaultKeys）で、偽レンダラ経由の variants が返り、各 path のファイルが存在する
- DB ありのとき、親行は PDF、子行は `original_id` / `variant_key` 付き（v0.3 と同じ）
- 返却 `hash` は原本 PDF と一致する

#### テスト: 異常系

- `PDFThumbnail=true` だが `PDFRenderer` 未注入はエラー（ファイルを残さない）
- レンダ失敗時は Store 全体失敗・補償（部分書き込みなし）
- `PDFThumbnail=false` なのに `ThumbnailKeys` のみ付けた PDF は拒否（`ErrInvalidImage` 相当で一貫）
- 未知の ThumbnailKeys は Store 全体失敗・補償

---


## 回帰スモーク（本版で触る経路）

#### テスト：正常系

- JPEG 等の既存画像 Store（圧縮・サムネ）が壊れていない
- 非 SVG・非 PDF の通常バイト Store が壊れていない

---


## 対象外（本版）

- 完全な SVG 仕様準拠検証、マルウェア検知全般
- PDF 全ページ変換、テキスト抽出、PDF 編集
- 暗号化 PDF・巨大ファイルの特別扱い
- 署名付き PUT / DELETE URL
- JS（`v0.5.0`）移植 / PHP（[unscheduled](../../plans/unscheduled/packages-php.md)）
- go-fitz 実バイナリを CI 必須とする結合（任意。偽レンダラで契約を固定）

----

以上

