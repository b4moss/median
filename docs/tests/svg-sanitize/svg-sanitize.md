---
type: Spec
title: テスト仕様 — svg-sanitize
description: SVG サニタイズの TDD 入力。
tags: [median, tests, svg-sanitize, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — svg-sanitize

対象ドメイン: `svg-sanitize`  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/svg-sanitize/`](../../specs/svg-sanitize/)  
関連: [`../pdf-thumbnail/`](../pdf-thumbnail/) / [`../media-pipeline/`](../media-pipeline/)  
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


## pipeline（SVG）

### SanitizeSVG

- SVG バイト列から危険な script / ハンドラ / 外部参照を除去し、安全な SVG バイトを返す
- 基本図形と `style` は残す

#### テスト：正常系

- 単純な `<svg><rect …/></svg>` は意味を保ったまま通る（要素・主要属性が残る）
- `<style>` 付きの図形は `style` が残る
- 危険要素を含む入力でも、許可された図形部分は残る

#### テスト: 異常系

- `<script>` 要素が出力に含まれない
- `onclick` 等の `on*` 属性が出力に含まれない
- `javascript:` や http(s) の外部 `href` / `xlink:href` が除去または無効化される
- 壊れた XML（パース不能）はエラーを返す

---


## core（SVG 接続）

### Store（`image/svg+xml`）

- MIME 検査後に必ずサニタイズし、サニタイズ後を本体として保存する
- ハッシュはサニタイズ後バイト

#### テスト：正常系

- 危険な script 付き SVG を Store すると、保存ファイルに `<script>` が無い
- 返却 `hash` / `size` がサニタイズ後バイトと一致する
- 基本図形のみの SVG は成功し、`mime` が `image/svg+xml` のまま

#### テスト: 異常系

- パース不能 SVG は Store 全体が失敗し、ストレージにファイルを残さない
- deny リストに `image/svg+xml` がある場合は従来どおり MIME 拒否（サニタイズ前）

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

