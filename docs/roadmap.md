---
type: Hub
title: median roadmap
description: SemVerマイルストーン一覧と短い受け入れのハブ。
tags: [median, roadmap]
timestamp: 2026-09-27T05:00:00Z
---

# median roadmap

SemVer（`docs/charter/versioning-rule.md`）に従う。`v0.n.0` は正式リリース前のため、MINOR 更新でも破壊的変更を許容する。

詳細仕様の正本は出荷済みが `specs/`（ドメイン切り）、未実装が `plans/`。本ファイルは版一覧と短い受け入れのハブに限る。

並行可能なドメインは同一マイルストーンに統合し、版内で並列実装してよい。

## マイルストーン

| 版 | 状態 | 短い受け入れ | 詳細（現行ドメイン） |
| --- | --- | --- | --- |
| `v0.1.0` | 出荷済 | Go スキャフォールドのみ（歴史: [`_archived/history/v0.1.0/`](./_archived/history/v0.1.0/)） | — |
| `v0.2.0` | 出荷済 | Go コア API + Local FS（storage key / ファイル名 / shardian） | [core-api](./specs/core-api/) / [storage](./specs/storage/) |
| `v0.3.0` | 出荷済 | Go DB・画像パイプライン・S3 互換 Adapter（署名付き GET） | [db-media](./specs/db-media/) / [media-pipeline](./specs/media-pipeline/) / [s3-adapter](./specs/s3-adapter/) |
| `v0.4.0` | 出荷済 | Go SVG サニタイズ + PDF 先頭ページサムネ（タグ `packages/go/v0.4.0`） | [svg-sanitize](./specs/svg-sanitize/) / [pdf-thumbnail](./specs/pdf-thumbnail/) |
| `v0.5.0` | 出荷済 | TypeScript（`packages/js` / npm `@b4moss/median`、タグ `v0.5.0`） | [packages-js](./specs/packages-js/) |
| `v0.6.0` | 出荷済 | E2E（RustFS + POSIX、CRUD+DB・サムネ、Go/JS。タグ `v0.6.0`。通常 CI 外・手動 Workflow） | [e2e](./specs/e2e/) |
| `v0.7.0` | 出荷済 | DB メディアスキーマ（`file_path` / `file_name` / `original_file_name`、テーブル名／カラムマップ。Go/JS `0.7.0`） | [db-media](./specs/db-media/)（パス列現行） |
| `v0.8.0` | 出荷済 | Go モジュールパスを `github.com/b4moss/median/packages/go` に揃え、proxy 解決可能にする（タグ `packages/go/v0.8.0`） | [go-module-path](./specs/go-module-path/) |
| `v0.9.0` | 出荷済 | DDL 生成 CLI（`median migrate dump`）。Go / npm `0.9.0` | [ddl-cli](./specs/ddl-cli/) |

旧版フォルダ索引は [`_archived/history/`](./_archived/history/) を参照。

## 未割当（unscheduled）

| 項目 | 状態 | 詳細 |
| --- | --- | --- |
| PHP（`packages/php`）移植 | 意図スタブ | [plans/unscheduled/packages-php.md](./plans/unscheduled/packages-php.md) |

## 開発順

1. Go（`packages/go`）… `v0.1.0`〜`v0.4.0`（完了）。DB スキーマ `v0.7.0`。モジュールパス `v0.8.0`。DDL CLI `v0.9.0`
2. TypeScript（`packages/js`）… `v0.5.0`（完了）。DB スキーマ `v0.7.0`。DDL CLI `v0.9.0`
3. E2E テスト… `v0.6.0`（完了）
4. DB メディアスキーマ… `v0.7.0`（完了）
5. Go モジュールパス修正… `v0.8.0`（完了）
6. DDL 生成 CLI… `v0.9.0`（完了）
7. PHP（`packages/php`）… マイルストーン未割当

CI/CD は他の `-an` 系（shardian / crudian）と同じ振る舞いとし、docs のみの変更で冗長な CI を起動しない。

----

以上
