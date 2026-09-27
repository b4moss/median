---
type: Spec
title: コア API（Store / Delete / Get）
description: 公開API・入出力・Get/Delete の仕様詳細。
tags: [median, specs, v0.2.0, api]
timestamp: 2026-09-25T06:30:00Z
---

# コア API（Store / Delete / Get）

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.2.0`
- **関連**: [storage.md](../storage/storage.md) / [db-media.md](../db-media/db-media.md) / [media-pipeline.md](../media-pipeline/media-pipeline.md)

## 公開語彙

- 正式: `Store` / `Delete` / `Get`
- `Upload` はエイリアス、または multipart/base64 向け入力アダプタ

## Store の入出力

### 入力

- `bytes | stream` を主とする
- コアで `multipart` / `base64`（または data URL）も受け付ける
- HTTP Request 丸ごとの multipart パースは薄いヘルパとして任意提供（主経路は抽出済み part）
- オプション: mime / filename / storage key / ファイル名モード / actor 等

### 返却

```text
{ id?, path, mime, size, width?, height?, hash?, storageKey, variants? }
```

- DB 非使用時は `id` を持たない場合がある
- `width` / `height` は寸法が分かるとき（画像等）。不明なら省略または null
- `variants` は派生（サムネ等）を生成した場合のみ。各要素はサムネイル **key** で区別する（[media-pipeline.md](../media-pipeline/media-pipeline.md)）
- `storageKey` は実際に使用したストレージ key（必須で含める）

## Delete / Get の識別

| モード | 正キー | 備考 |
| --- | --- | --- |
| DB なし | `path` | path で削除・取得 |
| DB あり | `id` | path は派生情報。メタ条件検索はスコープ外 |

ストレージ実体の解決は `path` とは別に **storage key** で行う（[storage.md](../storage/storage.md)）。key は DB に保存しない。

## Get の返却範囲

- デフォルト: メタデータのみ
- オプション: バイナリ（bytes / stream）を同梱可

## MIME

- 判定は呼び出し側の**宣言 MIME**のみ
- allow / deny は config の配列で設定

## 同時アップロード

- 同時アップロード数制限（デフォルト 20）。単位は実装時に `docs/tests/` で固定

----

以上
