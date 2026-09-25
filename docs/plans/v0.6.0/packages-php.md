---
type: Plan
title: packages/php
description: 契約語彙をPHPへ移植する意図スタブ。
tags: [median, plans, v0.6.0, php]
timestamp: 2026-09-25T06:30:00Z
---

# packages/php

- **状態**: 意図スタブ
- **マイルストーン**: `v0.6.0`
- **前提**: Go / JS の契約語彙が先行していること
- **関連**: [main.md](../../main.md) / [v0.5.0/](../v0.5.0/)

## 目的

Go / JS で固めた契約語彙を PHP（Composer）へ移植する。

## ざっくり範囲

- 配置: `packages/php`
- 開発順は Go → JS の後
- マイグレーションは言語に適切なツールを用いる（正本はリポジトリ直下 `migrations/`）

## 後日詳細

- 最低 PHP バージョン以外の依存（画像処理ライブラリ等）
- 公開パッケージ名

----

以上
