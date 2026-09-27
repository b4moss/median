---
type: Spec
title: packages/js（TypeScript）
description: Go契約をTypeScriptへ移植する方針。
tags: [median, specs, v0.5.0, js]
timestamp: 2026-09-27T00:40:00Z
---

# packages/js（TypeScript）

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.5.0`
- **前提**: Go `v0.2.0`〜`v0.4.0` 出荷済（契約の正本は Go テスト仕様）
- **関連**: [README.md](../../README.md) / [tests/packages-js/](../../tests/packages-js/) / [core-api](../core-api/)〜[svg-sanitize](../svg-sanitize/) / [pdf-thumbnail](../pdf-thumbnail/)

## 目的

Go で固めた契約語彙を TypeScript（Node.js 24+ / bun）へ移植し、`@b4moss/median` として提供する。

## 方針

| 項目 | 決定 |
| --- | --- |
| 配置 | `packages/js` |
| 公開名 | `@b4moss/median`（CI / publish-npm 既存） |
| 受け入れ | Go v0.2〜v0.4 相当（Store/Delete/Get/PresignGet・Local/S3・DB・画像・SVG・PDF サムネ） |
| テスト | `npm test` / `npm run test:coverage`、入力は [tests/packages-js/packages-js.md](../../tests/packages-js/packages-js.md) |
| SVG | **svgo** |
| PDF | `PDFRenderer` 注入。CI は偽レンダラ |
| 版上げ | `package.json` version + ルートタグ `v0.5.0`（Go の `packages/go/v*` とは独立） |

## レイアウト（実装時）

- `src/core` — createMedian / store / delete / get / presignGet / config
- `src/storage` + `local` + `s3`
- `src/db` — MediaRepo / migrate
- `src/pipeline` — process / sanitizeSvg / PDF thumbs
- `src/internal` — MIME / filename / shardian path

## やらぬこと（当面）

- exports 条件の過剰な最適化
- PHP 移植（[unscheduled](../../plans/unscheduled/packages-php.md)）
- 署名付き PUT/DELETE、ストレージ間移行

----

以上
