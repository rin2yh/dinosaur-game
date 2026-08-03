# Dinosaur Game

どこかで見たような恐竜ゲーム。Go + [Ebitengine](https://ebitengine.org/) 製で、Web (WebAssembly) 上で動作します。

**遊ぶ**: https://rin2yh.github.io/dinosaur-game/

## 操作

- **ジャンプ**: Space / ↑ / クリック / タップ
- **長押しで高く跳ぶ**。離すと上昇が打ち切られるので、小サボテンはチョン押し、大サボテンや3連は長押し
- サボテンを飛び越え、低い鳥は飛び越え、高い鳥はくぐる（ジャンプすると当たる）

進むほど速くなり、スコアに応じて新しい障害物が解禁されます。詳しくは [ゲーム仕様](docs/gameplay.md) を参照してください。

## 動かす

ツールバージョンは [mise](https://mise.jdx.dev/) で管理しています。

```sh
mise install
mise run run     # デスクトップで実行
mise run serve   # http://localhost:8000 に Web 版を配信
mise run check   # fmt + vet + test + ファームウェアビルド
```

Linux でデスクトップ実行する場合は ALSA のヘッダが必要です。その他のタスクや OS ごとの前提は [開発環境とタスク](docs/development.md) にまとめています。

## 構成

ゲーム本体 (`game/`) はエンジン非依存で、Ebitengine（デスクトップ・WASM）と [koebiten](https://github.com/sago35/koebiten)（マイコン実機）の 2 つのフロントエンドが `cmd/` にあります。詳しくは [構成と移植性](docs/architecture.md) を参照してください。

## ドキュメント

- [ゲーム仕様](docs/gameplay.md) — 操作、速度カーブ、障害物の出現、ナイトモード
- [効果音](docs/sound.md) — サイン波から生成している 3 種類の SE
- [開発環境とタスク](docs/development.md) — 必要環境、mise タスク、GitHub Pages へのデプロイ
- [構成と移植性](docs/architecture.md) — ディレクトリ構成、`game` パッケージの前提、ファームウェア
- [ライセンス](docs/licenses.md) — 依存ライブラリのライセンス

## ライセンス

[MIT License](LICENSE) です。依存ライブラリの内訳は [ライセンス](docs/licenses.md) にあります。
