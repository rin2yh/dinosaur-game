# 開発環境とタスク

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

## タスク

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
