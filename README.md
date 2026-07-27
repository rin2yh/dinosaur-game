# Dinosaur Game

どこかで見たような恐竜ゲーム。Go + [Ebitengine](https://ebitengine.org/) 製で、Web (WebAssembly) 上で動作します。

## 操作

- **ジャンプ**: Space / ↑ / クリック / タップ
- **長押しで高く跳ぶ**。離すと上昇が打ち切られるので、小サボテンはチョン押し、大サボテンや3連は長押し
- サボテンを飛び越え、低い鳥は飛び越え、高い鳥はくぐる（ジャンプすると当たる）
- 等加速度で徐々に速くなる（開始 1.29 → 最高 2.8、約117秒で最高速）
  - スコアに応じて大サボテン(125) → 2連サボテン(330) → 鳥(450) → 3連サボテン(800) が解禁される
  - 鳥は半分が地面より速く、半分が遅く飛ぶ
  - 障害物の間隔は「その障害物自身の幅 × 速度 + 定数」を 1.0〜1.5 倍したもの。幅の広い敵の後ほど長く空き、速度が上がるほどフレーム換算では詰まる（到達間隔は最短 38 → 28 フレーム）
  - 同じ種類は3連続では出ない
  - 最高速に達したあとは、スコア 2500 に向けて間隔のばらつきが狭まり、難しい敵の比率が上がる
- スコア 700 ごとにナイトモード（色反転）になり、12秒でライトモードに戻る

## 必要環境

ツールバージョンは [mise](https://mise.jdx.dev/) で管理しています。

```sh
mise install
```

## ビルドと実行

```sh
mise run run            # デスクトップで実行
mise run build-web      # web/ に WASM をビルド
mise run serve          # wasmserve で http://localhost:8000 に配信（リロードで再ビルド）
mise run check          # fmt + vet + test + ファームウェアビルド
mise run build-firmware # zero-kb02 向け UF2 を bin/ にビルド（TinyGo）
mise run flash-firmware # zero-kb02 に書き込み
```

タスクの一覧と説明は `mise tasks` で確認できます。

## デプロイ

`main` への push で `.github/workflows/deploy-pages.yml` が WASM をビルドし、`web/` を GitHub Pages に公開します（手動実行も可）。初回のみリポジトリの Settings → Pages → Build and deployment の Source を「GitHub Actions」にしてください。

公開先: https://rin2yh.github.io/dinosaur-game/

## リリース

バージョンは [tagpr](https://github.com/Songmu/tagpr) が管理します。`main` への push で `.github/workflows/release.yml` がリリース PR を作成・更新し、その PR をマージした時点でタグが打たれて GitHub Release が作られます。手で `git tag` を打つ必要はありません（feature branch に打つと squash マージ後にコミットが `main` の履歴から外れ、タグが宙に浮きます）。

- タグは `v0.1.0` 形式。`game.Version`（`game/version.go`）と `CHANGELOG.md` はリリース PR の中で tagpr が書き換えるので、手では触りません
- バンプはリリース PR に貼るラベルで決まります。既定はパッチで、`minor` でマイナー、`major` でメジャー
- 初回のリリース PR には `minor` を貼って `v0.1.0` から始めてください（既定のままだと `v0.0.1` になります）

バンプの基準は次のとおりです。0.x のうちはメジャーを上げず、破壊的変更もマイナーに入れます。

| 区分 | 対象 |
| --- | --- |
| メジャー | `game` パッケージの前提（下記）が変わるとき。このリポジトリの外でフロントエンドを書いている人のコードが壊れる変更 |
| マイナー | プレイして違いが分かる変更。難易度調整、新しい敵、操作の変更、対応ボードの追加 |
| パッチ | プレイヤーから見て何も変わらないもの。バグ修正、リファクタ、README や CI |

1.0 は機能が揃ったかどうかではなく、`game` パッケージの前提をもう動かさないと判断できたときに切ります。当面は 0.x のままです。

`game.Version`（`game/version.go`）はまだどこからも読んでいない定数です。リリース間の `main` では直前のリリース版を指すので、実機に版を出したくなったらこれを読んでください。

初回のみ、リポジトリの Settings → Actions → General → Workflow permissions で「Allow GitHub Actions to create and approve pull requests」を有効にし（無効のままだと tagpr がリリース PR を作れません）、バンプ用の `major` / `minor` ラベルを作成してください。

## 構成と移植性

ゲーム本体はエンジン非依存で、Ebitengine と [sago35/koebiten](https://github.com/sago35/koebiten) の 2 つのフロントエンドがあります。

```
game/            エンジン非依存のゲームロジックと描画（どのエンジンにも依存しない）
  game.go        状態遷移・物理・障害物・当たり判定・スコア
  draw.go        Display インターフェースへの描画
  sprites.go     1bit ビットマップスプライト
  font.go        3x5 ピクセルフォント
cmd/
  dinosaur-game/ Ebitengine フロントエンド（デスクトップと WASM、入力と RGBA フレームバッファのみ）
  firmware/      koebiten フロントエンド（マイコン実機向け、TinyGo でビルド）
targets/         TinyGo のカスタムターゲット定義（ボードごとに 1 ファイル）
web/             WASM 配信用ファイル
```

フロントエンドはどちらも `cmd/` 配下の独立した main パッケージで、共有するのは `game` パッケージだけです。ボードの選択は koebiten 側がビルドタグで行うため、対応ボードを増やすときに増えるのは `targets/` の JSON であって `cmd/` ではありません。

`game` パッケージの前提は次のとおりです。

- 画面は 128x64 の 1bit カラー固定（小型 OLED と同じ）
- 描画は `game.Display` インターフェース（`SetPixel(x, y int)`）のみを使用
- 毎フレーム `Update(jumpPressed bool)` を呼ぶ（60 TPS 想定）
- 乱数は内蔵 xorshift、`float64` と標準ライブラリ最小限のみ使用（TinyGo で動作可能）

### ファームウェア（koebiten フロントエンド）

`cmd/firmware/main.go` が上記の前提をそのまま実装しています。

- 任意のキーの押下エッジをジャンプ入力として `game.Update` に渡す
- `game.Display` を実装した薄いラッパーが koebiten の `Displayer.SetPixel` へ転送する
- koebiten v0.5.0 は消灯状態にクリアして点灯ピクセルを描く規約なので、通常時は黒背景に白で描画し、ナイトモードは白で塗り潰してから黒で描いて反転する
- koebiten は 32ms tick（約 31 TPS）のため、1 tick にゲームを 2 フレーム進めて 60 TPS 前提の速度を維持する

ボードには時計が無く毎回同じ状態で起動するため、乱数シードはランを開始したフレーム（＝プレイヤーの押下タイミング）を `game` 側で混ぜて確保しています。

## ライセンス

[MIT License](LICENSE) です。

依存ライブラリもいずれも許容的なライセンスで、再配布時はそれぞれの著作権表示を含めてください。

| 依存 | ライセンス |
| --- | --- |
| [hajimehoshi/ebiten](https://github.com/hajimehoshi/ebiten) | Apache-2.0 |
| [sago35/koebiten](https://github.com/sago35/koebiten) | MIT |
| [tinygo.org/x/drivers](https://github.com/tinygo-org/drivers) | BSD-3-Clause |

ゲームの見た目と手触りは既存の恐竜ゲームを参考にしていますが、スプライトもロジックもこのリポジトリで書き起こしたもので、元実装のコードやアセットは含んでいません。MIT が及ぶのはこのリポジトリのコードだけです。
