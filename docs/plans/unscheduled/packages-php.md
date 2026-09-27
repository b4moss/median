---
type: Plan
title: packages/php
description: 契約語彙をPHPへ移植する意図スタブ。
tags: [median, plans, unscheduled, php]
timestamp: 2026-09-27T01:45:00Z
---

# packages/php

- **状態**: 意図スタブ
- **マイルストーン**: `unscheduled`（旧 `v0.6.0` 候補。PHP は版未割当）
- **前提**: Go / JS の契約語彙が先行していること
- **関連**: [main.md](../../main.md) / [specs/v0.5.0/](../../specs/v0.5.0/) / [roadmap.md](../../roadmap.md)

## 目的

Go / JS で固めた契約語彙を PHP（Composer）へ移植する。

## ざっくり範囲

- 配置: `packages/php`
- 開発順は Go → JS の後（E2E `v0.6.0` の後でも可）
- マイグレーションは言語に適切なツールを用いる（正本はリポジトリ直下 `migrations/`）

## 後日詳細

- 最低 PHP バージョン以外の依存（画像処理ライブラリ等）
- 公開パッケージ名
- 割り当てる SemVer マイルストーン

----

以上
