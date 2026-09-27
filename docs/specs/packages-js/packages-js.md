---
type: Spec
title: packages/js（TypeScript）
description: Go契約をTypeScriptへ移植した現行仕様。
tags: [median, specs, packages-js, js]
timestamp: 2026-09-27T05:30:00Z
---

# packages/js（TypeScript）

- **状態**: 出荷済（現行仕様）
- **導入**: `v0.5.0`（以降 DB スキーマは `v0.7.0` で追随。npm **`0.7.0`**）
- **前提**: Go 契約語彙（[core-api](../core-api/)〜[pdf-thumbnail](../pdf-thumbnail/)）
- **関連**: [README.md](../../README.md) / [tests/packages-js/](../../tests/packages-js/) / [db-media](../db-media/)

## 目的

Go で固めた契約語彙を TypeScript（Node.js 24+ / bun）へ移植し、`@b4moss/median` として提供する。

## 方針

| 項目 | 決定 |
| --- | --- |
| 配置 | `packages/js` |
| 公開名 | `@b4moss/median`（CI / publish-npm 既存） |
| 入口 | `createMedian(config)` → `store` / `delete` / `get` / `presignGet` |
| 受け入れ | Go 相当（Store/Delete/Get/PresignGet・Local/S3・DB・画像・SVG・PDF サムネ） |
| DB | **crudian 不使用**。自前 MediaRepo + better-sqlite3（実行時）。DDL 生成は多方言 |
| テスト | `npm test` / `npm run test:coverage`、入力は [tests/packages-js/](../../tests/packages-js/) |
| SVG | **svgo** |
| PDF | `pdfRenderer` 注入必須（サムネ時）。CI は偽レンダラ |
| 現行版 | `package.json` **`0.7.0`**（ルートタグ `v0.7.0`。Go の `packages/go/v*` とは独立） |

## レイアウト

- `src/core` — createMedian / store / delete / get / presignGet / config / helpers
- `src/storage` + `local` + `s3`
- `src/db` — MediaRepo / migrate / schema
- `src/pipeline` — process / sanitizeSvg / PDF thumbs
- `src/internal` — MIME / filename / shardian path

## やらぬこと（当面）

- exports 条件の過剰な最適化
- PHP 移植（[unscheduled](../../plans/unscheduled/packages-php.md)）
- 署名付き PUT/DELETE、ストレージ間移行
- 実行時 MySQL/Postgres MediaRepo（DDL 生成のみ）

----

以上
