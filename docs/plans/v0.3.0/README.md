---
type: Index
title: v0.3.0 索引 — Go DB 連携
description: v0.3.0 マイルストーンの受け入れと文書索引。
tags: [median, plans, v0.3.0, index]
timestamp: 2026-09-25T05:28:00Z
---

# v0.3.0 索引 — Go DB 連携

- **状態**: 仕様詳細
- **マイルストーン**: `v0.3.0`
- **前提**: `v0.2.0`（コア + Local FS）
- **関連**: [er.dbml](../../er.dbml) / [roadmap](../../roadmap.md)

## 受け入れ（短い）

- b4moss/crudian 経由でメタデータを永続化できる（MySQL / MariaDB / Postgres / SQLite）
- 独自スキーマを `migrations/`（goose・**1 定義を dialect で解釈**）で適用できる
- 重複抑止・物理削除カスケード・採番・actor 契約が満たされる
- TDD・CI（Go path）あり

## 文書

| 文書 | 内容 |
| --- | --- |
| [db-media.md](./db-media.md) | crudian・カラム・削除・原子性・採番 |

----

以上
