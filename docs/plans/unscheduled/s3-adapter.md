# S3 互換 Adapter

- **状態**: 方針確定
- **マイルストーン**: `unscheduled`
- **関連**: [v0.2.0/storage.md](../v0.2.0/storage.md)

## 目的

S3 互換ストレージへ出し入れし、閲覧用の署名付き URL を返す。

## 方針

- SDK を用いた S3 互換 Adapter
- マルチストレージは既存の **storage key** モデル（config の `driver` / `bucket` / `path`）
- 署名付き URL は **GET のみ**
- 有効期限は config で定める（呼び出しで上書き可）
  - ライブラリが用意する config デフォルト値は **1時間**

## やらぬこと（当面）

- 署名付き PUT / DELETE URL
- ストレージ間の自動移行ツール（Adapter 差し替え可能性の維持は別）

----

以上
