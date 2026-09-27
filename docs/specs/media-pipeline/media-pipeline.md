---
type: Spec
title: メディア加工パイプライン
description: 圧縮・リサイズ・サムネイルプリセットの仕様詳細。
tags: [median, specs, v0.3.0, media]
timestamp: 2026-09-25T07:20:00Z
---

# メディア加工パイプライン

- **状態**: 出荷済（現行仕様）
- **マイルストーン**: `v0.3.0`
- **前提**: `v0.2.0`（必須）。DB 連携サムネ自己参照は同版の [db-media.md](../db-media/db-media.md) と併用
- **関連**: [core-api.md](../core-api/core-api.md) / [db-media.md](../db-media/db-media.md)

## 範囲（v0.3.0）

- 画像の圧縮（する / しない）
- 本体のリサイズ制約（長辺 / 短辺 / 絶対指定）
- サムネイル生成（名前付きプリセット、複数 key、個別圧縮率）

**含めない**: SVG サニタイズ（`v0.4.0`）、PDF 先頭ページサムネ（`v0.4.0`）

## 画像フォーマット（初回）

- JPEG / PNG / WebP / GIF
- アニメ GIF・APNG も対象
  - 方針: **リサイズのみ**行い、出力が**適切に再生できること**を確認すれば足りる
  - フレーム合成・遅延・ループ等のエッジケース追求は不要

## 圧縮とオリジナル

- 圧縮とサムネイルは**独立したオプション**
- 圧縮 ON 時のデフォルトは、アップロード先にオリジナルを保存しない
  - 「オリジナル保存」フラグで上書き可能
- 画像**本体**に長辺（または短辺・絶対サイズ）制約を付けた場合は、オリジナル保存を**明示的に否定した**とみなす
  - 制約適用後のバイナリが保存対象（原寸ファイルは書かない）

## サムネイル定義（config）

サムネイルは **名前付きプリセット** として config に持つ。1 プリセット = 1 **key**。

### プリセット辞書

各 key に対して次を定義する。

| 項目 | 内容 |
| --- | --- |
| **key**（名前） | プリセット識別子。返却 `variants` や DB 派生行の区別に使う |
| **基準サイズ (px)** | 次のいずれか |
| **fit**（トリミング方式） | 基準サイズが**縦横固定値**のとき必須 |
| **圧縮率** | `0`〜`1`（プリセットごと） |

#### 基準サイズのモード

| モード | 意味 | fit |
| --- | --- | --- |
| `longEdge` | 長辺を指定 px に合わせる（アスペクト比維持） | 不要 |
| `shortEdge` | 短辺を指定 px に合わせる（アスペクト比維持） | 不要 |
| `fixed` | 幅・高さの固定値（px） | **必須**: `cover` / `contain` / `stretch` |

- `cover` … 領域を埋め、はみ出しはトリミング
- `contain` … 全体が収まるようレターボックス。余白（レターボックス）の塗りは **containBackground** で指定する
- `stretch` … 比率を無視して押し込み（旧称 squeeze と同義）

##### containBackground（`fit: contain` のとき）

| 値 | 意味 |
| --- | --- |
| `black` | 不透明黒（**デフォルト**） |
| `white` | 不透明白 |
| `transparent` | 透明。出力が alpha を持てるときのみ（**PNG / WebP / GIF**）。JPEG は拒否 |
| `#RRGGBB` | 不透明 HEX（大文字小文字不問） |
| `#RRGGBBAA` | alpha 付き HEX（大文字小文字不問） |

- プリセットおよび本体リサイズ制約の両方で指定可。省略時は `black`
- `transparent` と HEX の同時指定はしない（排他）
- GIF / indexed PNG ではアンチエイリアスなしでよい（パレット制約）
- 余白はレターボックス矩形の塗り。画像本体エッジの AA はフォーマット依存

### デフォルト生成セット

プリセット辞書とは別に、Store 時に**自動生成する key の配列**を config で持つ。

```text
thumbnails: {
  presets: {
    sm: { mode: longEdge, size: 320, quality: 0.8 },
    md: { mode: fixed, width: 640, height: 360, fit: cover, quality: 0.75 },
    letterbox: { mode: fixed, width: 640, height: 360, fit: contain, containBackground: "#00FF00AA", quality: 0.8 },
    ...
  },
  defaultKeys: ["sm", "md"]
}
```

- `defaultKeys` に列挙された key だけを、引数未指定時に生成する
- 辞書にあって `defaultKeys` に無い key は、呼び出し側が明示したときだけ生成できる（API 上書き）
- 存在しない key を指定したらエラー

### 返却・永続化

- Store 返却の `variants` は、生成したサムネを **key 付き**で列挙する
- 本体・派生いずれも、寸法が分かる場合は **`width` / `height`（px）** をメタに載せる（不明なら NULL）
- DB 使用時、派生は同一テーブルの子行（`original_id` → 親）。区別用に **`variant_key`**（text、オリジナルは NULL）を持つ（[er.dbml](../../er.dbml) / [db-media.md](../db-media/db-media.md)）

## 推奨処理順

1. 入力解釈（multipart / base64 / bytes / stream）
2. MIME 検査・サイズ上限
3. SVG サニタイズ（対象時・`v0.4.0`）
4. 本体のリサイズ制約（指定時）
5. 圧縮（指定時）
6. サムネイル生成（`defaultKeys` または呼び出し上書きの key 群）
7. ストレージ保存 +（DB 使用時）メタデータ記録

----

以上
