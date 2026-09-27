---
type: Plan
title: E2E tests
description: エンドツーエンドテストを追加する意図スタブ。
tags: [median, plans, v0.6.0, e2e]
timestamp: 2026-09-27T01:45:00Z
---

# E2E tests

- **状態**: 意図スタブ
- **マイルストーン**: `v0.6.0`
- **前提**: Go（`v0.4.0`）および TypeScript（`v0.5.0`）が出荷済みであること
- **関連**: [main.md](../../main.md) / [roadmap.md](../../roadmap.md) / [charter/tdd.md](../../charter/tdd.md)
- **Issue**: [#40](https://github.com/b4moss/median/issues/40)

## 目的

出荷済みの Go / JS 契約に対し、クリティカルな利用シナリオをカバーするエンドツーエンド（E2E）テストを追加する。

## ざっくり範囲

- リポジトリ全体または言語パッケージ横断の E2E 層（配置・ランナーは後日詳細）
- MVP として必要なシナリオに絞る（charter の氷山パターン・E2E 方針に沿う）

## やらぬこと（当面）

- 全 API 面の網羅的 E2E
- ビジュアルリグレッション

## 後日詳細

- ツール選定（ランナー、フィクスチャ、CI 上の起動条件）
- シナリオ一覧（Store / Get / Delete / Presign / パイプライン等のどれを必須とするか）
- Go のみ / JS のみ / 両方を走らせるか

----

以上
