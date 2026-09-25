# median

ファイル操作を抽象化するDDD向けライブラリ

## 主な機能

### ファイル操作

- ファイルのアップロード
- ファイルの削除
- ファイルパスの返却(DB非使用時は、アップロードの返却時のみ)

### オプショナルな機能

- メディア管理DBとの連携
  - 独自のスキーマでマイグレーション可能
  - 既存のメディア管理テーブルとも連携可能
- 画像に対する圧縮をするかしないか
  - オリジナルをそのままアップロード
  - 圧縮する場合、ファイルフォーマットと圧縮率を指定可能
    - この場合、アップロード先にはオリジナルを保存しないことになる（デフォルト。フラグで上書き可能）
- サムネイル生成機能
  - サイズ指定(長辺 or 短辺基準を選択可能。または双方絶対指定)
    - アスペクト比変更時の処理指定(cover / contain / squeeze)
  - サイズ指定は複数可能
  - サイズ指定に対し、個別の圧縮率を設定可能(例えばサイズの大きいものほど圧縮率を高めるなど)
  - ファイルフォーマットを指定可能
  - DB使用時、オリジナルへの自己参照
  - PDFの場合、最初の1ページ目をサムネイルとして生成するか否かを選択可能
- SVGの脆弱性サニタイズ

### その他

- multipart/base64両対応
- MIME_TYPEのallow/denyリスト
- アップロードファイルサイズの上限チェック(byte)
- 同時アップロード数制限(デフォルトで20)
- SHA-256の記録による同ファイルアップロード抑止(DB使用時)
- ファイル名によるシャード階層ディレクトリ保存(b4moss/shardianを使用)
- ファイル名の扱いを切替可能（元名を維持 / ランダム文字列へ書き換え）。config デフォルト＋ API オーバーライド

## 対象ストレージ

Adapter化し、選択できるようにする
後からでも移行できるようにする

- ローカルFS(Linux/FreeBSD/Mac/Windows)
- S3互換ストレージ(SDK使用)
  - 署名付き URL（GET）を返せる
  - 複数のバケット／ルートを使い分けられるようにする

### マルチストレージ解決

- config に複数ストレージを **key** 付きで登録する（各 key に `driver`, `bucket` / `path` 等）
- API は storage key で判定・解決する（未指定時は config の default key）
- DB が保持するのは **ストレージルートからの相対 path のみ**（どの key / bucket かは持たない）
- どの key を渡すかの管理・再解決は client の責務

## 実装の方向性

- コア機能はファイルの保存と削除に絞る(Repository層相当)
- それ以外のオプショナルな機能はService層相当として必要に応じて呼び出せるようにする(サブモジュール化)
- DBがなくても、保存・削除が出来るようにする
- DBを使用する場合は、2種類の方式に対応する
  - 独自テーブルスキーマをマイグレーションする
  - 既存のメディア管理テーブルを利用する
    - その場合、APIが必要とするカラムを持っていて、それとマッピング出来ることが前提となる
  - DB操作は **b4moss/crudian** を用いる（MySQL/Postgres/SQLite/libSQL）
  - crudian の CRUD / DB ハンドルは呼び出し側注入、または config からの内部生成の両方可

## 契約（決定事項）

### 第1弾

#### 1. Upload / Store の入出力

- **入力**: `bytes | stream` に加え、コアで `multipart` / `base64` も受け付ける。オプションで mime / filename 等を付与可能
- **返却**: `{ id?, path, mime, size, hash?, storageKey }` および `variants`（サムネ等の派生一覧）
  - DB非使用時は `id` を持たない場合がある
  - `variants` は派生を生成した場合のみ含まれる
  - `storageKey` は実際に使用したストレージ key（第4弾）
- 正式語彙は第2弾のとおり `Store` を正とし、`Upload` はエイリアスまたは入力アダプタとする

#### 2. Delete / Get のキー

- **DB非使用時**: `path` が正。削除・取得は path で行う
- **DB使用時**: `id` が正。path は派生情報として保持・返却する
  - メタデータ条件検索はスコープ外（CRUD client の責務）
  - id 指定の get / delete は median の責務に含む
- **ストレージ解決**: DB の path とは別に storage key で Adapter を解決する（第4弾）。key は DB に保存しない

#### 3. 加工パイプラインとオリジナル保存

- 圧縮とサムネイル生成は**独立したオプション**
- 圧縮 ON 時のデフォルトは、アップロード先にオリジナルを保存しない
  - 「オリジナル保存」フラグで上書き可能
- 画像**本体**に対して長辺（または短辺・絶対サイズ）の制約を付けた場合は、実装者がオリジナル保存を**明示的に否定した**とみなす
  - その場合、制約適用後のバイナリが保存対象となる（原寸ファイルは書かない）
