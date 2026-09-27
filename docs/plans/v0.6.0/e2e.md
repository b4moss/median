---
type: Plan
title: E2E tests
description: Docker 上の RustFS・POSIX LocalFS に対する Go/JS E2E（CRUD+DB メタ・サムネ）。通常 CI 外・手動 Workflow。
tags: [median, plans, v0.6.0, e2e]
timestamp: 2026-09-27T02:25:00Z
---

# E2E tests

- **状態**: 方針確定（シナリオは [tests/v0.6.0.md](../../tests/v0.6.0.md) 第2部）
- **マイルストーン**: `v0.6.0`
- **前提**: Go（`v0.4.0`）および TypeScript（`v0.5.0`）が出荷済みであること
- **関連**: [main.md](../../main.md) / [roadmap.md](../../roadmap.md) / [charter/tdd.md](../../charter/tdd.md) / [tests/v0.6.0.md](../../tests/v0.6.0.md)
- **Issue**: [#40](https://github.com/b4moss/median/issues/40)

## 目的

本番に近いストレージ実体（RustFS / Docker POSIX）上で、median のクリティカル経路が通ることを検証する。

## 方針

### ストレージ軸（必須・2 本）

| 軸 | 内容 |
| --- | --- |
| S3 互換 | Docker 上の **RustFS** |
| Local FS | Docker 内の **POSIX** Local Adapter |

### S3 互換サーバ

- **採用: RustFS**（MinIO 不使用）

### 言語パッケージ（必須・両方）

- **Go / JS 両方**で同一シナリオ

### DB

- E2E では **インメモリ SQLite** でメタ整合を確認する（crudian 自体は信頼）

### シナリオ範囲（必須）

| ID | 内容 |
| --- | --- |
| L1 / S1 | no-DB のファイル往復（S1 は PresignGet 含む） |
| C-L / C-S | Local / RustFS + DB の **CRUD**（ファイルとメタ。U は内容置換 Delete→Store と同一 hash 再 Store） |
| T-L / T-S | Local / RustFS + DB の **サムネイル**（variants 実体・DB 親子・親削除カスケード） |

### CI / 実行場所

| 場所 | 扱い |
| --- | --- |
| 通常の PR / push CI（`ci.yml`） | **含めない** |
| ローカル | **任意実行可能** |
| GitHub Actions | **`workflow_dispatch` のみ** |

## やらぬこと（当面）

- 通常 CI への常時組み込み
- MinIO、実 MySQL/Postgres、SVG/PDF E2E、網羅的 API
- PHP（[unscheduled](../unscheduled/packages-php.md)）

## 後日詳細（実装時）

- RustFS イメージ pin・ヘルスチェック
- `e2e/` ランナー配置と起動手順

----

以上
