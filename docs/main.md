---
type: Hub
title: median
description: ファイル操作を抽象化するDDD向けライブラリの目的・スコープ・技術方針ハブ。
tags: [median, hub]
timestamp: 2026-09-27T01:40:00Z
---

# median

ファイル操作を抽象化する DDD 向けライブラリ。

## 目的

バイト列をストレージへ出し入れする Repository 相当のコアと、必要に応じて呼び出す加工・メタデータ連携（Service / サブモジュール）を、言語横断で同じ契約語彙で提供する。

## スコープ

### やること

- ファイルの保存（`Store`）・削除（`Delete`）・取得（`Get`）・署名付き GET（`PresignGet`）
- ローカル FS / S3 互換ストレージへの Adapter
- 任意のメディア DB 連携（b4moss/crudian）
- 画像の圧縮・リサイズ・サムネイル、SVG サニタイズ、PDF 先頭ページサムネ（版により範囲が異なる。詳細は specs）
- MIME allow/deny、サイズ上限、同時アップロード制限、shardian による階層 path

### やらないこと

- DB テーブルからのメタデータ条件検索（CRUD client の責務）
- 論理削除（当面。`Delete` は物理削除）
- 動画・音声の変換・圧縮・リサイズ
- マルウェア・ウイルス検知（MIME で弾くか、別ソリューションと組み合わせる）

## 技術方針

- 他の `-an` 系（shardian / crudian）と同様の開発方針を主軸とする
- 開発順は **Go → TypeScript（`packages/js`）→ E2E →（未割当）PHP**
- CI/CD は `-an` 系と同じ振る舞い（path filter、docs のみでは冗長起動しない）
- 憲章（`docs/charter/`）に従う。独自例外は `docs/override-charter.md`
- **現行仕様**は `docs/specs/`、これからやる内容は `docs/plans/`（未割当は `plans/unscheduled/`）
- テスト仕様は `docs/tests/`（TDD）。E2E は `v0.6.0`（[`e2e/`](../e2e/)、通常 CI 外）

## パッケージ配置

| パス | 内容 | 現状 |
|------|------|------|
| `packages/go` | Go（開発順 1） | 出荷済 `v0.4.0`（タグ `packages/go/v0.4.0`） |
| `packages/js` | TypeScript / Node.js 24+（開発順 2） | 出荷済 `v0.5.0`（npm `@b4moss/median`） |
| `packages/php` | PHP（開発順: 未割当） | 未着手（[unscheduled](./plans/unscheduled/packages-php.md)） |
| `migrations/` | スキーママイグレーション正本（goose・方言解釈） | 共用 |

## ランタイム

- Node.js 24+（bun でも開発可）
- Go 1.26+
- PHP 8.2+（予定）

## 索引

| 文書 | 内容 |
|------|------|
| [roadmap.md](./roadmap.md) | マイルストーン |
| [er.dbml](./er.dbml) | デフォルト media テーブルの論理 ER |
| [specs/](./specs/) | 現行仕様（v0.1〜v0.6） |
| [plans/](./plans/) | 未実装計画（unscheduled） |
| [tests/](./tests/) | テスト仕様 |
| [charter/](./charter/) | 憲章 |
| [override-charter.md](./override-charter.md) | 憲章オーバーライド（現状なし） |
| [`e2e/`](../e2e/) | E2E ハーネス（v0.6.0） |

----

以上
