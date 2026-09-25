# median roadmap

SemVer（`docs/charter/versioning-rule.md`）に従う。`v0.n.0` は正式リリース前のため、MINOR 更新でも破壊的変更を許容する。

## マイルストーン

| 版 | 内容 |
| --- | --- |
| `v0.1.0` | Go スキャフォールドのみ。`packages/go` 配下にディレクトリを用意し、ファイルは `.gitkeep` のみ |
| `v0.2.0` | 実行可能な Go 版の完成。第2弾 MVP 範囲（コア + Local FS + DB + 画像圧縮/リサイズ/サムネ）を TDD で実装し、テストおよび CI（Go path）を含む |
| （以降） | SVG サニタイズ、PDF サムネ、S3 互換 Adapter、TS（`packages/js`）、PHP（`packages/php`）を順次 |

## 開発順

1. Go（`packages/go`）
2. TypeScript（`packages/js`）
3. PHP（`packages/php`）

CI/CD は他の `-an` 系（shardian / crudian）と同じ振る舞いとし、docs のみの変更で冗長な CI を起動しない。

----

以上
