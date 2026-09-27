---
type: Index
title: tests
description: TDD入力となるテスト仕様（specs と同じドメイン切り）。
tags: [median, tests, index]
timestamp: 2026-09-27T05:00:00Z
---

# tests

TDD の入力となるテスト仕様を置く。書き方・氷山パターンは charter [`tdd.md`](../charter/tdd.md)。  
`specs/` と**同じドメイン切り**。**SemVer フォルダ（`vX.Y.Z`）では切らない**（OKF v0.1）。

| ドメイン | テスト仕様 |
| --- | --- |
| [scaffold/](./scaffold/) | スキャフォールドのみのため原則不要 |
| [core-api/](./core-api/) | Go コア API |
| [storage/](./storage/) | Local FS・key・ファイル名・shardian |
| [db-media/](./db-media/) | DB・メディアスキーマ（パス列含む） |
| [media-pipeline/](./media-pipeline/) | 画像パイプライン |
| [s3-adapter/](./s3-adapter/) | S3 互換 Adapter |
| [svg-sanitize/](./svg-sanitize/) | SVG サニタイズ |
| [pdf-thumbnail/](./pdf-thumbnail/) | PDF サムネ |
| [packages-js/](./packages-js/) | TypeScript / `packages/js` |
| [e2e/](./e2e/) | E2E（L1/S1・CRUD+DB・サムネ） |
| [go-module-path/](./go-module-path/) | Go モジュールパス整合 |

版ごとの歴史資料は [`_archived/history/`](../_archived/history/) を参照。

----

以上
