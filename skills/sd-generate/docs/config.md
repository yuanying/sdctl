# 共通設定リファレンス

## モデル系統の判別

`sdctl models list` のモデル名はプレフィックスで系統が分かる。**系統によって VAE / text encoder の扱いと推奨パラメータが変わる**ため、生成前に必ず系統を判定する。

| プレフィックス | 系統 | VAE / text encoder | 例 |
|---|---|---|---|
| `anima_` | Anima | **必須** | `anima_anima-base-v1.0` |
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

### SDXL（`IL_` / `Pony_`）・SD1 系 — `--model` のみ

VAE はチェックポイントに内蔵されているため `--vae` / `--text-encoder` は付けない。

```bash
sdctl txt2img "<prompt>" --model IL_illustrij_v4
```

## 系統をまたぐときはモジュールを明示的にクリアする

`override_settings` は `override_settings_restore_afterwards: false` で送信され、生成後も WebUI 側に維持される（ADR 0016）。
そのため **Anima で生成した直後に SDXL モデルを `--model` だけで実行すると、Anima の qwen モジュールが残ったままになる**。
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

| | Anima 系 | SDXL 系（`IL_` / `Pony_`） |
|---|---|---|
| steps | 30〜50 | 25〜35 |
| CFG scale | 4〜5 | 5〜7 |
| sampler | `ER SDE`（柔らかくしたい場合は `Euler a`） | `Euler a` / `DPM++ 2M` |
| scheduler | `simple` | `karras` |
| 解像度 | 総画素 約1MP（1024×1024 相当）。幅・高さとも16の倍数 | 同左（SDXL のネイティブ解像度） |

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

## 既定値（環境変数）

`sdctl txt2img` / `img2img` / `hires` は、フラグを省いたときに次の既定値を使う。優先順位は「フラグ > 環境変数 > config.yaml」。

| 環境変数 | config.yaml のキー | 使われる場面 |
|---|---|---|
| `SDCTL_PARAMS` | `params` | `--params` を省いたときの params ファイル |
| `SDCTL_OUTPUT_DIR` | `output_dir` | `-o` を省いたときの出力ディレクトリ（無ければ作られる） |

- 実行前に `echo $SDCTL_PARAMS $SDCTL_OUTPUT_DIR` で既定値を確認できる。
- 既定の params を使いたくないときは `--params ''` を付ける。別の params を使うときは `--params <file>` を付ける（既定の params とは合成されず、置き換わる）。
- `-o` を省くと `$SDCTL_OUTPUT_DIR/output-<YYYYMMDD-HHMMSS>-<n>.png` に保存される。同じ秒に何度実行しても上書きされない。
- 保存したパスは stdout に 1 行ずつ出る。進み具合のバーは stderr が端末のときだけ stderr に出るので、`path=$(sdctl txt2img "...")` のようにパスだけを受け取れる。

## 出力ファイル命名

- 名前を付けて保存するときは `-o` にディレクトリではなくファイルパスを渡す（`SDCTL_OUTPUT_DIR` があり、名前を気にしない場合は `-o` を省いてよい）。ユーザーがディレクトリを指定した場合はそのディレクトリ配下に適切なファイル名を付けて `-o <dir>/<filename>.png` にする。
- シナリオワークスペースで `prompt_XX_Y.yaml` を使う場合は `outputs/image_XX_Y.png` を標準名にする。例: `kutara_aki/01_example/prompt_02_1.yaml` → `-o kutara_aki/01_example/outputs/image_02_1.png`
- プロンプトファイル名がない場合は用途が分かる短い snake_case 名を付ける。例: `portrait_desk.png`, `window_reading.png`
- バッチ生成時もベース名を付ける（例: `-o result.png` → `result.0001.png`, `result.0002.png`, ...）
- hires の多段アップスケールでは段階が分かる名前を付ける（例: `base.png` → `hires1.png` → `final.png`）
