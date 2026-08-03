# 構成と移植性

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

フロントエンドはどちらも `cmd/` 配下の独立した main パッケージで、共有するのは `game` パッケージだけです。ボードの選択は koebiten 側がビルドタグで行うため、対応ボードを増やすときに増えるのは `targets/` の JSON であって `cmd/` ではありません。

## `game` パッケージの前提

- 画面は 128x64 の 1bit カラー固定（小型 OLED と同じ）
- 描画は `game.Display` インターフェース（`SetPixel(x, y int)`）のみを使用
- 毎フレーム `Update(jumpPressed bool)` を呼ぶ（60 TPS 想定）
- 音は鳴らさず、そのフレームで発生した効果音を `Sounds() Sound` のビット集合として返すだけ。鳴らせるフロントエンドが `Update` のあとに読み取る（フレーム単位なので、1 tick に複数フレーム進めるなら毎回読む）
- 乱数は内蔵 xorshift、`float64` と標準ライブラリ最小限のみ使用（TinyGo で動作可能）

## ファームウェア（koebiten フロントエンド）

`cmd/firmware/main.go` が上記の前提をそのまま実装しています。

- 任意のキーの押下エッジをジャンプ入力として `game.Update` に渡す
- `game.Display` を実装した薄いラッパーが koebiten の `Displayer.SetPixel` へ転送する
- koebiten v0.5.0 は消灯状態にクリアして点灯ピクセルを描く規約なので、通常時は黒背景に白で描画し、ナイトモードは白で塗り潰してから黒で描いて反転する
- koebiten は 32ms tick（約 31 TPS）のため、1 tick にゲームを 2 フレーム進めて 60 TPS 前提の速度を維持する
- `Sounds()` は読まない。音は koebiten の守備範囲外で、鳴らすならこのフロントエンドが自分でブザーを叩くことになる（#23）

ボードには時計が無く毎回同じ状態で起動するため、乱数シードはランを開始したフレーム（＝プレイヤーの押下タイミング）を `game` 側で混ぜて確保しています。
