# SVG サニタイズ

- **状態**: 方針確定
- **マイルストーン**: `unscheduled`
- **関連**: [v0.2.0/media-pipeline.md](../v0.2.0/media-pipeline.md)

## 目的

アップロードされる SVG から、よくある XSS / 外部参照系の脅威を除去する。

## 方針

- script / event handler / external 参照（xlink 等）を除去
- 基本図形・style は許可
- npm **svgo** のサニタイズ周りを参考にする
- TypeScript 実装では svgo をそのまま利用してよい
- Go / PHP は同等の挙動を目指す（ライブラリ選定は実装時）

## やらぬこと（当面）

- 完全な SVG 仕様準拠の検証
- マルウェア検知全般

----

以上
