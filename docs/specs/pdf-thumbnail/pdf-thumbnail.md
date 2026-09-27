---
type: Spec
title: PDF サムネイル
description: PDF先頭ページサムネの仕様。
tags: [median, specs, pdf-thumbnail, pdf]
timestamp: 2026-09-27T05:30:00Z
---

# PDF サムネイル

- **状態**: 出荷済（現行仕様）
- **導入**: `v0.4.0`
- **前提**: [media-pipeline.md](../media-pipeline/media-pipeline.md)（サムネ契約）
- **関連**: [svg-sanitize.md](../svg-sanitize/svg-sanitize.md) / [tests/pdf-thumbnail/](../../tests/pdf-thumbnail/)

## 目的

PDF アップロード時、最初の 1 ページ目をサムネイルとして生成するかどうかを選択可能にする。

## 振る舞い

- オプトイン: Go `StoreOptions.PDFThumbnail` / JS `pdfThumbnail`
- 先頭ページ（page 0）をラスタ化し、既存サムネイル契約（`Thumbnails` / `ThumbnailKeys`・`original_id` / `variant_key`）へ接続
- PDF 本体は原本のまま保存・ハッシュ。サムネは variant
- **`PDFRenderer` は呼び出し側が Config に注入する**（未注入かつサムネ要求時はエラー）
  - Go には CGO ビルドタグ付き `FitzPDFRenderer`（go-fitz）があるが、`New` / `createMedian` はデフォルト注入しない
  - 単体・CI は偽レンダラで契約を固定してよい

## やらぬこと（当面）

- 全ページ変換、テキスト抽出、PDF 編集
- 暗号化 PDF・巨大ファイルの特別扱い

----

以上
