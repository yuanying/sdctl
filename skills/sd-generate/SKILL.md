---
name: sd-generate
description: |
  sdctl CLI を使って Stable Diffusion WebUI (AUTOMATIC1111) で画像生成・変換・生成環境確認を行うスキル。
  トリガー: "sd-generate", "/sd-generate", "画像生成", "stable diffusion", "StableDiffusion", "SD画像", "txt2img", "img2img", "hires", "アップスケール", "モデル一覧", "model", "modules", "vae", "text encoder", "sampler", "scheduler", "upscaler", "params.yaml", "prompt.yaml"
  使用場面: (1) テキストプロンプトから画像を生成したいとき、(2) 既存画像をimg2imgで変換したいとき、(3) 既存画像をlatentアップスケールしたいとき、(4) モデル・サンプラー・スケジューラー・VAE・text encoder・アップスケーラーを確認したいとき、(5) seed・CFG・batch・model・VAE・text encoder・YAML設定ファイルなどsdctl生成パラメータを指定して実行したいとき
---

$ARGUMENTS

## 前提条件

以下がセットアップされていることを確認する。未セットアップの場合はユーザーに案内する。

- **sdctl**: `go install github.com/yuanying/sdctl@latest`（Go 1.21+ 必要）
- **WebUI**: AUTOMATIC1111 が `--api` フラグ付きで起動していること
- **接続先**: デフォルト `http://localhost:7860`。変更する場合は環境変数 `SDCTL_URL` または `--config` フラグを使う。

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

## フェーズ2: パラメータ収集 & コマンド実行

判定したインテントに対応するリファレンスファイルを読み、指示に従ってパラメータを収集してコマンドを実行する。

| インテント | リファレンスファイル |
|---|---|
| txt2img | `skills/sd-generate/docs/txt2img.md` |
| img2img | `skills/sd-generate/docs/img2img.md` |
| hires | `skills/sd-generate/docs/hires.md` |
| models / modules / upscalers / samplers / schedulers | `skills/sd-generate/docs/management.md` |

YAML設定ファイルの形式・出力命名・model/VAE/text-encoder の指定ルールは `skills/sd-generate/docs/config.md` を参照する。

## フェーズ3: 結果報告

コマンド実行後、以下を報告する：

- 生成された画像ファイルのパス
- 使用したパラメータのサマリー
- 管理系コマンドの場合は表示・変更した対象
