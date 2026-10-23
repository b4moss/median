---
type: Spec
title: テスト仕様 — core-api
description: Go コア API（Store / Delete / Get / GetWithAllThumbnails）の TDD 入力。
tags: [median, tests, core-api, go, v0.11.0]
timestamp: 2026-10-22T11:10:00Z
---

# テスト仕様 — core-api

対象ドメイン: `core-api`（Go コア API）  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/core-api/`](../../specs/core-api/)  
計画: [`../../plans/v0.11.0/get-with-all-thumbnails.md`](../../plans/v0.11.0/get-with-all-thumbnails.md)（Issue [#77](https://github.com/b4moss/median/issues/77)）  
関連: [`../storage/`](../storage/) / [`../db-media/`](../db-media/) / [`../packages-js/`](../packages-js/) / [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`tdd.md`](../../charter/tdd.md)（氷山パターン）

旧マイルストーン `v0.2.0` 由来。storage 単体は [`../storage/`](../storage/) を参照。  
**v0.11.0**: `GetWithAllThumbnails` は **DB 必須**（下記専用節）。それ以外の本ファイル節は従来どおり DB なし契約。

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/packages/go` |
| 対象パッケージ | `core` / `storage` / `storage/local`（`internal` は単体対象可） |
| ランタイム / テスト | Go `1.26.x` + `go test` |
| ストレージ | Local FS（テストは `t.TempDir()` 等） |
| 入口 | `core.New(Config)` → `Store` / `Delete` / `Get` / **`GetWithAllThumbnails`（v0.11.0・DB 必須）** |
| DB / 画像 / S3 | 従来節は **対象外**（`id`・`width`・`height`・`variants` は返さない）。**例外**: `GetWithAllThumbnails` 節のみ DB・variants を扱う |
| 依存 | `github.com/b4moss/shardian/packages/go`（DB 節は crudian / 既存 MediaRepo 経路） |

### 実装固定デフォルト

| 項目 | 値 |
| --- | --- |
| 同時アップロード | Median **インスタンスあたり**の in-flight `Store` 数。デフォルト **20** |
| サイズ上限 | デフォルト **20 MiB**（config で変更可） |
| MIME | 呼び出し側の**宣言 MIME のみ**。`allow` 非空なら許可リスト、`deny` は常に拒否 |
| `hash` | Store 時に **SHA-256 hex** を計算して返却 |
| ファイル名デフォルト | `random`（hex 長さデフォルト **32** 文字 = 16 byte） |
| shardian | config の **DirLetterCount / DirNestDepth 必須**（未設定は `New` 失敗） |
| `actor` | 受け取れるが DB なしでは無視（未指定でもエラーにしない） |

Store 返却（本版）: `{ path, mime, size, hash, storageKey }`

---


## Config / New

### New

- Config を検証し、Median インスタンスを生成する
- storages / defaultKey / shardian 必須項目を確認する

#### テスト：正常系

- 有効な Config（defaultKey が storages に存在、shardian 幅・深さあり、local ルートあり）でインスタンスが得られる
- MIME allow/deny・maxSize・maxConcurrentStores・filename デフォルトを省略すると実装固定デフォルトが使われる
- 複数 storage key を登録できる

#### テスト: 異常系

- `defaultKey` が storages に無いとき拒否する
- storages が空のとき拒否する
- DirLetterCount または DirNestDepth が未設定のとき拒否する
- local driver の root path が空のとき拒否する
- 未知の driver を指定したとき拒否する

---


## MIME / サイズ

### checkMIME

- 宣言 MIME のみを見る（スニッフィングしない）
- `deny` に含まれれば拒否。`allow` が非空ならその集合外は拒否

#### テスト：正常系

- allow/deny とも空なら任意の宣言 MIME を通す
- allow に含まれる MIME を通す
- deny に無い MIME を通す（allow 空のとき）

#### テスト: 異常系

- deny に含まれる MIME は拒否する
- allow 非空かつ候補外の MIME は拒否する
- 空の宣言 MIME は拒否する

### checkSize

- ペイロードサイズが上限以下であることを確認する

#### テスト：正常系

- デフォルト上限（20 MiB）未満を通す
- config で上げた上限内を通す

#### テスト: 異常系

