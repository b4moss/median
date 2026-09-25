---
type: Hub
title: median roadmap
description: SemVerマイルストーン一覧と短い受け入れのハブ。
tags: [median, roadmap]
timestamp: 2026-09-25T06:30:00Z
---

# median roadmap

SemVer（`docs/charter/versioning-rule.md`）に従う。`v0.n.0` は正式リリース前のため、MINOR 更新でも破壊的変更を許容する。

詳細仕様の正本は各 `plans/`（実装後は `specs/`）。本ファイルは版一覧と短い受け入れのハブに限る。

並行可能なドメインは同一マイルストーンに統合し、版内で並列実装してよい。

## マイルストーン

| 版 | 状態 | 短い受け入れ | 詳細 |
| --- | --- | --- | --- |
| `v0.1.0` | 出荷済 | Go スキャフォールドのみ（ディレクトリ＋`.gitkeep`） | [plans/v0.1.0/](./plans/v0.1.0/) |
| `v0.2.0` | 計画中 | Go コア API + Local FS（storage key / ファイル名 / shardian）。DB・画像加工なしでも Store/Delete/Get が動く | [plans/v0.2.0/](./plans/v0.2.0/) |
| `v0.3.0` | 計画中 | Go DB（crudian）・画像パイプライン・S3 互換 Adapter（署名付き GET）。版内は並列可 | [plans/v0.3.0/](./plans/v0.3.0/) |
| `v0.4.0` | 計画中 | SVG サニタイズ + PDF 先頭ページサムネ。版内は並列可 | [plans/v0.4.0/](./plans/v0.4.0/) |
| `v0.5.0` | 計画中 | TypeScript（`packages/js`）移植 | [plans/v0.5.0/](./plans/v0.5.0/) |
| `v0.6.0` | 計画中 | PHP（`packages/php`）移植 | [plans/v0.6.0/](./plans/v0.6.0/) |

## 開発順

1. Go（`packages/go`）… `v0.1.0`〜`v0.4.0`
2. TypeScript（`packages/js`）… `v0.5.0`
3. PHP（`packages/php`）… `v0.6.0`

CI/CD は他の `-an` 系（shardian / crudian）と同じ振る舞いとし、docs のみの変更で冗長な CI を起動しない。

----

以上