- 推奨処理順（実装の目安）:
  1. 入力解釈（multipart / base64 / bytes / stream）
  2. MIME 検査・サイズ上限
  3. SVG サニタイズ（対象時）
  4. 本体のリサイズ制約（指定時）
  5. 圧縮（指定時）
  6. サムネイル生成（指定時）
  7. ストレージ保存 +（DB使用時）メタデータ記録

#### 4. DB 最小カラム（既存テーブルマッピングの正）

median が必須とするカラム集合は次のとおり。既存テーブル利用時も、これらとマッピングできることが前提となる。

| カラム | 役割 |
| --- | --- |
| `id` | 主キー（DB使用時の正）。採番方式は第2弾参照 |
| `path` | ストレージ上のパス |
| `mime` | MIMEタイプ |
| `size` | バイトサイズ |
| `hash` | SHA-256（重複抑止に使用） |
| `created_at` | 作成日時 |
| `original_id` | オリジナルへの自己参照。派生（サムネ等）は子行 |
| `created_by` | 作成者（Store 時は `owned_by` と同値） |
| `owned_by` | 所有者（Store 時は `created_by` と同値） |
| `status` | 状態（text） |

### 第2弾（MVP向け）

実装前の一問一答で固定した事項。

| # | 項目 | 決定 |
|---|------|------|
| 1 | Go 初回 MVP の範囲 | コア + Local FS + DB + 画像圧縮/リサイズ/サムネ。SVG・PDF・S3 は後続 |
| 2 | 公開 API 語彙 | 正式は `Store` / `Delete` / `Get`。`Upload` はエイリアスまたは入力アダプタ |
| 3 | SHA-256 重複時 | デフォルトは既存レコード返却（成功）。オプションで拒否に切替可 |
| 4 | MIME 判定 | 呼び出し側の宣言 MIME のみ。allow/deny は config の配列で設定 |
| 5 | multipart / base64 入力境界 | 抽出済み part bytes/stream と base64（または data URL）を主とする。HTTP Request 丸ごとの multipart パースは薄いヘルパとして任意提供 |
| 6 | path / shardian | 最終ファイル名に対し shardian で階層化。幅・深さ等パラメータは config 必須。ファイル名自体は preserve / random（第4弾） |
| 7 | `id` 採番 | `auto increment` / `UUID v4` / `UUID v7` / `ULID` を config で選択可能。**デフォルトは auto increment**（第3弾） |
| 8 | `created_by` / `owned_by` | Store 時に同値で挿入する。未指定時は config の default actor、なければエラー（第3弾） |
| 9 | Delete カスケード | 親削除時に子レコード（`original_id` 参照）と子ファイルもすべて削除。**物理削除**（論理削除は当面スコープ外・第3弾） |
| 10 | Get の返却範囲 | デフォルトはメタデータのみ。オプションでバイナリ（bytes / stream）同梱可 |

### 第3弾（残り決定事項）

実装前の一問一答で固定した事項。

| # | 項目 | 決定 |
|---|------|------|
| 1 | DB＋ストレージの原子性 | ライブラリが補償する（片方成功時は可能な範囲で戻す） |
| 2 | `status` と削除 | `Delete` は物理削除（行＋ファイル）。論理削除はこのライブラリの責務外（当面） |
| 3 | 画像フォーマット（初回） | JPEG / PNG / WebP / GIF。アニメ GIF・APNG も対象 |
| 4 | マイグレーション配置 | リポジトリ直下 `migrations/` を正とする。Go は **goose**。他言語はそれぞれ適切なマイグレーションツールを用いる |
| 5 | 初回ロードマップ版切り | `v0.1.0` = Go スキャフォールドのみ（ディレクトリ＋`.gitkeep`）。`v0.2.0` = 実行可能な Go 完成（TDD・テスト＋CI） |
| 6 | Adapter / bucket 指定 | **第4弾の storage key モデルに置き換え**（呼び出しでは key を渡し、config の driver / bucket / path を解決） |
| 7 | SVG サニタイズ | script / event handler / external 参照を除去。基本図形・style は許可。npm **svgo** のサニタイズ周りを参考（TS はそのまま利用可） |
| 8 | TS パッケージ配置 | `packages/js` |
| 9 | `id` デフォルト戦略 | **auto increment** |
| 10 | actor 未指定時 | config の default actor があればそれを使い、なければエラー |
| 11 | ファイル名サニタイズ | Unicode は残す。危険文字（`/\\..\0` 等）と先頭末尾の `.` / 空白のみ除去（`preserve` 時に適用） |

### 第4弾（追加要件）

#### 要件（固定）

