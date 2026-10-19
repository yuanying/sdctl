# 共通設定リファレンス

## モデル系統の判別

`sdctl models list` のモデル名はプレフィックスで系統が分かる。**系統によって VAE / text encoder の扱いと推奨パラメータが変わる**ため、生成前に必ず系統を判定する。

| プレフィックス | 系統 | VAE / text encoder | 例 |
|---|---|---|---|
| `anima_` | Anima | **必須** | `anima_anima-base-v1.0` |
| `moss-image_` | MOSS-Image | **必須**（text encoder と VAE の 2 つ） | `moss-image_moss_image_sft_bf16` |
| `IL_` | Illustrious（SDXL） | **不要** | `IL_illustrij_v4` |
| `Pony_` | Pony（SDXL） | **不要** | `Pony_prefectPonyXL_v40` |
| `SD1_` | SD 1.5 | **不要** | `SD1_QuinceMixV2` |

ユーザーが系統を指定していない場合は、どの系統で生成するかを確認する。

### Anima 系 — VAE / text encoder が必須

指定しないと API が 500 を返す（`AssertionError: You do not have VAE state dict!`）。

```bash
sdctl txt2img "<prompt>" \
  --model anima_anima-base-v1.0 \
  --vae qwen_image_vae.safetensors \
  --text-encoder qwen_3_06b_base.safetensors
```

VAE と text encoder は `sdctl modules` の一覧から選ぶ。module name と full path のどちらでも指定できる。

### MOSS-Image 系 — text encoder と VAE の 2 つが必須

text encoder（Qwen3.5-9B）と VAE（FLUX.2 VAE）の**両方**を指定する。片方だけ、あるいは Anima のモジュールが残っていると失敗する（`You do not have Qwen3.5 state dict!` など）。
推奨設定はシフト（`--distilled-cfg-scale 5`）まで含むので、`params.yaml`（後述の MOSS-Image 系の例）を使うのが確実。

```bash
sdctl txt2img "<prompt>" \
  --model moss-image_moss_image_sft_bf16 \
  --vae flux2-vae.safetensors \
  --text-encoder qwen3.5_9b_bf16.safetensors \
  --width 1024 --height 1024 --steps 50 --cfg-scale 3 \
  --distilled-cfg-scale 5 --sampler "Euler" --scheduler "simple"
```

モデルは SFT（`moss-image_moss_image_sft_bf16`）と Base（`moss-image_moss_image_base_bf16`）がある。通常は SFT を使う。

### SDXL（`IL_` / `Pony_`）・SD1 系 — `--model` のみ

VAE はチェックポイントに内蔵されているため `--vae` / `--text-encoder` は付けない。

```bash
sdctl txt2img "<prompt>" --model IL_illustrij_v4
```

## 系統をまたぐときはモジュールを明示的にクリアする

`override_settings` は `override_settings_restore_afterwards: false` で送信され、生成後も WebUI 側に維持される（ADR 0016）。
そのため **Anima で生成した直後に SDXL モデルを `--model` だけで実行すると、Anima の qwen モジュールが残ったままになる**。
MOSS-Image から他の系統に切り替えるときも同じで、MOSS-Image のモジュールが残る。Anima ⇔ MOSS-Image の切り替えでは、新しい系統のモジュールを必ず 2 つとも指定し直す。
このとき API はエラーを返さず、**真っ黒な画像が出力される**（検証済み）。

`--vae` / `--text-encoder` に空文字を渡してもクリアされない。`params.yaml` で空配列を渡す。

```yaml
override_settings:
  sd_model_checkpoint: "IL_illustrij_v4"
  forge_additional_modules: []
```

```bash
sdctl txt2img "<prompt>" --params clear_modules.yaml
```

現在ロードされているモデルとモジュールは次で確認できる。SDXL に切り替えたのにモジュールが残っていたら上記でクリアする。

```bash
curl -s http://localhost:7860/sdapi/v1/options | jq '{sd_model_checkpoint, forge_additional_modules}'
```

## 系統別の推奨生成パラメータ

指定がない場合の出発点。ユーザー指定があればそちらを優先する。

| | Anima 系 | MOSS-Image 系 | SDXL 系（`IL_` / `Pony_`） |
|---|---|---|---|
| steps | 30〜50 | 50 | 25〜35 |
| CFG scale | 4〜5 | 3 | 5〜7 |
| shift（`distilled_cfg_scale`） | 指定しない | 5 | 指定しない |
| sampler | `ER SDE`（柔らかくしたい場合は `Euler a`） | `Euler` | `Euler a` / `DPM++ 2M` |
| scheduler | `simple` | `simple` | `karras` |
| ネガティブプロンプト | 使う | **空にする** | 使う |
| 解像度 | 総画素 約1MP（1024×1024 相当）。幅・高さとも16の倍数 | 1024×1024 / 1456×1456 など | 同左（SDXL のネイティブ解像度） |

