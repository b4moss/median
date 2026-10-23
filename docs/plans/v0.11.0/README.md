---
type: Index
title: v0.11.0 — GetWithAllThumbnails
description: v0.11.0 マイルストーン索引（関連サムネイルを含めてファミリー全体を返す）。
tags: [median, plans, v0.11.0, core-api, index]
timestamp: 2026-10-22T10:54:00Z
---

# v0.11.0 — GetWithAllThumbnails

- **状態**: 仕様詳細（未実装・PO 承認待ち。U1〜U12 決定済み）
- **マイルストーン**: `v0.11.0`（GitHub: `med-v0.11.0`）
- **Issue**: [#77](https://github.com/b4moss/median/issues/77)
- **前提**: 現行 Go / npm **`0.9.0`**（[core-api](../../specs/core-api/) / [db-media](../../specs/db-media/) / [packages-js](../../specs/packages-js/)）
- **詳細計画**: [get-with-all-thumbnails.md](./get-with-all-thumbnails.md)
- **受け入れ（短い）**: DB ありで `GetWithAllThumbnails` / `getWithAllThumbnails` が `StoreResult` を返す（派生 id でも親をたどり、active なオリジナル＋variants。無しは空配列。幅昇順・同幅は key 昇順）。DB なしは `ErrDBRequired`。Go / npm `0.11.0`

## 文書

| 文書 | 内容 |
| --- | --- |
| [get-with-all-thumbnails.md](./get-with-all-thumbnails.md) | 対象範囲・API・DB クエリ・TDD 観点・リリース・作業順・**決定事項**・要検討事項 |

## 決定事項（要約）

詳細は [get-with-all-thumbnails.md §11](./get-with-all-thumbnails.md#11-決定事項)。

- 戻り値は既存 `StoreResult`。サムネ無しは `variants: []`
- メタのみ（WithBody なし）。storageKey は Get と同じく任意
- inactive 派生 id でも親をたどる。絞り込みは core。親判定は `original_id` のみ
- 並び: 幅昇順（0 含む）→ 同幅は `variant_key` 昇順 → 幅 NULL のみ末尾
- DB なし: Go・JS とも `ErrDBRequired`（既存の `ErrIDRequired` 等と同じ番兵エラー）
- 既存 `Get` は本版で変えない（要検討事項に残す）
- 単体厚め＋Local E2E 任意 1 本

## 含めない（Issue 決定）

- `getWithAllThmbnailsAsHTMLSrcSet` 等、srcset / img タグを HTML で返すメソッド（サービス層の仕事）

----

以上
