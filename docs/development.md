# 開発環境

## ツール

バージョンは [mise](https://mise.jdx.dev/) で管理しています（`mise.toml`）。

```sh
mise install
```

## OS ごとの前提

デスクトップ実行には音声デバイスが必要です。Ebitengine は音声の初期化失敗をゲームのエラーとして返すため、鳴らせない環境では起動できません。

| OS | 追加で必要なもの |
| --- | --- |
| macOS | なし（`AudioToolbox.framework` が自動でリンクされる） |
| Windows | なし |
| Linux | ALSA のヘッダ。Ubuntu / Debian なら `sudo apt install libasound2-dev`、RedHat 系なら `sudo dnf install alsa-lib-devel` |

Web (WASM) ビルドとファームウェアビルドはどの OS でも不要です。

## タスク

`mise tasks` で一覧と説明が出ます。定義は `mise.toml`。

## デプロイ

`main` への push で WASM をビルドし、`web/` を GitHub Pages に公開します（手動実行も可）。定義は `.github/workflows/deploy-pages.yml`。

初回のみリポジトリの Settings → Pages → Build and deployment の Source を「GitHub Actions」にする必要があります。

公開先: https://rin2yh.github.io/dinosaur-game/
