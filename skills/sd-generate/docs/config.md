# 共通設定リファレンス

## YAML 設定ファイル

### params.yaml — 生成設定とデフォルトのネガティブプロンプト

```yaml
negative_prompt: "bad quality, blurry, worst quality"
steps: 30
width: 768
height: 768
cfg_scale: 8.0
sampler: "DPM++ 2M"
scheduler: "Karras"
seed: -1
batch_count: 1
batch_size: 2
denoising_strength: 0.75  # img2img / hires のみ
enable_hr: false           # txt2img: Hires. fix を有効にする
hr_scale: 1.25
hr_upscaler: "Latent (nearest)"
hr_second_pass_steps: 25
hr_denoise: 0.30
override_settings:
  sd_model_checkpoint: "SD1_QuinceMixV2"
  forge_additional_modules:
    - "qwen_image_vae.safetensors"   # VAE を先
    - "qwen_3_06b_base.safetensors"  # text encoder を後
```

### prompt.yaml — プロンプト

```yaml
prompt: "a beautiful landscape, golden hour, cinematic"
negative_prompt: "ugly, distorted"  # params.yaml の値を上書きする
```

CLI フラグは YAML より優先される。プロンプト引数を指定した場合は `prompt.yaml` の `prompt` を上書きする。

## model / VAE / text encoder の指定

CLI での指定：

```bash
sdctl txt2img "anime girl" \
  --model animagineXLV31_v31 \
  --vae qwen_image_vae.safetensors \
  --text-encoder qwen_3_06b_base.safetensors
```

`params.yaml` での指定：

```yaml
override_settings:
  sd_model_checkpoint: "animagineXLV31_v31"
  forge_additional_modules:
    - "qwen_image_vae.safetensors"
    - "qwen_3_06b_base.safetensors"
```

`params.yaml` には `model:` / `vae:` / `text_encoder:` キーは書かない。必ず `override_settings` 配下に書く。
VAE / text encoder は module name と full path のどちらでも指定できる。`sdctl modules` で確認した値をそのまま使う。

## 出力ファイル命名

- `-o` にはディレクトリではなくファイルパスを渡す。ユーザーがディレクトリを指定した場合はそのディレクトリ配下に適切なファイル名を付けて `-o <dir>/<filename>.png` にする。
- シナリオワークスペースで `prompt_XX_Y.yaml` を使う場合は `outputs/image_XX_Y.png` を標準名にする。例: `kutara_aki/01_example/prompt_02_1.yaml` → `-o kutara_aki/01_example/outputs/image_02_1.png`
- プロンプトファイル名がない場合は用途が分かる短い snake_case 名を付ける。例: `portrait_desk.png`, `window_reading.png`
- バッチ生成時もベース名を付ける（例: `-o result.png` → `result.0001.png`, `result.0002.png`, ...）
- hires の多段アップスケールでは段階が分かる名前を付ける（例: `base.png` → `hires1.png` → `final.png`）
