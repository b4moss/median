---
type: Hub
title: median roadmap
description: SemVerマイルストーン一覧と短い受け入れのハブ。
tags: [median, roadmap]
timestamp: 2026-09-25T05:28:00Z
---

# median roadmap

SemVer（`docs/charter/versioning-rule.md`）に従う。`v0.n.0` は正式リリース前のため、MINOR 更新でも破壊的変更を許容する。

詳細仕様の正本は各 `plans/`（実装後は `specs/`）。本ファイルは版一覧と短い受け入れのハブに限る。

## マイルストーン

| 版 | 状態 | 短い受け入れ | 詳細 |
| --- | --- | --- | --- |
| `v0.1.0` | 計画中 | Go スキャフォールドのみ（ディレクトリ＋`.gitkeep`） | [plans/v0.1.0/](./plans/v0.1.0/) |
| `v0.2.0` | 計画中 | Go コア API + Local FS（storage key / ファイル名 / shardian）。DB・画像加工なしでも Store/Delete/Get が動く | [plans/v0.2.0/](./plans/v0.2.0/) |
| `v0.3.0` | 計画中 | Go DB 連携（crudian・migrations・カスケード・重複抑止・採番） | [plans/v0.3.0/](./plans/v0.3.0/) |
| `v0.4.0` | 計画中 | Go 画像パイプライン（圧縮・リサイズ・サムネ。対象フォーマット込み） | [plans/v0.4.0/](./plans/v0.4.0/) |
| `v0.5.0` | 計画中 | S3 互換 Adapter + 署名付き GET URL | [plans/v0.5.0/](./plans/v0.5.0/) |
| `v0.6.0` | 計画中 | SVG サニタイズ | [plans/v0.6.0/](./plans/v0.6.0/) |
| `v0.7.0` | 計画中 | PDF 先頭ページサムネ | [plans/v0.7.0/](./plans/v0.7.0/) |
| `v0.8.0` | 計画中 | TypeScript（`packages/js`）移植 | [plans/v0.8.0/](./plans/v0.8.0/) |
| `v0.9.0` | 計画中 | PHP（`packages/php`）移植 | [plans/v0.9.0/](./plans/v0.9.0/) |

## 開発順

1. Go（`packages/go`）… `v0.1.0`〜`v0.7.0`
2. TypeScript（`packages/js`）… `v0.8.0`
3. PHP（`packages/php`）… `v0.9.0`

CI/CD は他の `-an` 系（shardian / crudian）と同じ振る舞いとし、docs のみの変更で冗長な CI を起動しない。

----

以上
