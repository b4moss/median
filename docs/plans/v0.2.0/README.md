# v0.2.0 索引 — 実行可能な Go MVP

- **状態**: 仕様詳細
- **マイルストーン**: `v0.2.0`
- **関連**: [roadmap](../../roadmap.md) / [er.dbml](../../er.dbml)

## 受け入れ（短い）

Go で次が TDD（`docs/tests/` 追加）および CI（Go path）付きで動くこと。

- コア: `Store` / `Delete` / `Get`
- Local FS Adapter
- DB 連携（crudian、デフォルトスキーマ）
- 画像圧縮 / リサイズ / サムネイル

**含めない（後続）**: SVG サニタイズ、PDF サムネ、S3、TS、PHP

## ドメイン分割

| 文書 | 内容 |
| --- | --- |
| [core-api.md](./core-api.md) | 公開 API・入出力・Get/Delete |
| [media-pipeline.md](./media-pipeline.md) | 加工パイプライン・画像フォーマット |
| [storage.md](./storage.md) | ストレージ key・ファイル名・path |
| [db-media.md](./db-media.md) | crudian・カラム・削除・原子性・採番 |

----

以上