- デフォルト上限を超えるサイズは拒否する
- サイズ 0 は通すか拒否かを実装で固定（本仕様では **通す** = 空ファイル許可）
- 負のサイズは拒否する

---


## core.Median

### Store

- 入力バイトを MIME/サイズ検査し、ファイル名解決・shardian・storage key 解決のうえ Adapter.Put する
- SHA-256 hex を計算して返却に含める
- 未指定の storage key は defaultKey。ファイル名モード未指定は config デフォルト（random）
- in-flight Store が上限に達している間はブロックまたはエラー（実装は **セマフォで待機**、キャンセル可能なら context で解除）

#### テスト：正常系

- `[]byte` 入力で `{ path, mime, size, hash, storageKey }` が返り、Local FS 上に実体がある
- 未指定 key で defaultKey が `storageKey` に入る
- 明示した storage key が `storageKey` に入り、そのルート配下に書かれる
- `hash` が入力の SHA-256 hex と一致する
- `preserve` 指定時、サニタイズ後名が path 末尾に反映される
- `io.Reader` 入力でも同等に保存できる

#### テスト: 異常系

- deny MIME / サイズ超過で Put されない
- 存在しない storage key は拒否する
- context キャンセルで中断できる（セマフォ待ちまたは書き込み中）
- Adapter.Put が失敗したらエラーを返し、呼び出し側に成功メタを返さない

### Delete

- DB なしでは **path + storage key** で実体を削除する（key 省略時は default）

#### テスト：正常系

- Store した path を Delete するとファイルが消える
- 明示 key のルート配下のみ削除される

#### テスト: 異常系

- 存在しない path は not found 相当、または冪等成功のいずれかで一貫する
- 存在しない storage key は拒否する
- path 空は拒否する

### Get

- DB なしでは **path + storage key** で解決する
- デフォルトはメタのみ（少なくとも path / size / storageKey。mime/hash は保持していなければ省略可）
- オプションでバイナリを同梱できる

#### テスト：正常系

- Store 後 Get（メタのみ）で path / size / storageKey が分かる
- バイナリ同梱オプションで内容が Store 時と一致する
- 明示 key で正しいルートから読む

#### テスト: 異常系

- 存在しない path は not found 相当
- 存在しない storage key は拒否する
- path 空は拒否する

---


## 同時アップロード

### acquireStoreSlot

- インスタンスあたり in-flight Store を maxConcurrentStores（デフォルト 20）に制限する

#### テスト：正常系

- 上限未満の並列 Store はすべて成功する
- 上限ちょうどまで同時に取得できる

#### テスト: 異常系

- 上限中の追加 Store は、スロット解放まで待機する（テストでは短い上限 + 解放後の成功で確認）
- 待機中に context キャンセルするとエラーになり、スロットを漏らさない

---


## 入力ヘルパ

### DecodeBase64

- base64 文字列、または data URL（`data:<mime>;base64,...`）からバイト列と MIME を得る

#### テスト：正常系

- 素の base64 から元バイトが復元される
- data URL から MIME とバイトが取れる
- パディングあり/なしの正当な入力を受け入れる（実装が一方のみならそれに合わせる）

#### テスト: 異常系

- 不正な base64 は拒否する
- data URL で base64 でないスキームは拒否する
- 空文字は拒否する

### PartToStoreOptions（JS: `partToStoreInput`）

- 抽出済み multipart part（ファイル名 + MIME + バイト / Reader）を Store 入力に変換する
- HTTP Request 丸ごとのパースは必須テスト対象外
- 関数名は Go `PartToStoreOptions` / JS `partToStoreInput`（`PartFromMultipart` という別名は無い）

#### テスト：正常系

- part の filename / MIME / 本文が Store オプションと入力に載る
- 小さな part を Store まで通せる

#### テスト: 異常系

- 本文が無い part は拒否する
- ファイル名も MIME も無い part の扱いは実装で固定（本仕様では **MIME 必須**、filename は random デフォルトで補える）

---


## GetWithAllThumbnails（v0.11.0・DB 必須）

計画正本: [`../../plans/v0.11.0/get-with-all-thumbnails.md`](../../plans/v0.11.0/get-with-all-thumbnails.md)（U1〜U12）。  
JS 同等: [`../packages-js/packages-js.md`](../packages-js/packages-js.md)。Repo 契約は変えない（[`../db-media/`](../db-media/)）。status 絞り込みとソートは **core**（U7）。

