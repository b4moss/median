# packages/js（TypeScript）

- **状態**: 意図スタブ
- **マイルストーン**: `v0.8.0`
- **前提**: Go 側の契約が `v0.2.0` 以降で固まっていること（段階投入可）
- **関連**: [main.md](../../main.md) / [v0.2.0/](../v0.2.0/) / [v0.6.0/svg-sanitize.md](../v0.6.0/svg-sanitize.md)

## 目的

Go で固めた契約語彙を TypeScript（bun / Node.js）へ移植する。

## ざっくり範囲

- 配置: `packages/js`（crudian と同型）
- Go で実装済みの版に相当する振る舞いを目標（段階投入可）
- SVG サニタイズは svgo を利用可

## やらぬこと（このスタブの範囲）

- 公開 npm 名や exports 条件の最終決定（実装着手時に詳細化）

----

以上
