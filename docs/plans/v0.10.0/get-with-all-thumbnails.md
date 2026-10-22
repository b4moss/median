---
type: Plan
title: GetWithAllThumbnails（ファミリー取得）
description: オリジナルと関連サムネイル全体を返す API の仕様詳細計画。
tags: [median, plans, v0.10.0, core-api, db-media]
timestamp: 2026-10-22T09:34:00Z
---

# GetWithAllThumbnails（ファミリー取得）

- **状態**: 仕様詳細（未実装・PO 承認待ち）
- **マイルストーン**: `v0.10.0`
- **Issue**: [#77](https://github.com/b4moss/median/issues/77)（本文が正。2026-10-22 壁打ち決定を反映）
- **前提**: [specs/core-api](../../specs/core-api/core-api.md) / [specs/db-media](../../specs/db-media/db-media.md) / [specs/packages-js](../../specs/packages-js/packages-js.md) / [specs/media-pipeline](../../specs/media-pipeline/media-pipeline.md)
- **関連**: [README.md](./README.md) / [roadmap](../../roadmap.md) / [charter/tdd.md](../../charter/tdd.md) / [charter/versioning-rule.md](../../charter/versioning-rule.md)

## 1. 目的

画像 1 件について、関連するサムネイル（派生行）を含めた **ファミリー全体** を返すメソッドを、Go / JS の公開 API に追加する。

## 2. 対象範囲 / 対象外

### 対象

| 項目 | 内容 |
| --- | --- |
| Go | `(*Median).GetWithAllThumbnails(ctx, id)` |
| JS | `median.getWithAllThumbnails(id)`（`createMedian` が返す面） |
| 返り値 | `Get` が返す **オリジナル** のメタデータ ＋ `Store` 返却と同じ形の `variants`（variant key ごと） |
| id | オリジナルでも派生行でも、`original_id` で親をたどりファミリー全体を返す |
| DB | **必須**。未設定ならエラー |
| status | `status = active` の行のみ。オリジナルが active でなければファミリー全体を「見つからない」 |
| 並び | `variants` は幅の小さい順。幅不明は最後 |
| 版 | minor 上げて Go / npm **`0.10.0`** |

### 対象外

- srcset / `<img>` 等を HTML 文字列で返すメソッド（Issue で中止。サービス層の仕事）
- `Get` / `Delete` / `Store` 既存契約の破壊的変更
- 論理削除の導入、status の書き込み API、status 値の語彙拡張
- DDL / マイグレーション変更（既存 `original_id` / `variant_key` / `status` / `width` で足りる）
- メタ条件検索・一覧 API 一般化
- PHP 移植
- HTML / ベース URL / エスケープの関心

## 3. API シグネチャ案と返り値の型

### 3.1 Go

```go
// 案（名前は Issue どおり。引数・戻り型の細部は未決を参照）
func (m *Median) GetWithAllThumbnails(ctx context.Context, id any) (*GetWithAllThumbnailsResult, error)
```

返り値の骨格（オリジナル＝`Get` メタ、`variants`＝`Store` と同形）:

```go
// Get メタ（Body なし）に Store の variants を足したもの
type GetWithAllThumbnailsResult struct {
	ID         any
	Path       string
	MIME       string
	Size       int64
	Hash       string
	StorageKey string
	Width      *int
	Height     *int
	Variants   []VariantResult // 既存 Store 用型を再利用
}

// VariantResult は現行どおり（Store 返却と同一）
// Key, Path, MIME, Size, Hash, Width, Height, ID, StorageKey
```

既存型の再利用案（推奨・未決）: フィールド集合は現行 `StoreResult` と一致するため、**戻り型を `*StoreResult` にする**、または型エイリアスにする、のどちらかを実装時に選ぶ。専用型を切る場合も JSON / 公開語彙は同じ形にする。

### 3.2 JS

```ts
getWithAllThumbnails(id: string | number): Promise<GetWithAllThumbnailsResult>;
```

```ts
// Get メタ（body なし）+ Store と同形の variants
type GetWithAllThumbnailsResult = {
  id?: string | number;
  path: string;
  mime: string;
  size: number;
  hash: string;
  storageKey: string;
  width?: number | null;
  height?: number | null;
  variants?: VariantResult[]; // 現行 StoreResult.variants と同形
};
```

`Median` 型・`createMedian` 返却オブジェクト・`packages/js` の export にメソッドを追加する。

### 3.3 返り値の意味づけ（確定）

| フィールド | 出所 |
| --- | --- |
| トップレベル | **オリジナル行**のメタ（`Get` が DB あり時に返す範囲。本 API では **body は含めない** 方針 — 未決で確認） |
| `variants` | 親 `id` を `original_id` に持つ **active な派生行**。各要素は `Store` の `variants[]` と同形（`key` = `variant_key`） |

- 派生が 0 件でもオリジナルが active なら成功。`variants` は空配列か省略（未決）
- `storageKey` は DB に無いため、解決したストレージ key（省略時は `defaultKey`）をオリジナル・各 variant に載せる（`Get` と同様）

## 4. エラーの種類

| 状況 | 扱い | 推奨エラー（未決・名前は実装で固定） |
| --- | --- | --- |
| DB 未設定（`dbEnabled == false`） | エラー | 専用: Go `ErrDBRequired` / JS 同等メッセージ（例: `median: db required`） |
| id 空 / nil | エラー | 既存 `ErrIDRequired` 相当を再利用可 |
| 行が存在しない | 見つからない | 既存 `db.ErrNotFound` / JS `db: media not found` |
| オリジナルが `active` でない | ファミリー全体を見つからない | 同上 not found（呼び出し側は存在有無と区別しない） |
| 未知の storage key（オプションを取る場合） | エラー | 既存 `ErrUnknownKey` |

「見つからない」と「DB なし」は **別エラー** とし、呼び出し側が分岐できるようにする。

## 5. DB クエリの考え方

スキーマ変更なし。テーブル名上書き・カラムマップ・ID 戦略は既存 Repo 経路をそのまま使う。

### 5.1 処理手順（言語共通）

1. DB 必須チェック。無ければ DB なしエラー
2. `FindByID(id)` で行を取得（存在しなければ not found）
3. 親 id の決定:
   - `original_id` が非 NULL → それが親 id
   - NULL → 当該行がオリジナル（親 id = 自身の id）
4. 親行を取得（手順 2 の行が親なら再取得不要）
5. 親の `status == "active"` でなければ not found（ファミリー全体）
6. 親 id で子行を列挙し、**`status == "active"` のみ**残す
7. 子を幅昇順にソート（幅 NULL / 不明は末尾）。タイブレークは未決
8. 親メタ ＋ `variants` を組み立てて返す（バイナリは読まない）

### 5.2 カラムマップ・テーブル名・ ID 戦略

| 関心 | 対応 |
| --- | --- |
| テーブル名 | Config の `tableName`（既定 `media`）。Repo の既存経路 |
| カラムマップ | 論理名 → 物理名（`id` / `original_id` / `variant_key` / `status` / `width` 等）。既存 `ColumnMap` |
| ID 戦略 | `id` / `original_id` は `any`（Go）/ `string \| number`（JS）。auto_increment / uuid_v4 / uuid_v7 / ulid いずれも FindByID・ListChildren の既存契約に乗る |
| 方言 | Go: crudian + 既存 Dialects。JS 実行時: better-sqlite3 のみ（従来どおり） |

### 5.3 Repo 層の拡張案（実装詳細・未決）

| 案 | 内容 |
| --- | --- |
| A（推奨） | 既存 `FindByID` + `ListChildren` を使い、**status フィルタとソートは core 層**で行う |
| B | `ListChildren` に status 条件を足す、または `ListActiveChildren` を追加 |

A なら DB テスト追加は最小、B なら Repository の契約変更が必要。

派生 id が inactive のときも親へ辿るかは **未決**（推奨: 行が存在すれば status に関わらず親を辿り、最終的に親が active ならファミリーを返す）。

## 6. 更新が必要な specs / tests（実装時）

実装完了後に plans → specs 反映（charter: 実装 worktree は specs を触らない。main / 設計側で昇格）。

| 文書 | 更新内容 |
| --- | --- |
| [specs/core-api](../../specs/core-api/core-api.md) | 公開語彙に `GetWithAllThumbnails` を追加。入出力・DB 必須・status・並び・エラー |
| [specs/packages-js](../../specs/packages-js/packages-js.md) | 入口メソッド一覧に `getWithAllThumbnails` |
| [specs/db-media](../../specs/db-media/db-media.md) | Repo 契約を変える場合のみ追記（status フィルタ付き List 等）。変えないなら「core が status を見る」と core-api 側に書く |
| [tests/core-api](../../tests/core-api/core-api.md) | 下記 TDD 観点を正式なテスト仕様へ転記（ドメインが DB 寄りなら [tests/db-media](../../tests/db-media/db-media.md) と分担可） |
| [tests/packages-js](../../tests/packages-js/packages-js.md) | JS 同等観点 |
| [specs/e2e](../../specs/e2e/e2e.md) / [tests/e2e](../../tests/e2e/e2e.md) | シナリオ追加する場合のみ（§7） |
| ルート / 言語 README | 公開 API 一覧にメソッド名を 1 行追加（任意・実装時） |

**本 PR（計画のみ）では `specs/` / `tests/` は更新しない。**

## 7. TDD 受け入れテスト観点一覧

実装時に `docs/tests/` へ正式化し、Red → Green。氷山パターン（過剰カバレッジを追わない）。

### GetWithAllThumbnails / getWithAllThumbnails

- 渡された id からオリジナルを解決し、active なオリジナルメタと active な variants を返す
- variants は Store 返却と同形（key / path / mime / size / hash / width / height / id / storageKey）
- variants は幅の小さい順。幅不明は最後
- テーブル名上書き・カラムマップ・各 ID 戦略でも同じ契約（代表ケースでよい）

#### テスト：正常系

- オリジナル id で、複数サムネ付きファミリーが返り、variants が幅昇順
- 派生行 id で同じファミリーが返る（オリジナルメタは親のもの）
- サムネ無しの active オリジナルは成功し、variants は空（または省略）
- 子に inactive が混在する場合、active の子だけが variants に入る

#### テスト: 異常系

- DB 未設定で呼ぶと DB 必須エラー（not found ではない）
- 存在しない id は not found
- オリジナルが active でないとき（派生 id 経由でも）not found
- id 空は拒否

### Repo（案 B を採る場合のみ追加）

- status 条件付き ListChildren がカラムマップ・テーブル名を尊重する

## 8. E2E への影響

| 項目 | 方針 |
| --- | --- |
| 必須度 | **未決**。クリティカルなら T-L / T-S の延長、または薄い G シナリオを 1 本 |
| 推奨 | 単体・結合（SQLite）で契約を厚くし、E2E は「Store でサムネ生成 → GetWithAllThumbnails → variants 件数・幅順」を **Local 1 本**に留める |
| 通常 CI | 引き続き含めない（既存 e2e 方針） |
| ハーネス | Compose / RustFS 変更は不要の見込み |

## 9. リリース（タグ）手順

現行 `0.9.0` → **`0.10.0`**（MINOR。後方互換の新 API）。

1. 実装・テスト完了後、`dev-v0.10.0` → `develop` → `main`（git-rule）
2. 版ファイルを `0.10.0` に揃える
   - `packages/go/VERSION`
   - `packages/js/package.json`（lock があれば追随）
3. **main** 上でタグを打つ（versioning-rule）
   - Go: `packages/go/v0.10.0`（`publish-go` / Release on tag）
   - npm: ルート `v0.10.0`（`publish-npm` / Release on tag）
4. GitHub Release は既存 `release-on-tag.yml` に任せる
5. roadmap / specs の版表記を `0.10.0` に更新し、本 plans を specs へ昇格・履歴は `_archived/history/v0.10.0/` へ

## 10. 作業の順番

1. **本計画の PO 承認**（マージは承認後）
2. `docs/tests/` にテスト仕様を追記（TDD 入力）
3. Go: 失敗するテスト → `GetWithAllThumbnails` 実装 → Green
4. JS: 同等テスト → `getWithAllThumbnails` 実装 → Green
5. （承認された範囲で）E2E シナリオ追加・手動 Workflow
6. 版バンプ `0.10.0`、README の API 一覧更新
7. `dev-v0.10.0` PR → develop → main、タグ打ち
8. main 側で plans → specs 昇格、roadmap を出荷済に更新

並行: Go と JS は契約を先に固定すれば並列可。Repo 拡張が必要なら Go DB を先に。

## 11. 未決事項

勝手に確定しない。推奨案のみ付す。

| ID | 未決 | 推奨案 |
| --- | --- | --- |
| U1 | 戻り型を専用型にするか、`StoreResult` を再利用するか | フィールド一致のため **`StoreResult` 再利用**（または型エイリアス）。文書上の名前は Issue どおり GetWithAllThumbnails |
| U2 | `variants` 0 件時に空配列か省略か | Go は空スライス / JS は `[]` を返す（呼び出し側の分岐を簡単に） |
| U3 | `WithBody` / バイナリ同梱を取るか | **取らない**（メタ＋variants のみ） |
| U4 | `storageKey` 引数をメソッドに持たせるか | **省略可の第 2 引数 or options**（`Get` に合わせる）。省略時 defaultKey |
| U5 | 派生 id が inactive のとき親へ辿るか | **行が存在すれば辿る**。最終判定は親の active |
| U6 | 幅が同じ variants のタイブレーク | `variant_key` 昇順（安定ソート） |
| U7 | status フィルタを Repo に下ろすか core か | **core でフィルタ**（既存 ListChildren のまま） |
| U8 | DB なしエラーの型名 | Go `ErrDBRequired`、JS `MedianError("median: db required")` 等。既存エラー階層に合わせる |
| U9 | 既存 `Get` も active のみに揃えるか | **本版では変更しない**（本 API のみ status を見る）。Get の status 扱いは別 Issue |
| U10 | E2E にシナリオを足すか | 単体厚め＋E2E は Local 1 本（任意）。必須化は PO 判断 |
| U11 | 幅 NULL 同士の順序、および width=0 の扱い | NULL/不明のみ末尾。0 は通常の数値として昇順に含める |
| U12 | オリジナル行に `variant_key` が入っている不正データの扱い | 親判定は `original_id` のみ。不正データはスコープ外（見つかればそのまま） |

## 12. 受け入れ（短い・roadmap 用）

DB ありで `GetWithAllThumbnails` / `getWithAllThumbnails` が、オリジナルまたは派生 id から active なファミリー（オリジナルメタ＋幅昇順 variants）を返す。DB なしは専用エラー。Go / npm `0.10.0`。

----

以上