### 前提（本節のみ）

| 項目 | 値 |
| --- | --- |
| 入口 | `(*Median).GetWithAllThumbnails(ctx, id any, opt GetWithAllThumbnailsOptions)`（構造体は `{StorageKey}` のみ。`StorageKey` は既存 `Get` と同じく任意。id は必須の別引数） |
| 戻り値 | 既存 `*StoreResult`（トップレベル＝オリジナルメタ。`Variants` は Store 返却と同形。無しは **空スライス**、nil にしない） |
| Body | 付けない（メタのみ。`WithBody` なし） |
| DB | **必須**。未設定は `ErrDBRequired`（not found ではない） |
| 親解決 | `original_id` のみ（非 NULL → 親 id、NULL → 自身が親）。`variant_key` 等から推測しない |
| status | 返すのは `active` のみ。絞り込みは core。渡された id が inactive 派生でも親をたどる。親が active でなければファミリー全体 not found |
| 並び | 幅昇順（`0` 含む）→ 同幅は `variant_key` 昇順 → 幅 NULL（不明）のみ末尾 |
| Repo | 既存 `FindByID` + `ListChildren`。status 条件付き List は追加しない |
| テスト DB | SQLite（既存 db-media 経路）。テーブル名上書き・カラムマップ・ID 戦略は代表ケースでよい |
| 既存 `Get` | 本版では変えない（active 絞り込みしない。U9） |

### GetWithAllThumbnails

- 渡された id からオリジナルを解決し、active なオリジナルメタと active な variants を `StoreResult` で返す
- variants は Store 返却と同形（Key / Path / MIME / Size / Hash / Width / Height / ID / StorageKey）
- variants は幅昇順（0 含む）。同幅は `variant_key` 昇順。幅 NULL のみ末尾
- バイナリは読まない（ストレージ Get を呼ばない）

#### テスト：正常系

- オリジナル id で、複数サムネ付きファミリーが返り、トップレベルは親メタ、`Variants` が幅昇順・同幅は key 昇順
- 派生行（active）id で同じファミリーが返る（トップレベルは親のもの。派生自身のメタがトップにならない）
- inactive な派生 id でも親が active ならファミリーが返る（入力行の status は問わない）
- サムネ無しの active オリジナルは成功し、`Variants` は長さ 0 のスライス（nil ではない）
- 子に inactive が混在する場合、active の子だけが `Variants` に入る（inactive 子は除外）
- `width=0` の variant が数値として昇順に含まれる（NULL 扱いにしない）
- 幅 NULL の variant のみ、数値幅のあとに並ぶ
- 同幅の複数 variant は `variant_key` 昇順
- `opt.StorageKey` 省略時（ゼロ値）は defaultKey、明示時はその key がオリジナル・各 variant の `StorageKey` に載る
- テーブル名上書き、またはカラムマップ、または非 `auto_increment` ID 戦略のいずれか代表ケースでも同じ契約（全組み合わせの網羅は不要）

#### テスト: 異常系

- DB 未設定（`dbEnabled == false`）で呼ぶと `ErrDBRequired`（既存 not found / `ErrNotFound` ではない）
- 存在しない id は not found（既存 `db.ErrNotFound` 相当）
- オリジナルが active でないとき（派生 id 経由でも）not found（ファミリー全体）
- id 空 / nil は拒否（既存 `ErrIDRequired` 相当を再利用可）
- 未知の storage key は拒否（既存 `ErrUnknownKey`）

### 本節の対象外

- `getWithAllThmbnailsAsHTMLSrcSet` 等 HTML / srcset 返却
- 既存 `Get` の active 絞り込み変更（将来検討。計画 §12）
- Repo への status 条件付き List 追加
- DDL / マイグレーション変更
- Local E2E（任意 1 本は [`../e2e/e2e.md`](../e2e/e2e.md)）

---


## 対象外（本版・従来節）

- DB 永続化、`id` 採番、重複 hash 抑止、actor 永続化（**例外**: 上記 `GetWithAllThumbnails`）
- 画像の width/height / variants / パイプライン（**例外**: 上記 `GetWithAllThumbnails` のメタ）
- S3 Adapter・署名付き URL
- SVG サニタイズ、PDF サムネ
- HTTP Request 全体の multipart パース必須化

----

以上