MOSS-Image は自然文のプロンプト（英語推奨、最大 4K トークン）で、看板などの文字も描ける。描かせたい文字は `"MOONLIGHT COFFEE"` のように引用符で囲む。1024×1024 / 50 steps で 1 枚およそ 30〜50 秒かかる。

Anima 側の値と、プロンプトのタグ構造は `anima-prompt` スキルに準拠する。Anima 向けプロンプトを組み立てる場合はそちらを使う。

**sampler / scheduler の名前の書き方**（間違えるとバリデーションエラーになる）:

- sampler は `sdctl samplers list` の表示名をそのまま使う（`ER SDE`, `DPM++ 2M`）。`er_sde` のような内部 ID は不可。
- scheduler は `sdctl schedulers list` の**左カラムの小文字 ID** を使う（`karras`, `simple`）。右カラムの表示ラベル（`Karras`）は不可。

## YAML 設定ファイル

### params.yaml — 生成設定とデフォルトのネガティブプロンプト

Anima 系の例：

```yaml
negative_prompt: "worst quality, low quality, blurry, jpeg artifacts"
steps: 35
width: 1024
height: 1024
cfg_scale: 4.5
sampler: "ER SDE"
scheduler: "simple"
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
  sd_model_checkpoint: "anima_anima-base-v1.0"
  forge_additional_modules:
    - "qwen_image_vae.safetensors"   # VAE を先
    - "qwen_3_06b_base.safetensors"  # text encoder を後
```

MOSS-Image 系の例（推奨設定）：

```yaml
negative_prompt: ""          # ネガティブプロンプトは空
steps: 50
width: 1024
height: 1024
cfg_scale: 3
distilled_cfg_scale: 5       # shift。省略すると Forge の既定 3.5 になる
sampler: "Euler"
scheduler: "simple"
seed: -1
batch_count: 1
batch_size: 1
override_settings:
  sd_model_checkpoint: "moss-image_moss_image_sft_bf16"
  forge_additional_modules:
    - "qwen3.5_9b_bf16.safetensors"  # text encoder
    - "flux2-vae.safetensors"        # VAE
  moss_uncond_skip_layers: "22"      # ネガティブ側でスキップする層。"" で無効
```

- `distilled_cfg_scale`（CLI では `--distilled-cfg-scale`）は要求の `distilled_cfg_scale` として送られる。Forge では Flux 系の Distilled CFG、MOSS-Image ではシフトとして使われる。書かなければ送らない。
- `moss_uncond_skip_layers` はネガティブ側（unconditional）の計算でスキップする DiT の層。Forge の既定は `"22"`。`override_settings` は生成後も残る（ADR 0016）ので、前の実行に左右されないよう明示しておく。
- 効いていれば、生成情報（PNG の `parameters`）に `CFG scale: 3.0, Shift: 5.0` と `Module 1: flux2-vae, Module 2: qwen3.5_9b_bf16` が出る。

SDXL 系の例（`forge_additional_modules` は空配列にして前の系統の残留を防ぐ）：

```yaml
negative_prompt: "bad quality, blurry, worst quality"
steps: 30
width: 1024
height: 1024
cfg_scale: 6.0
sampler: "DPM++ 2M"
scheduler: "karras"
seed: -1
batch_count: 1
batch_size: 2
override_settings:
  sd_model_checkpoint: "IL_illustrij_v4"
  forge_additional_modules: []
```

### prompt.yaml — プロンプト

```yaml
prompt: "a beautiful landscape, golden hour, cinematic"
negative_prompt: "ugly, distorted"  # params.yaml の値を上書きする
```

CLI フラグは YAML より優先される。プロンプト引数を指定した場合は `prompt.yaml` の `prompt` を上書きする。

`params.yaml` には `model:` / `vae:` / `text_encoder:` キーは書かない。必ず `override_settings` 配下に書く。

### alwayson_scripts — 拡張（ADetailer など）の指定

`params.yaml` の `alwayson_scripts` は、txt2img・img2img・hires の要求の `alwayson_scripts` にそのまま送られる。CLI フラグは無い。書かなければ要求に含まれない。
読む params は 1 ファイルだけ（`--params`、無ければ `$SDCTL_PARAMS`）なので、`--params` を渡すと既定の params の `alwayson_scripts` も含めて丸ごと置き換わる。

