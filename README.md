# Dinosaur Game

Chrome の恐竜ゲーム（T-Rex Runner)のクローン。Go + [Ebitengine](https://ebitengine.org/) 製で、Web (WebAssembly) 上で動作します。

## 操作

- **ジャンプ**: Space / ↑ / クリック / タップ
- サボテンを飛び越え、低い鳥は飛び越え、高い鳥はくぐる（ジャンプすると当たる）
- 本家準拠の等加速度で徐々に速くなり（約100秒で最高速）、速度に応じて大サボテン → 2連サボテン → 鳥が解禁される
- スコア 700 ごとにナイトモード（色反転）になり、12秒でライトモードに戻る

## 必要環境

ツールバージョンは [mise](https://mise.jdx.dev/) で管理しています。

```sh
mise install
```

## ビルドと実行

```sh
make run        # デスクトップで実行
make build-web  # web/ に WASM をビルド
make serve      # wasmserve で http://localhost:8000 に配信（リロードで再ビルド）
make check      # fmt + vet + test
```

## 構成と移植性

[sago35/koebiten](https://github.com/sago35/koebiten) への移植を想定し、ゲーム本体をエンジン非依存にしています。

```
game/        エンジン非依存のゲームロジックと描画（Ebitengine に依存しない）
  game.go    状態遷移・物理・障害物・当たり判定・スコア
  draw.go    Display インターフェースへの描画
  sprites.go 1bit ビットマップスプライト
  font.go    3x5 ピクセルフォント
main.go      Ebitengine フロントエンド（入力と RGBA フレームバッファのみ）
web/         WASM 配信用ファイル
```

`game` パッケージの前提は次のとおりです。

- 画面は 128x64 の 1bit カラー固定（小型 OLED と同じ）
- 描画は `game.Display` インターフェース（`SetPixel(x, y int)`）のみを使用
- 毎フレーム `Update(jumpPressed bool)` を呼ぶ（60 TPS 想定）
- 乱数は内蔵 xorshift、`float64` と標準ライブラリ最小限のみ使用（TinyGo で動作可能）

### koebiten への移植手順

1. koebiten 側の `Game.Update` でジャンプボタンのエッジ検出をして `game.Update` に渡す
2. `Draw` で `game.Display` を実装した薄いラッパー（koebiten の描画 API へ `SetPixel` を転送）を渡す
3. `game` パッケージはそのまま利用する
