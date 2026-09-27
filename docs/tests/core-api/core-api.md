---
type: Spec
title: テスト仕様 — core-api
description: Go コア API（Store / Delete / Get）の TDD 入力。
tags: [median, tests, core-api, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — core-api

対象ドメイン: `core-api`（Go コア API）  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/core-api/`](../../specs/core-api/)  
関連: [`../storage/`](../storage/) / [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`tdd.md`](../../charter/tdd.md)（氷山パターン）

旧マイルストーン `v0.2.0` 由来。storage 単体は [`../storage/`](../storage/) を参照。

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/go` |
| 対象パッケージ | `core` / `storage` / `storage/local`（`internal` は単体対象可） |
| ランタイム / テスト | Go `1.26.x` + `go test` |
| ストレージ | Local FS（テストは `t.TempDir()` 等） |
| 入口 | `core.New(Config)` → `Store` / `Delete` / `Get` |
| DB / 画像 / S3 | **対象外**（`id`・`width`・`height`・`variants` は返さない） |
| 依存 | `github.com/b4moss/shardian/packages/go` |

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

### PartFromMultipart

- 抽出済み multipart part（ファイル名 + MIME + バイト / Reader）を Store 入力に変換する
- HTTP Request 丸ごとのパースは本版の必須テスト対象外

#### テスト：正常系

- part の filename / MIME / 本文が Store オプションと入力に載る
- 小さな part を Store まで通せる

#### テスト: 異常系

- 本文が無い part は拒否する
- ファイル名も MIME も無い part の扱いは実装で固定（本仕様では **MIME 必須**、filename は random デフォルトで補える）

---


## 対象外（本版）

- DB 永続化、`id` 採番、重複 hash 抑止、actor 永続化
- 画像の width/height / variants / パイプライン
- S3 Adapter・署名付き URL
- SVG サニタイズ、PDF サムネ
- HTTP Request 全体の multipart パース必須化

----

以上

