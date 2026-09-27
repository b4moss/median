---
type: Spec
title: E2E tests
description: Docker 上の RustFS・POSIX LocalFS に対する Go/JS E2E（CRUD+DB メタ・サムネ）。通常 CI 外・手動 Workflow。
tags: [median, specs, v0.6.0, e2e]
timestamp: 2026-09-27T02:40:00Z
---

# E2E tests

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.6.0`
- **前提**: Go（`v0.4.0`）および TypeScript（`v0.5.0`）が出荷済みであること
- **関連**: [main.md](../../main.md) / [roadmap.md](../../roadmap.md) / [charter/tdd.md](../../charter/tdd.md) / [tests/v0.6.0.md](../../tests/v0.6.0.md)
- **実装**: [`e2e/`](../../../e2e/)
- **Issue**: [#40](https://github.com/b4moss/median/issues/40)

## 目的

本番に近いストレージ実体（RustFS / Docker 共有 POSIX）上で、median のクリティカル経路が通ることを検証する。

## 方針

### ストレージ軸（必須・2 本）

| 軸 | 内容 |
| --- | --- |
| S3 互換 | Docker 上の **RustFS**（イメージ pin: `rustfs/rustfs:1.0.0`） |
| Local FS | Compose の `localfs` と bind-mount した **POSIX** ディレクトリ上の Local Adapter |

### 言語パッケージ（必須・両方）

- **Go / JS 両方**で同一シナリオ（[`e2e/run.sh`](../../../e2e/run.sh)）

### DB

- E2E では **インメモリ SQLite** でメタ整合を確認する（crudian 自体は信頼）

### シナリオ範囲（必須）

| ID | 内容 |
| --- | --- |
| L1 / S1 | no-DB のファイル往復（S1 は PresignGet 含む） |
| C-L / C-S | Local / RustFS + DB の **CRUD**（U は同一 hash 再 Store と Delete→Store 置換） |
| T-L / T-S | Local / RustFS + DB の **サムネイル**（variants・DB 親子・カスケード） |

詳細手順は [tests/v0.6.0.md](../../tests/v0.6.0.md) 第2部。

### CI / 実行場所

| 場所 | 扱い |
| --- | --- |
| 通常の PR / push CI（`ci.yml`） | **含めない** |
| ローカル | **任意実行**（`./e2e/run.sh`）。版上げ前に推奨 |
| GitHub Actions | **`workflow_dispatch` のみ**（[`.github/workflows/e2e.yml`](../../../.github/workflows/e2e.yml)） |

## やらぬこと

- 通常 CI への常時組み込み、MinIO、実 MySQL/Postgres、SVG/PDF E2E、PHP

----

以上
