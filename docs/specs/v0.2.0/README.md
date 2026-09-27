---
type: Index
title: v0.2.0 索引 — Go コア API + Local FS
description: v0.2.0 マイルストーンの受け入れと文書索引。
tags: [median, plans, v0.2.0, index]
timestamp: 2026-09-25T06:30:00Z
---

# v0.2.0 索引 — Go コア API + Local FS

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.2.0`
- **関連**: [roadmap](../../roadmap.md) / 後続 [v0.3.0](../v0.3.0/)

## 受け入れ（短い）

Go で次が TDD（`docs/tests/`）および CI（Go path）付きで動くこと。

- `Store` / `Delete` / `Get`
- Local FS Adapter
- storage key・ファイル名モード・shardian

**含めない**: DB、画像加工、S3、SVG、PDF（以降の版）

## ドメイン分割

| 文書 | 内容 |
| --- | --- |
| [core-api.md](./core-api.md) | 公開 API・入出力・Get/Delete |
| [storage.md](./storage.md) | ストレージ key・ファイル名・path |

----

以上