1. **DB 操作は b4moss/crudian を用いる**
2. **ファイル名モード** `preserve`（元名維持・サニタイズ適用）/ `random`（ランダム文字列へ書き換え）を切替可能
   - config でデフォルトを定め、API 引数でオーバーライド可
3. **S3 SDK は署名付き URL を返せる**（対象操作は下記）
4. **マルチストレージは config の key で解決**
   - 各 key に `driver`, `bucket` / `path` 等を持たせる
   - DB はストレージルートからの相対 path のみ保持し、どの key が渡されるかは client が解決する

#### 一問一答で固定した事項

| # | 項目 | 決定 |
|---|------|------|
| 1 | `random` ファイル名 | 暗号論的ランダム hex（JS は `crypto` 相当。他言語も同等）。長さは config + オーバーライド引数で指定可。拡張子は元ファイルから付与 |
| 2 | 署名付き URL の対象 | **GET のみ** |
| 3 | 署名付き URL の有効期限 | config で定める（呼び出しで上書き可）。ライブラリが用意する config デフォルト値は **1時間** |
| 4 | storage key の渡し方 | config に default key を置く。`Store` / `Delete` / `Get` で未指定時はそれを使い、呼び出しで上書き可 |
| 5 | Delete / Get と key | Store と同じ（default key 可）。誤った key による失敗は client 責任 |
| 6 | Store 返却 | 実際に使った **`storageKey` を含める** |
| 7 | crudian 注入境界 | 呼び出し側注入があればそれを使う。なければ config から DB / crudian を内部生成 |

ロードマップの正本は `docs/roadmap.md`。

## デフォルトテーブルスキーマ（DBML）

独自スキーマでマイグレーションする場合のデフォルト定義。実体は `docs/default-schema.dbml` を正とする。

```dbml
Table media {
  id varchar(64) [pk, note: 'DB使用時の正キー。型は採番方式に依存: serial/bigserial または UUID/ULID 文字列。デフォルトは auto increment']
  path varchar(2048) [not null, unique, note: 'ストレージルートからの相対パス。storage key は持たない。preserve時は危険文字と端の . / 空白を除去']
  mime varchar(255) [not null]
  size bigint [not null, note: 'バイトサイズ']
  hash char(64) [note: 'SHA-256 hex。重複抑止に使用']
  original_id varchar(64) [note: 'NULLならオリジナル。派生は親を参照。型は id に合わせる']
  created_by varchar(255) [note: '作成者識別子。Store時は owned_by と同値。未指定時は default actor']
  owned_by varchar(255) [note: '所有者識別子。Store時は created_by と同値。未指定時は default actor']
  status text [not null, default: 'active', note: 'アプリ側の状態。median Delete は物理削除（論理削除は当面スコープ外）']
  created_at timestamptz [not null, default: `now()`]

  indexes {
    hash
    original_id
    (owned_by, status)
    created_at
  }
}

Ref: media.original_id > media.id
```

補足:

- `id` / `original_id` の物理型は採番方式（デフォルト auto increment / UUID v4 / UUID v7 / ULID）に合わせてマイグレーション側で切り替える。上の DBML は論理表現
- `path` はストレージルートからの相対パスのみ。storage key / bucket は DB に持たない（第4弾）
- DB 操作は b4moss/crudian 経由（注入または config 生成）
- マイグレーション SQL 等の正本はリポジトリ直下 `migrations/`（Go は goose）
- `status` は仕様上 **text**（DB Enum にはしない）。値の意味はアプリ側で定義する。median の `Delete` は物理削除であり、論理削除は当面行わない
- サムネイル等の派生は同一テーブルの子行とし、`original_id` で親（オリジナル）を指す
- 親 Delete 時は子レコードおよび子ファイルもカスケード削除する
- 既存テーブルマッピング時は、上表の論理カラムを物理カラム名へ対応づけられればよい

## スコープ外

- DBテーブルからメタデータによる検索(CRUD clientの責務とする)
- 論理削除（`status` 更新による削除。当面。`Delete` は物理削除）
- 動画、音声ファイルの変換・圧縮・リサイズ
  - アップロード・削除以外の責務は負わない
- マルウェア・ウイルス検知
  - 必要であれば別のソリューションを組み合わせてもらう
  - そうでなければ、危険なバイナリはMIMEタイプで弾く

## パッケージ配置

- `packages/go` — Go（開発順: 1番目）
- `packages/js` — TypeScript / bun・Node.js（開発順: 2番目）
- `packages/php` — PHP（開発順: 3番目）
- `migrations/` — スキーママイグレーション正本（言語横断）

## ランタイム

- bun / Node.js 24+
- Go 1.26+
- PHP 8.2+

-----

以上
