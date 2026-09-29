---
name: sd-generate
description: |
  sdctl CLI を使って Stable Diffusion WebUI (AUTOMATIC1111) で画像生成・変換・生成環境確認を行うスキル。
  トリガー: "sd-generate", "/sd-generate", "画像生成", "stable diffusion", "StableDiffusion", "SD画像", "txt2img", "img2img", "hires", "アップスケール", "モデル一覧", "model", "modules", "anima", "sdxl", "illustrious", "pony", "vae", "text encoder", "sampler", "scheduler", "upscaler", "params.yaml", "prompt.yaml"
  使用場面: (1) テキストプロンプトから画像を生成したいとき、(2) 既存画像をimg2imgで変換したいとき、(3) 既存画像をlatentアップスケールしたいとき、(4) モデル・サンプラー・スケジューラー・VAE・text encoder・アップスケーラーを確認したいとき、(5) seed・CFG・batch・model・VAE・text encoder・YAML設定ファイルなどsdctl生成パラメータを指定して実行したいとき
---

$ARGUMENTS

## 前提条件

以下がセットアップされていることを確認する。未セットアップの場合はユーザーに案内する。

- **sdctl**: `go install github.com/yuanying/sdctl@latest`（Go 1.21+ 必要）
- **WebUI**: AUTOMATIC1111 が `--api` フラグ付きで起動していること
- **接続先**: デフォルト `http://localhost:7860`。変更する場合は環境変数 `SDCTL_URL` または `--config` フラグを使う。
- **既定値の環境変数**: `SDCTL_PARAMS`（`--params` を省いたときの params ファイル）と `SDCTL_OUTPUT_DIR`（`-o` を省いたときの出力ディレクトリ）、`SDCTL_FORMAT`（`png` / `jpeg`。`--format` も `-o` の拡張子も無いときの形式）が設定されていることがある。フラグを付ければフラグが優先される。詳細は `docs/config.md` を参照。

## フェーズ1: インテント判定

`$ARGUMENTS` と会話の文脈から以下のいずれかを判定する：

| インテント | 説明 |
|---|---|
| **txt2img** | テキストプロンプトから画像を生成（デフォルト） |
| **img2img** | 既存画像をプロンプトで変換 |
| **hires** | 既存画像にlatentアップスケールを適用 |
| **models** | モデルの一覧確認または切り替え |
| **modules** | VAE / text encoder の一覧確認 |
| **upscalers** | アップスケーラーの一覧確認 |
| **samplers** | サンプラー一覧の確認 |
| **schedulers** | スケジューラー一覧の確認 |

生成系（txt2img / img2img / hires）の場合は、あわせて**モデル系統**を判定する。系統によって VAE / text encoder の要否と推奨パラメータが変わる。

| モデル名のプレフィックス | 系統 | VAE / text encoder |
|---|---|---|
| `anima_` | Anima | 必須 |
| `IL_` / `Pony_` | SDXL | 不要 |
| `SD1_` | SD 1.5 | 不要 |

指定がなく文脈からも判断できない場合はユーザーに確認する。

## フェーズ2: パラメータ収集 & コマンド実行

判定したインテントに対応するリファレンスファイルを読み、指示に従ってパラメータを収集してコマンドを実行する。

> **必須**: インテント別リファレンスを読む前に **必ず** `docs/config.md` を読むこと。設定ファイル（params.yaml / prompt.yaml）の形式や model/VAE/text-encoder の指定が誤っていると生成が失敗する。とくに Anima 系と SDXL 系を切り替えるときは、前の系統のモジュールが残って**エラーなしで真っ黒な画像が出力される**ため、`config.md` のクリア手順に従うこと。

| インテント | リファレンスファイル |
|---|---|
| **共通（全インテント）** | `docs/config.md` |
| txt2img | `docs/txt2img.md` |
| img2img | `docs/img2img.md` |
| hires | `docs/hires.md` |
| models / modules / upscalers / samplers / schedulers | `docs/management.md` |

## フェーズ3: 結果報告

コマンド実行後、以下を報告する：

- 生成された画像ファイルのパス
- 使用したパラメータのサマリー
- 管理系コマンドの場合は表示・変更した対象
