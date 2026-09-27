---
type: Index
title: v0.7.0 — DB media schema
description: v0.7.0 マイルストーン索引（パス列再編・テーブル名設定）。
tags: [median, plans, v0.7.0, db, schema, index]
timestamp: 2026-09-27T03:15:07Z
---

# v0.7.0 — DB media schema

- **状態**: 方針確定（計画中）
- **マイルストーン**: `v0.7.0`
- **文書**: [db-media-schema.md](./db-media-schema.md)
- **テスト仕様**: [tests/v0.7.0.md](../../tests/v0.7.0.md)
- **受け入れ（短い）**: `file_path` / `file_name` / `original_file_name` への再編。`hash` は非 UNIQUE。テーブル名（とカラムマップ）を設定可能に。Go / JS。マイグレーションホスト連携は版外
- **前提**: [specs/v0.3.0/db-media.md](../../specs/v0.3.0/db-media.md) / [er.dbml](../../er.dbml)

----

以上
