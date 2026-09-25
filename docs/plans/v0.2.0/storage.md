---
type: Plan
title: ストレージ・パス・ファイル名
description: Local FS・storage key・ファイル名・shardian の仕様詳細。
tags: [median, plans, v0.2.0, storage]
timestamp: 2026-09-25T05:28:00Z
---

# ストレージ・パス・ファイル名

- **状態**: 仕様詳細
- **マイルストーン**: `v0.2.0`（Local FS）。S3 は `v0.5.0`
- **関連**: [core-api.md](./core-api.md) / [v0.5.0/s3-adapter.md](../v0.5.0/s3-adapter.md)

## Adapter（v0.2.0）

- Local FS（Linux / FreeBSD / Mac / Windows）
- Adapter 化し、後から S3 等へ移行可能な境界を維持する

## マルチストレージ（key 解決）

- config に複数ストレージを **key** 付きで登録する
  - 各 key に `driver`, `bucket` / `path`（ルート）等
- config に **default key** を置く
- `Store` / `Delete` / `Get` で未指定時は default key、呼び出しで上書き可
- DB が保持するのは **ストレージルートからの相対 path のみ**（key / bucket は持たない）
- どの key を渡すかの管理は client の責務
- 誤った key による失敗も client 責任
- Store 返却には実際に使った `storageKey` を含める

## ファイル名モード

| モード | 意味 |
| --- | --- |
| `preserve` | 元ファイル名を維持（サニタイズ適用） |
| `random` | ランダム文字列へ書き換え |

- config でデフォルトを定め、API 引数でオーバーライド可

### `random`

- 暗号論的ランダム hex（JS は `crypto` 相当。他言語も同等）
- 長さは config + オーバーライド引数で指定可
- 拡張子は元ファイルから付与

### `preserve` 時のサニタイズ

- Unicode は残す
- 危険文字（`/\\..\0` 等）と、先頭末尾の `.` / 空白のみ除去

## path / shardian

- 最終ファイル名に対し b4moss/shardian で階層化
- 幅・深さ等のパラメータは **config 必須**

----

以上
