---
type: Plan
title: SVG サニタイズ
description: SVGから危険なscript等を除去する方針。
tags: [median, plans, v0.4.0, svg]
timestamp: 2026-09-25T06:30:00Z
---

# SVG サニタイズ

- **状態**: 方針確定
- **マイルストーン**: `v0.4.0`
- **前提**: `v0.2.0`。画像パイプライン接続は `v0.3.0` 以降が自然
- **関連**: [v0.3.0/media-pipeline.md](../v0.3.0/media-pipeline.md) / [pdf-thumbnail.md](./pdf-thumbnail.md)

## 目的

アップロードされる SVG から、よくある XSS / 外部参照系の脅威を除去する。

## 方針

- script / event handler / external 参照（xlink 等）を除去
- 基本図形・style は許可
- npm **svgo** のサニタイズ周りを参考にする
- TypeScript 実装（`v0.5.0`）では svgo をそのまま利用してよい
- Go / PHP は同等の挙動を目指す（ライブラリ選定は実装時）

## やらぬこと（当面）

- 完全な SVG 仕様準拠の検証
- マルウェア検知全般

## 受け入れ（短い）

- 危険な script / handler / 外部参照が除去される
- 基本図形・style は残る
- TDD・CI（Go path）あり

----

以上
