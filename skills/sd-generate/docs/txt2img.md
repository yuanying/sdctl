# txt2img リファレンス

> **共通設定**（モデル系統・YAML形式・出力命名）→ [`config.md`](config.md)

テキストプロンプトから画像を生成する。

## パラメータ収集順序

すでに判明している項目はスキップし、指定がない場合は `sdctl` のデフォルト値を使う。
リポジトリ内に `params.yaml` と `prompt_XX_Y.yaml` があれば `--params` / `--prompt` を優先する。
`params.yaml` に `batch_size` がない場合はユーザーが別指定しない限り `--batch-size 2` を付ける。

1. **モデル系統** — Anima 系（`anima_`）か SDXL 系（`IL_` / `Pony_`）かを最初に決める。以降のデフォルト値と VAE / text encoder の要否がここで決まる（`config.md` 参照）
2. **プロンプト** — 英語推奨。Anima 系はタグ構造が独自なので `anima-prompt` スキルを使う
3. **ネガティブプロンプト** — 省略可
4. **設定ファイル** — `--params params.yaml` / `--prompt prompt.yaml`（あれば優先）
5. **画像サイズ** — `sdctl` のデフォルトは `512x512`。SDXL 系・Anima 系ともに総画素 約1MP（1024×1024 相当）を推奨
6. **ステップ数** — `sdctl` のデフォルトは `20`。推奨は Anima 30〜50 / SDXL 25〜35
7. **CFG scale** — `sdctl` のデフォルトは `7`。推奨は Anima 4〜5 / SDXL 5〜7
8. **sampler** — `sdctl` のデフォルトは `Euler a`。Anima は `ER SDE` を推奨（一覧は `sdctl samplers list`）
9. **scheduler** — 省略可。Anima は `simple`、SDXL は `karras` を推奨（一覧は `sdctl schedulers list`）
10. **model** — `--model`（値は `sdctl models list` と完全一致）
11. **VAE / text encoder** — Anima 系のみ必須。SDXL 系では付けない。SDXL 系に切り替える際は残留モジュールのクリアが要る（`config.md` 参照）
12. **Hires. fix** — 生成と同時にアップスケールする場合のみ有効にする（下記参照）
13. **seed** — デフォルト: `-1`（ランダム）
14. **batch** — 標準: `--batch-size 2 --batch-count 1`
15. **出力先** — ファイルパスを指定（`config.md` の出力命名ルール参照）

### Hires. fix フラグ

`--hires-fix` を有効にした場合のみ以下を収集する：

| フラグ | デフォルト | 説明 |
|---|---|---|
| `--hr-scale` | `1.25` | アップスケール倍率 |
| `--hr-upscaler` | `Latent (nearest)` | アップスケーラー名（下記参照） |
| `--hr-steps` | `0`（=`--steps` と同じ） | セカンドパスのステップ数 |
| `--hr-denoise` | `0.30` | セカンドパスのデノイジング強度 |

latent 系のアップスケーラー名は `Latent`, `Latent (antialiased)`, `Latent (bicubic)`, `Latent (bicubic antialiased)`, `Latent (nearest)`, `Latent (nearest-exact)`。
これらは `sdctl upscalers` の一覧には**含まれない**（`sdctl upscalers` は ESRGAN 系など非 latent のアップスケーラーを表示する）。

## コマンド

```bash
sdctl txt2img "<prompt>" \
  -n "<negative_prompt>" \
  --width <width> --height <height> \
  --steps <steps> \
  --cfg-scale <cfg_scale> \
  --sampler "<sampler>" \
  --scheduler "<scheduler>" \
  --model "<model_name>" \
  --hires-fix \
  --hr-scale <hr_scale> \
  --hr-upscaler "<hr_upscaler>" \
  --hr-steps <hr_steps> \
  --hr-denoise <hr_denoise> \
  --seed <seed> \
  --batch-count <batch_count> \
  --batch-size <batch_size> \
  -o <output_file>
```

Anima 系ではこれに `--vae` / `--text-encoder` を加える：

```bash
sdctl txt2img "<prompt>" \
  --model anima_anima-base-v1.0 \
  --vae qwen_image_vae.safetensors \
  --text-encoder qwen_3_06b_base.safetensors \
  --width 1024 --height 1024 \
  --steps 35 --cfg-scale 4.5 \
  --sampler "ER SDE" --scheduler "simple" \
  -o <output_file>
```

SDXL 系は `--model` のみ：

```bash
sdctl txt2img "<prompt>" \
  --model IL_illustrij_v4 \
  --width 1024 --height 1024 \
  --steps 30 --cfg-scale 6 \
  --sampler "DPM++ 2M" --scheduler "karras" \
  -o <output_file>
```

設定ファイルを使う場合：

```bash
sdctl txt2img --params params.yaml --prompt prompt.yaml -o <output_file>
sdctl txt2img "override prompt" --params params.yaml -o <output_file>
```

省略値を使うフラグ・未指定のオプションはコマンドから省く。
