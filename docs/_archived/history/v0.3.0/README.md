---
type: Index
title: v0.3.0 索引 — Go DB・画像・S3
description: v0.3.0 マイルストーンの受け入れと文書索引。
tags: [median, plans, v0.3.0, index]
timestamp: 2026-09-25T06:30:00Z
---

# v0.3.0 索引 — Go DB・画像・S3

> 歴史資料。現行の仕様・テスト正本は [`../../../specs/`](../../../specs/) / [`../../../tests/`](../../../tests/)（ドメイン切り）。

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.3.0`
- **前提**: `v0.2.0`（コア + Local FS）
- **関連**: [er.dbml](../../../er.dbml) / [roadmap](../../../roadmap.md)

`v0.2.0` 完了後に並行可能なドメイン（DB / 画像パイプライン / S3）を同一版に統合する。実装はドメイン単位で並列してよい。DB 連携サムネ自己参照は DB とパイプラインの両方を要する。

## 受け入れ（短い）

- b4moss/crudian 経由でメタデータを永続化できる（MySQL / MariaDB / Postgres / SQLite）
- 独自スキーマを `migrations/`（goose・**1 定義を dialect で解釈**）で適用できる
- 重複抑止・物理削除カスケード・採番・actor 契約が満たされる
- 画像の圧縮・本体リサイズ・サムネイル生成がオプションとして動く（JPEG/PNG/WebP/GIF・アニメ GIF/APNG）
- S3 互換へ Store / Delete / Get でき、署名付き GET URL を発行できる
- TDD・CI（Go path）あり

**含めない**: SVG サニタイズ・PDF サムネ（`v0.4.0`）

## ドメイン分割

| 文書 | 内容 |
| --- | --- |
| [db-media.md](../../../specs/db-media/db-media.md) | crudian・カラム・削除・原子性・採番 |
| [media-pipeline.md](../../../specs/media-pipeline/media-pipeline.md) | 加工パイプライン・画像フォーマット |
| [s3-adapter.md](../../../specs/s3-adapter/s3-adapter.md) | S3 互換 Adapter・署名付き GET |

----

以上
