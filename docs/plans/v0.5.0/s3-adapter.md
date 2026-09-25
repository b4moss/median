---
type: Plan
title: S3 互換 Adapter
description: S3互換ストレージと署名付きGET URLの方針。
tags: [median, plans, v0.5.0, s3]
timestamp: 2026-09-25T05:28:00Z
---

# S3 互換 Adapter

- **状態**: 方針確定
- **マイルストーン**: `v0.5.0`
- **前提**: `v0.2.0`（storage key モデル）
- **関連**: [v0.2.0/storage.md](../v0.2.0/storage.md)

## 目的

S3 互換ストレージへ出し入れし、閲覧用の署名付き URL を返す。

## 方針

- SDK を用いた S3 互換 Adapter
- マルチストレージは既存の **storage key** モデル（config の `driver` / `bucket` / `path`）
- 署名付き URL は **GET のみ**
- 有効期限は config で定める（呼び出しで上書き可）
  - ライブラリが用意する config デフォルト値は **1時間**

## やらぬこと（当面）

- 署名付き PUT / DELETE URL
- ストレージ間の自動移行ツール（Adapter 差し替え可能性の維持は別）

## 受け入れ（短い）

- S3 互換へ Store / Delete / Get できる
- 署名付き GET URL を発行できる（期限は config / 上書き）
- TDD・CI（Go path）あり

----

以上
