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

## 効果音

3 種類の SE が鳴ります。音声ファイルは持たず、起動時にサイン波から生成しています。

- **ジャンプ**: 地面を離れたフレームで上がる短いブリップ（空中での押下や、タイトル・ゲームオーバー画面の押下では鳴らない）
- **スコア**: 100 点ごとに 5 度上がる 2 音のチャイム
- **ゲームオーバー**: 下降する長めの音

音量は `cmd/dinosaur-game/sound.go` の `masterVolume`、音色は同ファイルの `effects` で調整できます。

| ビルド | 効果音 |
| --- | --- |
| デスクトップ / Web (WASM) | サイン波を合成した PCM |
| ファームウェア: conf2025badge | ブザー（GPIO1 / PWM0）の矩形波。高さと長さだけ共通 |
| ファームウェア: その他のボード | 鳴らない（#23） |

## 必要環境

ツールバージョンは [mise](https://mise.jdx.dev/) で管理しています。

```sh
mise install
```

デスクトップ実行には音声デバイスが必要です（Ebitengine は音声の初期化失敗をゲームのエラーとして返すため、鳴らせない環境では起動できません）。ビルドに追加パッケージが要るかどうかは OS によります。

| OS | 追加で必要なもの |
| --- | --- |
| macOS | なし（`AudioToolbox.framework` が自動でリンクされる） |
| Windows | なし |
| Linux | ALSA のヘッダ。Ubuntu / Debian なら `sudo apt install libasound2-dev`、RedHat 系なら `sudo dnf install alsa-lib-devel` |

Web (WASM) ビルドとファームウェアビルドはどの OS でも不要です。

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

## 構成と移植性

ゲーム本体はエンジン非依存で、Ebitengine と [sago35/koebiten](https://github.com/sago35/koebiten) の 2 つのフロントエンドがあります。

```
game/            エンジン非依存のゲームロジックと描画（どのエンジンにも依存しない）
cmd/
  dinosaur-game/ Ebitengine フロントエンド（デスクトップと WASM、入力・RGBA フレームバッファ・効果音の再生）
  firmware/      koebiten フロントエンド（マイコン実機向け、TinyGo でビルド）
internal/
  beep/          効果音の PCM 生成（標準ライブラリのみ、ディスプレイ無しでテストできる）
targets/         TinyGo のカスタムターゲット定義（ボードごとに 1 ファイル）
web/             WASM 配信用ファイル
```

フロントエンドはどちらも `cmd/` 配下の独立した main パッケージで、共有するのは `game` パッケージだけです。ボードの選択は koebiten 側がビルドタグで行うため、対応ボードを増やすときに増えるのは基本的に `targets/` の JSON だけです。例外は音で、koebiten の範囲外なので、鳴らすボードはピンを指定する小さなファイルを `cmd/firmware/` に足します。

`game` パッケージの前提は次のとおりです。

- 画面は 128x64 の 1bit カラー固定（小型 OLED と同じ）
- 描画は `game.Display` インターフェース（`SetPixel(x, y int)`）のみを使用
- 毎フレーム `Update(jumpPressed bool)` を呼ぶ（60 TPS 想定）
- 音は鳴らさず、そのフレームで発生した効果音を `Sounds() Sound` のビット集合として返すだけ。鳴らせるフロントエンドが `Update` のあとに読み取る（フレーム単位なので、1 tick に複数フレーム進めるなら毎回読む）
- 乱数は内蔵 xorshift、`float64` と標準ライブラリ最小限のみ使用（TinyGo で動作可能）

### ファームウェア（koebiten フロントエンド）

`cmd/firmware/main.go` が上記の前提をそのまま実装しています。

- 任意のキーの押下エッジをジャンプ入力として `game.Update` に渡す
- `game.Display` を実装した薄いラッパーが koebiten の `Displayer.SetPixel` へ転送する
- koebiten v0.5.0 は消灯状態にクリアして点灯ピクセルを描く規約なので、通常時は黒背景に白で描画し、ナイトモードは白で塗り潰してから黒で描いて反転する
- koebiten は 32ms tick（約 31 TPS）のため、1 tick にゲームを 2 フレーム進めて 60 TPS 前提の速度を維持する
- 効果音は毎フレーム `speaker` インターフェース（`play(game.Sound)`）に渡す。音は koebiten の守備範囲外なので、鳴らすのはこのフロントエンド自身の仕事で、conf2025badge は `tinygo.org/x/drivers/tone` でブザーを鳴らす。ブザーは単音なので、同じフレームに複数鳴るときはゲームオーバー > スコア > ジャンプの順で1つだけ鳴らす

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
