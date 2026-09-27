---
type: Index
title: specs
description: 現行バージョンに存在する機能の仕様正本（ドメイン切り）。
tags: [median, specs, index]
timestamp: 2026-09-27T05:00:00Z
---

# specs

現行バージョンに存在する機能の仕様正本を置く。**SemVer フォルダでは切らない**（OKF v0.1 / doc-rule）。  
テスト仕様は [`../tests/`](../tests/) と**同じドメイン切り**。

| ドメイン | 内容 |
| --- | --- |
| [core-api/](./core-api/) | 公開 API（Store / Delete / Get） |
| [storage/](./storage/) | Local FS・storage key・ファイル名・shardian |
| [db-media/](./db-media/) | メディアスキーマ・パス列・テーブル設定（Go: crudian / JS: MediaRepo） |
| [media-pipeline/](./media-pipeline/) | 画像圧縮・リサイズ・サムネイル |
| [s3-adapter/](./s3-adapter/) | S3 互換 Adapter・署名付き GET |
| [svg-sanitize/](./svg-sanitize/) | SVG サニタイズ |
| [pdf-thumbnail/](./pdf-thumbnail/) | PDF 先頭ページサムネ |
| [packages-js/](./packages-js/) | TypeScript（`packages/js`） |
| [e2e/](./e2e/) | E2E（RustFS + POSIX、Go/JS） |
| [go-module-path/](./go-module-path/) | Go モジュールパス整合 |

未実装は [plans/](../plans/)（現在は [unscheduled/](../plans/unscheduled/) のみ）。  
完了マイルストーンの版索引は [`_archived/history/`](../_archived/history/) を参照。

----

以上
