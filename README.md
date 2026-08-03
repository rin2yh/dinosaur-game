# Dinosaur Game

どこかで見たような恐竜ゲーム。Go + [Ebitengine](https://ebitengine.org/) 製で、Web (WebAssembly) 上で動作します。

**遊ぶ**: https://rin2yh.github.io/dinosaur-game/

## 操作

- **ジャンプ**: Space / ↑ / クリック / タップ
- **長押しで高く跳ぶ**。離すと上昇が打ち切られるので、小サボテンはチョン押し、大サボテンや3連は長押し
- サボテンを飛び越え、低い鳥は飛び越え、高い鳥はくぐる（ジャンプすると当たる）

## 動かす

```sh
mise install                  # ツールを揃える（mise.toml）
go run ./cmd/dinosaur-game    # デスクトップで実行
mise run serve                # http://localhost:8000 に Web 版を配信
```

タスクの一覧は `mise tasks`。Linux でデスクトップ実行する場合は ALSA のヘッダが要ります → [開発環境](docs/development.md)

## 構成

ゲーム本体 (`game/`) はエンジン非依存で、フロントエンドが `cmd/` に 2 つあります。Ebitengine（デスクトップ・WASM）と [koebiten](https://github.com/sago35/koebiten)（マイコン実機）です。→ [構成と移植性](docs/architecture.md)

## ドキュメント

- [ゲーム仕様](docs/gameplay.md) — 難易度の上げ方、障害物の出現間隔の考え方
- [効果音](docs/sound.md) — サイン波から生成している理由と、生成・再生の担当箇所
- [開発環境](docs/development.md) — OS ごとの前提、タスク、デプロイ
- [構成と移植性](docs/architecture.md) — `game` パッケージの制約と理由、ファームウェア固有の調整

## ライセンス

[MIT License](LICENSE) です。

直接依存は `go.mod` を参照してください。いずれも許容的なライセンスで、再配布時はそれぞれの著作権表示を含める必要があります。

| 依存 | ライセンス |
| --- | --- |
| [hajimehoshi/ebiten](https://github.com/hajimehoshi/ebiten) | Apache-2.0 |
| [sago35/koebiten](https://github.com/sago35/koebiten) | MIT |
| [tinygo.org/x/drivers](https://github.com/tinygo-org/drivers) | BSD-3-Clause |

ゲームの見た目と手触りは既存の恐竜ゲームを参考にしていますが、スプライトもロジックもこのリポジトリで書き起こしたもので、元実装のコードやアセットは含んでいません。MIT が及ぶのはこのリポジトリのコードだけです。