ADetailer で顔を描き直す例（WebUI に ADetailer が入っていること）。args の先頭 2 つは「ADetailer を有効にするか」「img2img では ADetailer を飛ばすか」で、その後ろのオブジェクト 1 つが 1 ユニット。

```yaml
alwayson_scripts:
  ADetailer:
    args:
      - true
      - false
      - ad_model: "face_yolov8n.pt"
        ad_denoising_strength: 0.4
        ad_inpaint_only_masked: true
```

効いていれば、生成情報（PNG の `parameters`）に `ADetailer model: face_yolov8n.pt` が出る。

## 既定値（環境変数）

`sdctl txt2img` / `img2img` / `hires` は、フラグを省いたときに次の既定値を使う。優先順位は「フラグ > 環境変数 > config.yaml」。

| 環境変数 | config.yaml のキー | 使われる場面 |
|---|---|---|
| `SDCTL_PARAMS` | `params` | `--params` を省いたときの params ファイル |
| `SDCTL_OUTPUT_DIR` | `output_dir` | `-o` を省いたときの出力ディレクトリ（無ければ作られる） |
| `SDCTL_FORMAT` | `format` | 画像の形式（`png` / `jpeg`）。`--format` も `-o` の拡張子も無いときに使う。どれも無ければ PNG |
| `SDCTL_JPEG_QUALITY` | `jpeg_quality` | JPEG の品質（1〜100）。`--quality` を省いたときに使う。どれも無ければ 90 |

- 実行前に `echo $SDCTL_PARAMS $SDCTL_OUTPUT_DIR $SDCTL_FORMAT` で既定値を確認できる。
- 既定の params を使いたくないときは `--params ''` を付ける。別の params を使うときは `--params <file>` を付ける（既定の params とは合成されず、置き換わる）。
- `-o` を省くと `$SDCTL_OUTPUT_DIR/output-<YYYYMMDD-HHMMSS>-<n>.png`（JPEG のときは `.jpg`）に保存される。同じ秒に何度実行しても上書きされない。
- 保存したパスは stdout に 1 行ずつ出る。進み具合のバーは stderr が端末のときだけ stderr に出るので、`path=$(sdctl txt2img "...")` のようにパスだけを受け取れる。

## 出力形式（PNG / JPEG）

形式は次の順で決まる。

1. `--format png|jpeg`（`jpg` も可）
2. `-o` のファイル名の拡張子（`.png` → PNG、`.jpg` / `.jpeg` → JPEG。大文字も可）。`-o` がディレクトリなら使わない
3. `SDCTL_FORMAT` → config.yaml の `format`
4. PNG

- `--format` と `-o` の拡張子が食い違う（例: `--format jpeg -o out.png`）とエラーになり、生成は行われない。
- JPEG は WebUI が返した PNG を sdctl が変換して書く。**PNG に埋め込まれた生成パラメータ（infotext）は JPEG には残らない。** 後で PNG Info から設定を読み返したい画像や、`hires` の途中段階の画像は PNG で保存する。
- 品質は `--quality`（1〜100）> `SDCTL_JPEG_QUALITY` > config の `jpeg_quality` > 90。

## 出力ファイル命名

- 名前を付けて保存するときは `-o` にディレクトリではなくファイルパスを渡す（`SDCTL_OUTPUT_DIR` があり、名前を気にしない場合は `-o` を省いてよい）。拡張子が形式を決めるので、`SDCTL_FORMAT=jpeg` の環境で名前を付けるときは `.jpg` にする（`.png` にすると PNG で保存される）。ユーザーがディレクトリを指定した場合はそのディレクトリ配下に適切なファイル名を付けて `-o <dir>/<filename>.png` にする。
- シナリオワークスペースで `prompt_XX_Y.yaml` を使う場合は `outputs/image_XX_Y.png` を標準名にする。例: `kutara_aki/01_example/prompt_02_1.yaml` → `-o kutara_aki/01_example/outputs/image_02_1.png`
- プロンプトファイル名がない場合は用途が分かる短い snake_case 名を付ける。例: `portrait_desk.png`, `window_reading.png`
- バッチ生成時もベース名を付ける（例: `-o result.png` → `result.0001.png`, `result.0002.png`, ...）
- hires の多段アップスケールでは段階が分かる名前を付ける（例: `base.png` → `hires1.png` → `final.png`）
