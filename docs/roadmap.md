# median roadmap

SemVer（`docs/charter/versioning-rule.md`）に従う。`v0.n.0` は正式リリース前のため、MINOR 更新でも破壊的変更を許容する。

詳細仕様の正本は各 `plans/`（実装後は `specs/`）。本ファイルは版一覧と短い受け入れのハブに限る。

## マイルストーン

| 版 | 状態 | 短い受け入れ | 詳細 |
| --- | --- | --- | --- |
| `v0.1.0` | 計画中 | Go スキャフォールドのみ（ディレクトリ＋`.gitkeep`） | [plans/v0.1.0/](./plans/v0.1.0/) |
| `v0.2.0` | 計画中 | 実行可能な Go 完成（コア + Local FS + DB + 画像圧縮/リサイズ/サムネ、TDD・CI） | [plans/v0.2.0/](./plans/v0.2.0/) |
| （以降） | 未割当 | SVG、PDF サムネ、S3（署名付き GET URL）、TS、PHP | [plans/unscheduled/](./plans/unscheduled/) |

## 開発順

1. Go（`packages/go`）
2. TypeScript（`packages/js`）
3. PHP（`packages/php`）

CI/CD は他の `-an` 系（shardian / crudian）と同じ振る舞いとし、docs のみの変更で冗長な CI を起動しない。

----

以上
