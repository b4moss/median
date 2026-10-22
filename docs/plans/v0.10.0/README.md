---
type: Index
title: v0.10.0 — GetWithAllThumbnails
description: v0.10.0 マイルストーン索引（関連サムネイルを含めてファミリー全体を返す）。
tags: [median, plans, v0.10.0, core-api, index]
timestamp: 2026-10-22T09:34:00Z
---

# v0.10.0 — GetWithAllThumbnails

- **状態**: 仕様詳細（未実装・PO 承認待ち）
- **マイルストーン**: `v0.10.0`（GitHub: `med-v0.10.0`）
- **Issue**: [#77](https://github.com/b4moss/median/issues/77)
- **前提**: 現行 Go / npm **`0.9.0`**（[core-api](../../specs/core-api/) / [db-media](../../specs/db-media/) / [packages-js](../../specs/packages-js/)）
- **詳細計画**: [get-with-all-thumbnails.md](./get-with-all-thumbnails.md)
- **受け入れ（短い）**: DB ありで `GetWithAllThumbnails` / `getWithAllThumbnails` が、渡した id（オリジナルでも派生でも）から親をたどり、active なオリジナル＋variants（幅昇順）を返す。DB なしはエラー。Go / npm `0.10.0`

## 文書

| 文書 | 内容 |
| --- | --- |
| [get-with-all-thumbnails.md](./get-with-all-thumbnails.md) | 対象範囲・API・DB クエリ・TDD 観点・リリース・作業順・未決 |

## 含めない（Issue 決定）

- `getWithAllThmbnailsAsHTMLSrcSet` 等、srcset / img タグを HTML で返すメソッド（サービス層の仕事）

----

以上
