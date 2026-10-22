---
type: Plan
title: GetWithAllThumbnails（ファミリー取得）
description: オリジナルと関連サムネイル全体を返す API の仕様詳細計画。
tags: [median, plans, v0.10.0, core-api, db-media]
timestamp: 2026-10-22T10:54:00Z
---

# GetWithAllThumbnails（ファミリー取得）

- **状態**: 仕様詳細（未実装・PO 承認待ち。U1〜U12 は PO 決定済み）
- **マイルストーン**: `v0.10.0`
- **Issue**: [#77](https://github.com/b4moss/median/issues/77)（本文が正。2026-10-22 壁打ち決定＋同日 U1〜U12 PO 決定を反映）
- **前提**: [specs/core-api](../../specs/core-api/core-api.md) / [specs/db-media](../../specs/db-media/db-media.md) / [specs/packages-js](../../specs/packages-js/packages-js.md) / [specs/media-pipeline](../../specs/media-pipeline/media-pipeline.md)
- **関連**: [README.md](./README.md) / [roadmap](../../roadmap.md) / [charter/tdd.md](../../charter/tdd.md) / [charter/versioning-rule.md](../../charter/versioning-rule.md)

## 1. 目的

画像 1 件について、関連するサムネイル（派生行）を含めた **ファミリー全体** を返すメソッドを、Go / JS の公開 API に追加する。

## 2. 対象範囲 / 対象外

### 対象

| 項目 | 内容 |
| --- | --- |
| Go | `(*Median).GetWithAllThumbnails(ctx, id, …)`（storageKey は Get と同じく任意） |
| JS | `median.getWithAllThumbnails(id, …)`（同上） |
| 返り値 | 既存 **`StoreResult`**（各要素も既存の Store / Variant 形）。トップレベルはオリジナルのメタ、`variants` は Store と同形 |
| id | オリジナルでも派生行でも、`original_id` だけで親をたどりファミリー全体を返す |
| DB | **必須**。未設定なら Go `ErrDBRequired` / JS `DBRequiredError` |
| status | 返すのは `status = active` の行のみ（絞り込みは **core**）。オリジナルが active でなければファミリー全体を「見つからない」。渡された id が inactive な派生でも親をたどる |
| 並び | `variants` は幅の小さい順。幅 `0` は通常の数値として含める。幅 NULL（不明）のみ最後。同幅は `variant_key` 昇順 |
| 版 | minor 上げて Go / npm **`0.10.0`** |

### 対象外

- srcset / `<img>` 等を HTML 文字列で返すメソッド（Issue で中止。サービス層の仕事）
- `Get` / `Delete` / `Store` 既存契約の破壊的変更（本版の `Get` は active 絞り込みしない）
- `WithBody` 等のバイナリ同梱オプション（本 API はメタデータのみ）
- 論理削除の導入、status の書き込み API、status 値の語彙拡張
- DDL / マイグレーション変更（既存 `original_id` / `variant_key` / `status` / `width` で足りる）
- メタ条件検索・一覧 API 一般化
- PHP 移植
- HTML / ベース URL / エスケープの提供

## 3. API シグネチャと返り値の型

### 3.1 Go

```go
// storageKey は Get と同じく任意（省略時は defaultKey）。第3引数の形は実装で Get に揃える
func (m *Median) GetWithAllThumbnails(ctx context.Context, id any, storageKey ...string) (*StoreResult, error)
```

- 戻り値は既存 **`*StoreResult`**（U1）
- `Variants` はサムネ無しでも **空スライス**（nil にしない。U2）
- Body / WithBody は持たない（U3）

`VariantResult` は現行 Store 返却と同一（Key, Path, MIME, Size, Hash, Width, Height, ID, StorageKey）。

### 3.2 JS

```ts
getWithAllThumbnails(
  id: string | number,
  opt?: { storageKey?: string }, // Get と同じ扱い（任意）
): Promise<StoreResult>;
```

- 戻り値は既存 **`StoreResult`**（U1）
- サムネ無しでも `variants: []`（U2）
- `body` / `withBody` は付けない（U3）

`Median` 型・`createMedian` 返却オブジェクト・`packages/js` の export にメソッドを追加する。

### 3.3 返り値の意味づけ

| フィールド | 出所 |
| --- | --- |
| トップレベル | **オリジナル行**のメタ（`StoreResult` のトップレベル。body なし） |
| `variants` | 親 `id` を `original_id` に持つ **active な派生行**。各要素は `Store` の `variants[]` と同形（`key` = `variant_key`）。0 件なら空配列 |

- `storageKey` は DB に無いため、解決したストレージ key（省略時は `defaultKey`）をオリジナル・各 variant に載せる（`Get` と同様。U4）

## 4. エラーの種類

| 状況 | 扱い | エラー |
| --- | --- | --- |
| DB 未設定（`dbEnabled == false`） | エラー | Go **`ErrDBRequired`** / JS **`DBRequiredError`**（U8） |
| id 空 / nil | エラー | 既存 `ErrIDRequired` 相当を再利用可 |
| 行が存在しない | 見つからない | 既存 `db.ErrNotFound` / JS `db: media not found` |
| オリジナルが `active` でない | ファミリー全体を見つからない | 同上 not found（呼び出し側は存在有無と区別しない） |
| 未知の storage key | エラー | 既存 `ErrUnknownKey` |

「見つからない」と「DB なし」は **別エラー** とし、呼び出し側が分岐できるようにする。

## 5. DB クエリの考え方

スキーマ変更なし。テーブル名上書き・カラムマップ・ID 戦略は既存 Repo 経路をそのまま使う。

### 5.1 処理手順（言語共通）

1. DB 必須チェック。無ければ `ErrDBRequired` / `DBRequiredError`
2. `FindByID(id)` で行を取得（存在しなければ not found）。**この行の status は問わない**（inactive な派生 id でも続行。U5）
3. 親 id の決定（**`original_id` のみ**。`variant_key` 等から推測しない。U12）:
   - `original_id` が非 NULL → それが親 id
   - NULL → 当該行がオリジナル（親 id = 自身の id）
4. 親行を取得（手順 2 の行が親なら再取得不要）
5. 親の `status == "active"` でなければ not found（ファミリー全体）
6. 親 id で子行を列挙し、**core 側で** `status == "active"` のみ残す（U7）
7. 子をソート（U6 / U11）:
   - 幅が分かるもの（`0` 含む）を幅昇順
   - 同幅は `variant_key` 昇順
   - 幅 NULL（不明）のみ末尾
8. 親メタ ＋ `variants`（空配列可）を `StoreResult` として返す（バイナリは読まない）

### 5.2 カラムマップ・テーブル名・ ID 戦略

| 関心 | 対応 |
| --- | --- |
| テーブル名 | Config の `tableName`（既定 `media`）。Repo の既存経路 |
| カラムマップ | 論理名 → 物理名（`id` / `original_id` / `variant_key` / `status` / `width` 等）。既存 `ColumnMap` |
| ID 戦略 | `id` / `original_id` は `any`（Go）/ `string \| number`（JS）。auto_increment / uuid_v4 / uuid_v7 / ulid いずれも FindByID・ListChildren の既存契約に乗る |
| 方言 | Go: crudian + 既存 Dialects。JS 実行時: better-sqlite3 のみ（従来どおり） |

### 5.3 Repo 層

既存 `FindByID` + `ListChildren` を使い、**status フィルタとソートは core 層**で行う（U7）。Repo に status 条件付き List は追加しない。

## 6. 更新が必要な specs / tests（実装時）

実装完了後に plans → specs 反映（charter: 実装 worktree は specs を触らない。main / 設計側で昇格）。

| 文書 | 更新内容 |
| --- | --- |
| [specs/core-api](../../specs/core-api/core-api.md) | 公開語彙に `GetWithAllThumbnails` を追加。入出力（`StoreResult`）・DB 必須・status・並び・エラー |
| [specs/packages-js](../../specs/packages-js/packages-js.md) | 入口メソッド一覧に `getWithAllThumbnails` |
| [specs/db-media](../../specs/db-media/db-media.md) | Repo 契約は変えない。「core が status を見る」は core-api 側に書く |
| [tests/core-api](../../tests/core-api/core-api.md) | 下記 TDD 観点を正式なテスト仕様へ転記（ドメインが DB 寄りなら [tests/db-media](../../tests/db-media/db-media.md) と分担可） |
| [tests/packages-js](../../tests/packages-js/packages-js.md) | JS 同等観点 |
| [specs/e2e](../../specs/e2e/e2e.md) / [tests/e2e](../../tests/e2e/e2e.md) | Local E2E を任意 1 本足す場合（§8） |
| ルート / 言語 README | 公開 API 一覧にメソッド名を 1 行追加（任意・実装時） |

**本 PR（計画のみ）では `specs/` / `tests/` は更新しない。**

## 7. TDD 受け入れテスト観点一覧

実装時に `docs/tests/` へ正式化し、Red → Green。氷山パターン。**単体テストを厚めに書く**（U10）。

### GetWithAllThumbnails / getWithAllThumbnails

- 渡された id からオリジナルを解決し、active なオリジナルメタと active な variants を `StoreResult` で返す
- variants は Store 返却と同形（key / path / mime / size / hash / width / height / id / storageKey）
- variants は幅昇順（0 含む）。同幅は variant_key 昇順。幅 NULL のみ末尾
- テーブル名上書き・カラムマップ・各 ID 戦略でも同じ契約（代表ケースでよい）

#### テスト：正常系

- オリジナル id で、複数サムネ付きファミリーが返り、variants が幅昇順・同幅は key 昇順
- 派生行 id で同じファミリーが返る（オリジナルメタは親のもの）
- inactive な派生 id でも親が active ならファミリーが返る
- サムネ無しの active オリジナルは成功し、`variants` は空配列
- 子に inactive が混在する場合、active の子だけが variants に入る
- width=0 の variant が数値として昇順に含まれる
- storageKey 省略時は defaultKey、明示時はその key が載る

#### テスト: 異常系

- DB 未設定で呼ぶと `ErrDBRequired` / `DBRequiredError`（not found ではない）
- 存在しない id は not found
- オリジナルが active でないとき（派生 id 経由でも）not found
- id 空は拒否
- 未知の storage key は拒否

## 8. E2E への影響

| 項目 | 方針（U10） |
| --- | --- |
| 単体 | 厚めに書く（契約の主担保） |
| E2E | **任意で Local 1 本**（例: Store でサムネ生成 → GetWithAllThumbnails → variants 件数・幅順） |
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
5. （任意）Local E2E 1 本・手動 Workflow
6. 版バンプ `0.10.0`、README の API 一覧更新
7. `dev-v0.10.0` PR → develop → main、タグ打ち
8. main 側で plans → specs 昇格、roadmap を出荷済に更新

並行: Go と JS は契約固定済みのため並列可。Repo 拡張は不要。

## 11. 決定事項

PO（こうきさん）決定。旧 U1〜U12。

| ID | 決定 |
| --- | --- |
| U1 | 戻り値の各要素は既存の **`StoreResult`** を再利用する |
| U2 | サムネイルがない場合、`variants` は **空配列**で返す（エラーにしない） |
| U3 | `WithBody` などバイナリ取得のオプションは付けない（**メタデータのみ**） |
| U4 | ストレージキーの指定は **任意**とし、既存の `Get` と同じ扱いにする |
| U5 | 渡された id が **inactive な派生**でも親をたどる。返すのは active なファミリーだけ（親が inactive なら not found） |
| U6 | 幅が同じ場合は **`variant_key` の昇順**で並べる |
| U7 | active の絞り込みは **core（本体）側**で行う |
| U8 | DB なしで呼ばれたときのエラーは Go **`ErrDBRequired`**、JS **`DBRequiredError`** |
| U9 | 既存の `Get` は今回の版では変えない（active のみに絞らない）。将来の扱いは [§12 要検討事項](#12-要検討事項) |
| U10 | テストは **単体を厚め**に書き、**Local E2E を任意で 1 本**だけ足す |
| U11 | 幅 **0** はふつうの数値として並べ順に含める。最後に回すのは幅が **NULL（不明）** のものだけ |
| U12 | 親の判定は **`original_id` だけ**で行う（`variant_key` など他の値から推測しない） |

## 12. 要検討事項

Issue にはしない。将来の版で検討するメモ。

| 項目 | 内容 |
| --- | --- |
| 既存 `Get` の status | 本版（`v0.10.0`）では `Get` は active 絞り込みしない。将来「`Get` も active のみに揃えるか」を検討する（別 Issue 化はしない。必要になったら本節または wishlist / 次版 plans で起こす） |

## 13. 受け入れ（短い・roadmap 用）

DB ありで `GetWithAllThumbnails` / `getWithAllThumbnails` が、オリジナルまたは派生 id から active なファミリー（`StoreResult`：オリジナルメタ＋幅昇順 variants、無しは空配列）を返す。DB なしは `ErrDBRequired` / `DBRequiredError`。Go / npm `0.10.0`。

----

以上
