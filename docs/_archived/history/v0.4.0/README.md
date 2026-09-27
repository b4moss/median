---
type: Index
title: v0.4.0 索引 — SVG・PDF
description: v0.4.0 マイルストーンの受け入れと文書索引。
tags: [median, plans, v0.4.0, index]
timestamp: 2026-09-25T06:30:00Z
---

# v0.4.0 索引 — SVG・PDF

> 歴史資料。現行の仕様・テスト正本は [`../../../specs/`](../../../specs/) / [`../../../tests/`](../../../tests/)（ドメイン切り）。

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.4.0`
- **前提**: `v0.2.0`。パイプライン接続は `v0.3.0` 以降が自然
- **関連**: [roadmap](../../../roadmap.md) / [v0.3.0/media-pipeline.md](../media-pipeline/media-pipeline.md)

互いに独立な拡張（SVG サニタイズ / PDF 先頭ページサムネ）を同一版に統合する。版内は並列してよい。

## 受け入れ（短い）

- 危険な script / handler / 外部参照が SVG から除去され、基本図形・style は残る
- PDF アップロード時、先頭 1 ページのサムネイル生成を選択できる（既存サムネ契約へ接続）
- TDD・CI（Go path）あり

## ドメイン分割

| 文書 | 内容 |
| --- | --- |
| [svg-sanitize.md](../../../specs/svg-sanitize/svg-sanitize.md) | SVG サニタイズ |
| [pdf-thumbnail.md](../../../specs/pdf-thumbnail/pdf-thumbnail.md) | PDF 先頭ページサムネ |

----

以上
