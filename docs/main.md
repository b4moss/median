---
type: Hub
title: median
description: ファイル操作を抽象化するDDD向けライブラリの目的・スコープ・技術方針ハブ。
tags: [median, hub]
timestamp: 2026-09-25T05:28:00Z
---

# median

ファイル操作を抽象化する DDD 向けライブラリ。

## 目的

バイト列をストレージへ出し入れする Repository 相当のコアと、必要に応じて呼び出す加工・メタデータ連携（Service / サブモジュール）を、言語横断で同じ契約語彙で提供する。

## スコープ

### やること

- ファイルの保存（`Store`）・削除（`Delete`）・取得（`Get`）
- ローカル FS /（後続）S3 互換ストレージへの Adapter
- 任意のメディア DB 連携（b4moss/crudian）
- 画像の圧縮・リサイズ・サムネイル生成（版により範囲が異なる。詳細は plans）
- MIME allow/deny、サイズ上限、同時アップロード制限、shardian による階層 path

### やらないこと

- DB テーブルからのメタデータ条件検索（CRUD client の責務）
- 論理削除（当面。`Delete` は物理削除）
- 動画・音声の変換・圧縮・リサイズ
- マルウェア・ウイルス検知（MIME で弾くか、別ソリューションと組み合わせる）

## 技術方針

- 他の `-an` 系（shardian / crudian）と同様の開発方針を主軸とする
- 開発順は **Go → TypeScript（`packages/js`）→ PHP**
- CI/CD は `-an` 系と同じ振る舞い（path filter、docs のみでは冗長起動しない）
- 憲章（`docs/charter/`）に従う。独自例外は `docs/override-charter.md`
- 未実装の詳細仕様は `docs/plans/`、現行の振る舞いは `docs/specs/`（現時点では未実装のため specs は空）
- テスト仕様は実装版に入るとき `docs/tests/` へ（TDD）

## パッケージ配置

| パス | 内容 |
|------|------|
| `packages/go` | Go（開発順 1） |
| `packages/js` | TypeScript / bun・Node.js（開発順 2） |
| `packages/php` | PHP（開発順 3） |
| `migrations/` | スキーママイグレーション正本（goose・方言解釈。MySQL/MariaDB/Postgres/SQLite） |

## ランタイム

- bun / Node.js 24+
- Go 1.26+
- PHP 8.2+

## 索引

| 文書 | 内容 |
|------|------|
| [roadmap.md](./roadmap.md) | マイルストーン |
| [er.dbml](./er.dbml) | デフォルト media テーブルの論理 ER |
| [plans/README.md](./plans/README.md) | 未実装計画の索引 |
| [specs/](./specs/) | 現行仕様（未実装のためプレースホルダ） |
| [charter/](./charter/) | 憲章 |
| [override-charter.md](./override-charter.md) | 憲章オーバーライド（現状なし） |

----

以上
