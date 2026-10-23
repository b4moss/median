---
type: Spec
title: テスト仕様 — storage
description: Local FS Adapter・storage key・ファイル名・shardian の TDD 入力。
tags: [median, tests, storage, go]
timestamp: 2026-09-27T05:00:00Z
---

# テスト仕様 — storage

対象ドメイン: `storage`（Local FS・key・ファイル名・shardian）  
製品: [`../../README.md`](../../README.md)  
仕様: [`../../specs/storage/`](../../specs/storage/)  
関連: [`../core-api/`](../core-api/) / [`../../roadmap.md`](../../roadmap.md)  
書き方: charter [`tdd.md`](../../charter/tdd.md)（氷山パターン）

旧マイルストーン `v0.2.0` 由来。コア API は [`../core-api/`](../core-api/) を参照。

## 共通前提

| 項目 | 値 |
| --- | --- |
| Module | `github.com/b4moss/median/packages/go` |
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


## storage.Adapter / local

### Put

- ストレージルート配下の相対 path にバイト列を書き込む
- 親ディレクトリが無ければ作成する

#### テスト：正常系

- 相対 path に書いた内容が読める
- ネストした相対 path でも親ディレクトリが作られる
- 同一 path への上書きで内容が置き換わる

#### テスト: 異常系

- `..` を含む相対 path は拒否する
- 絶対 path は拒否する
- 空の相対 path は拒否する

### Get

- 相対 path の内容を読み取る

#### テスト：正常系

- Put した内容と一致するバイト列が返る
- サイズが期待どおり

#### テスト: 異常系

- 存在しない path は not found 相当で失敗する
- `..` を含む path は拒否する

### Delete

- 相対 path のファイルを削除する

#### テスト：正常系

- 存在するファイルを削除すると Exists が false になる
- 削除後 Get は not found 相当になる

#### テスト: 異常系

- 存在しない path の Delete は not found 相当、または冪等な成功のいずれかで一貫する（実装で一方に固定しテストも合わせる）
- `..` を含む path は拒否する

### Exists

- 相対 path の存在を真偽で返す

#### テスト：正常系

- Put 後は true、Delete 後は false
- 一度も書いていない path は false

#### テスト: 異常系

- `..` を含む path は拒否する（false ではなくエラー）

---


## ファイル名

### resolveFilename（preserve）

- 元ファイル名を維持しつつサニタイズする
- Unicode は残す。危険文字（`/\\..\0` 等）と先頭末尾の `.` / 空白のみ除去

#### テスト：正常系

- 安全な ASCII 名はそのまま（または拡張子付きのまま）使われる
- 日本語など Unicode を含む名は残る
- 先頭末尾の `.` / 空白が除去される

#### テスト: 異常系

- `/` や `\\` や `..` を含む名はサニタイズ後にパス区切りを含まない
- サニタイズ結果が空になる入力は拒否する
- NUL を含む名は除去または拒否する

### resolveFilename（random）

- 暗号論的ランダム hex に書き換え、拡張子は元ファイルから付与する
- 長さは config デフォルト 32、呼び出しで上書き可

#### テスト：正常系

- 生成名は hex のみ（拡張子を除く）で、デフォルト長 32 になる
- 元の拡張子が付与される（例: `.png`）
- 長さオーバーライドが反映される

#### テスト: 異常系

- 長さ 0 以下の指定は拒否する
- 極端に大きな長さ指定は拒否する（上限は実装で固定しテストに明記）

---


## path / shardian

### buildStoragePath

- 最終ファイル名に shardian を適用し、ストレージ相対 path を得る
- DirLetterCount / DirNestDepth は Config の値を使う

#### テスト：正常系

- 既知のファイル名と幅・深さで、shardian の期待 path と一致する
- 返却 path はストレージルートからの相対（実装の先頭スラッシュ有無は Config / shardian オプションに合わせ一貫させる）
- 同一入力で決定的な path になる（random ファイル名自体は Store 側で変わる）

#### テスト: 異常系

- 空のファイル名は拒否する
- shardian が拒否する名前（拡張子のみ等）はエラーになる

---


## 対象外（本版）

- DB 永続化、`id` 採番、重複 hash 抑止、actor 永続化
- 画像の width/height / variants / パイプライン
- S3 Adapter・署名付き URL
- SVG サニタイズ、PDF サムネ
- HTTP Request 全体の multipart パース必須化

----

以上

