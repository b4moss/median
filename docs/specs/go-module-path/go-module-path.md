---
type: Spec
title: Go module path alignment
description: module path を packages/go ディレクトリ／タグ接頭辞に揃え、proxy で解決可能にする。
tags: [median, specs, v0.8.0, go, module]
timestamp: 2026-09-27T04:15:00Z
---

# Go module path alignment（v0.8.0）

- **状態**: 出荷済（`VERSION` `0.8.0`、タグ `packages/go/v0.8.0`、module path 整合済み）
- **関連**: [tests/go-module-path/](../../tests/go-module-path/)

## 背景

`packages/go/go.mod` が `module github.com/b4moss/median/go` を宣言しつつ、ソースは `packages/go/`、タグは `packages/go/vX.Y.Z` だった。  
Go のネストモジュール規約ではパス末尾・ディレクトリ・タグ接頭辞が一致する必要があり、`proxy.golang.org` は `go/vX.Y.Z` を探して 404 になっていた（0.1〜0.7 共通）。

## 決定

| 項目 | 値 |
| --- | --- |
| Module path | `github.com/b4moss/median/packages/go` |
| ソース | `packages/go/`（変更なし） |
| タグ | `packages/go/v0.8.0`（変更なしの規約） |
| SemVer | Go **`0.8.0`**（npm は据え置き `0.7.0`） |
| 破壊的変更 | import パスが変わる（旧パスは proxy 未解決のため実害は限定的） |

## 非目標

- JS / マイグレーション / 公開 API の振る舞い変更
- crudian 側の同様のパス問題の修正

## 公開

1. `packages/go/VERSION` = `0.8.0`
2. タグ `packages/go/v0.8.0`
3. `publish-go.yml` が proxy を ping（Release 既存でも skip しない）

----

以上
