---
type: Spec
title: コア API（Store / Delete / Get / PresignGet）
description: 公開API・入出力・Get/Delete/PresignGet の仕様詳細。
tags: [median, specs, core-api, api]
timestamp: 2026-09-27T05:30:00Z
---

# コア API（Store / Delete / Get / PresignGet）

- **状態**: 出荷済（現行仕様）
- **導入**: `v0.2.0`（`PresignGet` は S3 と同時期）
- **関連**: [storage.md](../storage/storage.md) / [db-media.md](../db-media/db-media.md) / [media-pipeline.md](../media-pipeline/media-pipeline.md) / [s3-adapter.md](../s3-adapter/s3-adapter.md)

## 公開語彙

- 正式（Go）: `Store` / `Delete` / `Get` / `PresignGet`
- 正式（JS）: `store` / `delete` / `get` / `presignGet`（`createMedian` が返す）
- `Upload` という別名 API は**提供しない**
- 入力ヘルパ（任意）: Go `DecodeBase64` / `PartToStoreOptions`、JS `decodeBase64` / `partToStoreInput`  
  （HTTP Request 丸ごとの multipart パースは提供しない。抽出済み part を渡す）

## Store の入出力

### 入力

- 主経路は **バイト列**（Go: `io.Reader` + size。実装は指定 size を一括読込。真のストリーミング保存ではない）
- base64 / data URL・multipart part は上記ヘルパ経由で `Store` に渡す
- オプション（代表）: mime / filename / storage key / ファイル名モード / actor / `RejectDuplicate` / 画像パイプライン系（`Compress`・`KeepOriginal`・`Quality`・`Resize`・`ThumbnailKeys`） / `PDFThumbnail`

### 返却

```text
{ id?, path, mime, size, width?, height?, hash?, storageKey, variants? }
```

- DB 非使用時は `id` を持たない場合がある
- `width` / `height` は寸法が分かるとき（画像等）。不明なら省略または null
- `variants` は派生（サムネ等）を生成した場合のみ。各要素はサムネイル **key** で区別する（[media-pipeline.md](../media-pipeline/media-pipeline.md)）
- `storageKey` は実際に使用したストレージ key（必須で含める）
- **hash**: Store 返却の `hash` は**保存した主バイナリ**の SHA-256。DB の重複抑止用 hash は加工前入力バイト基準（圧縮・リサイズ後に両者は異なりうる）

## Delete / Get の識別

| モード | 正キー | 備考 |
| --- | --- | --- |
| DB なし | `path` | path で削除・取得 |
| DB あり | `id` | path は派生情報。メタ条件検索はスコープ外 |

ストレージ実体の解決は `path` とは別に **storage key** で行う（[storage.md](../storage/storage.md)）。key は DB に保存しない。

## Get の返却範囲

- デフォルト: メタデータのみ
- オプション: バイナリを同梱可（Go `GetOptions.WithBody` / JS `withBody`。返却はバイト列）

## PresignGet

- S3 互換 Adapter のみ。Local では拒否
- 詳細は [s3-adapter.md](../s3-adapter/s3-adapter.md)

## MIME

- 判定は呼び出し側の**宣言 MIME**のみ
- allow / deny は config の配列で設定

## 同時アップロード

- 同時アップロード数制限（デフォルト 20）。Median インスタンスあたりの in-flight `Store` 数

----

以上
