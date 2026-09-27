---
type: Spec
title: テスト仕様 v0.8.0
description: Go モジュールパス整合の TDD 入力。
tags: [median, tests, v0.8.0, go, module]
timestamp: 2026-09-27T04:15:00Z
---

# テスト仕様 v0.8.0

対象マイルストーン: `v0.8.0`（Go module path）  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/go-module-path/`](../../specs/go-module-path/)  
前提: [`./v0.7.0.md`](../db-media/db-media.md)（振る舞い・スキーマは維持）  
ロードマップ: [`../../roadmap.md`](../../roadmap.md)

## 共通前提

| 項目 | 値 |
| --- | --- |
| 対象 | Go `packages/go`（全パッケージ）と `e2e/go` の replace |
| パッケージ SemVer | Go **`0.8.0`**（タグ `packages/go/v0.8.0`）。npm 変更なし |
| Module path | `github.com/b4moss/median/packages/go` |

## 受け入れケース

### M1 — go.mod 宣言

- **Given** `packages/go/go.mod`
- **Then** 先頭行が `module github.com/b4moss/median/packages/go`
- **And** `VERSION` が `0.8.0`

### M2 — 内部 import

- **Given** `packages/go` 配下の `.go` ソース
- **Then** `github.com/b4moss/median/go/...` への参照が残っていない
- **And** `go test ./...` が成功する

### M3 — e2e replace

- **Given** `e2e/go/go.mod`
- **Then** `require` / `replace` が `github.com/b4moss/median/packages/go` を指す

### M4 — 公開解決（出荷後）

- **Given** タグ `packages/go/v0.8.0` が push され Publish Go が完了している
- **Then** `https://proxy.golang.org/github.com/b4moss/median/packages/go/@v/v0.8.0.info` が 200
- **And** `go get github.com/b4moss/median/packages/go@v0.8.0` が成功する

----

以上
