---
type: Hub
title: median roadmap
description: SemVerマイルストーン一覧と短い受け入れのハブ。
tags: [median, roadmap]
timestamp: 2026-09-27T03:07:27Z
---

# median roadmap

SemVer（`docs/charter/versioning-rule.md`）に従う。`v0.n.0` は正式リリース前のため、MINOR 更新でも破壊的変更を許容する。

詳細仕様の正本は出荷済みが `specs/`、未実装が `plans/`。本ファイルは版一覧と短い受け入れのハブに限る。

並行可能なドメインは同一マイルストーンに統合し、版内で並列実装してよい。

## マイルストーン

| 版 | 状態 | 短い受け入れ | 詳細 |
| --- | --- | --- | --- |
| `v0.1.0` | 出荷済 | Go スキャフォールドのみ | [specs/v0.1.0/](./specs/v0.1.0/) |
| `v0.2.0` | 出荷済 | Go コア API + Local FS（storage key / ファイル名 / shardian） | [specs/v0.2.0/](./specs/v0.2.0/) |
| `v0.3.0` | 出荷済 | Go DB・画像パイプライン・S3 互換 Adapter（署名付き GET） | [specs/v0.3.0/](./specs/v0.3.0/) |
| `v0.4.0` | 出荷済 | Go SVG サニタイズ + PDF 先頭ページサムネ（タグ `packages/go/v0.4.0`） | [specs/v0.4.0/](./specs/v0.4.0/) |
| `v0.5.0` | 出荷済 | TypeScript（`packages/js` / npm `@b4moss/median`、タグ `v0.5.0`） | [specs/v0.5.0/](./specs/v0.5.0/) |
| `v0.6.0` | 出荷済 | E2E（RustFS + POSIX、CRUD+DB・サムネ、Go/JS。タグ `v0.6.0`。通常 CI 外・手動 Workflow） | [specs/v0.6.0/](./specs/v0.6.0/) |
| `v0.7.0` | 計画中 | DB メディアスキーマ（パス列再編・テーブル名設定。hash 非 UNIQUE。マイグレーション連携は版外） | [plans/v0.7.0/](./plans/v0.7.0/) |

## 未割当（unscheduled）

| 項目 | 状態 | 詳細 |
| --- | --- | --- |
| PHP（`packages/php`）移植 | 意図スタブ | [plans/unscheduled/packages-php.md](./plans/unscheduled/packages-php.md) |

## 開発順

1. Go（`packages/go`）… `v0.1.0`〜`v0.4.0`（完了）
2. TypeScript（`packages/js`）… `v0.5.0`（完了）
3. E2E テスト… `v0.6.0`（完了）
4. DB メディアスキーマ… `v0.7.0`（次）
5. PHP（`packages/php`）… マイルストーン未割当

CI/CD は他の `-an` 系（shardian / crudian）と同じ振る舞いとし、docs のみの変更で冗長な CI を起動しない。

----

以上
