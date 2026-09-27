---
type: Plan
title: E2E tests
description: Docker 上の RustFS（S3 互換）・POSIX LocalFS に対する Go/JS E2E。通常 CI 外・手動 Workflow あり。
tags: [median, plans, v0.6.0, e2e]
timestamp: 2026-09-27T02:00:00Z
---

# E2E tests

- **状態**: 方針確定（シナリオ・ランナー詳細は後日）
- **マイルストーン**: `v0.6.0`
- **前提**: Go（`v0.4.0`）および TypeScript（`v0.5.0`）が出荷済みであること
- **関連**: [main.md](../../main.md) / [roadmap.md](../../roadmap.md) / [charter/tdd.md](../../charter/tdd.md)
- **Issue**: [#40](https://github.com/b4moss/median/issues/40)

## 目的

モックやホスト一時ディレクトリだけに頼らず、**本番に近いストレージ実体**上で median の Store / Get / Delete（S3 時は PresignGet 含む）が通ることを検証する。

## 方針

### ストレージ軸（必須・2 本）

| 軸 | 内容 |
| --- | --- |
| S3 互換 | Docker 上の **RustFS** に対し、実 PUT/GET/DELETE（必要なら署名付き GET）を実行する |
| Local FS | Docker 内の **POSIX ファイルシステム**上で Local Adapter を動かし、実ファイル I/O を検証する |

### S3 互換サーバ

- **採用: RustFS**（MinIO は採用しない。両者の並列メンテもしない）
- 理由: MinIO コミュニティ版は停滞・方針転換が進んでいる一方、RustFS は S3 互換・Apache 2.0・1.0 GA を前面に出している

### 言語パッケージ（必須・両方）

- **Go（`packages/go`）と TypeScript（`packages/js`）の両方**で、同一のクリティカルシナリオを通す
- 片言語のみの E2E では受け入れない

### CI / 実行場所

| 場所 | 扱い |
| --- | --- |
| 通常の PR / push CI（`ci.yml`） | **含めない**（必須ゲートにしない） |
| ローカル | **任意実行可能**。特に版上げ前は手元で通す運用 |
| GitHub Actions | **`workflow_dispatch` の手動 Workflow** のみ用意。必要なときだけ Actions から実行 |

## ざっくり範囲

- Docker Compose（または同等）で RustFS + POSIX ボリュームを起動
- Go / JS 双方のクリティカルシナリオ（詳細は後日）
- 手動 Workflow とローカル実行手順の文書化

## やらぬこと（当面）

- 通常 CI への常時組み込み
- MinIO の常設・併走
- 全 API 面の網羅的 E2E、ビジュアルリグレッション
- PHP（[unscheduled](../unscheduled/packages-php.md)）

## 後日詳細

- RustFS のイメージ pin・ヘルスチェック
- シナリオ一覧（Store / Get / Delete / Presign / パイプラインのどれを必須とするか）
- テストランナー配置（例: `e2e/`）と Go / JS の起動手順
- `docs/tests/v0.6.0.md` の TDD 入力

----

以上
