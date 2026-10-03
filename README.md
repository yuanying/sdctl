# sdctl

CLI for [AUTOMATIC1111 Stable Diffusion WebUI](https://github.com/AUTOMATIC1111/stable-diffusion-webui).

## Requirements

- Go 1.21+
- Running AUTOMATIC1111 WebUI instance with API enabled (`--api` flag)

## Installation

```bash
go install github.com/yuanying/sdctl@latest
```

## Configuration

By default, sdctl connects to `http://localhost:7860`.

**Config file** (`~/.config/sdctl/config.yaml`):

```yaml
url: http://localhost:7860
params: /path/to/default-params.yaml  # default --params for txt2img / img2img / hires
output_dir: /path/to/images           # default -o for txt2img / img2img / hires
format: jpeg                          # default image format: png (default) or jpeg
jpeg_quality: 90                      # default JPEG quality, 1-100 (default 90)
```

**Environment variables** (take priority over the config file):

| Variable | Config key | Meaning |
|---|---|---|
| `SDCTL_URL` | `url` | WebUI URL |
| `SDCTL_PARAMS` | `params` | Params file used when `--params` is not given |
| `SDCTL_OUTPUT_DIR` | `output_dir` | Output directory used when `-o` is not given (created if missing) |
| `SDCTL_FORMAT` | `format` | Image format (`png` or `jpeg`) used when neither `--format` nor the `-o` extension decides it |
| `SDCTL_JPEG_QUALITY` | `jpeg_quality` | JPEG quality (1-100) used when `--quality` is not given |

```bash
export SDCTL_URL=http://myserver:7860
export SDCTL_PARAMS=/path/to/default-params.yaml
export SDCTL_OUTPUT_DIR=/path/to/images
sdctl txt2img "a cat"   # → /path/to/images/output-<YYYYMMDD-HHMMSS>-<n>.png
```

Precedence is: command-line flag > environment variable > config file.

- `--params ''` runs without any params file, even if a default is set.
- Setting `SDCTL_PARAMS=`, `SDCTL_OUTPUT_DIR=`, `SDCTL_FORMAT=` or `SDCTL_JPEG_QUALITY=` (empty) ignores the config file value.
- Files in the output directory are named `output-<YYYYMMDD-HHMMSS>-<n>.png` (`.jpg` for JPEG). `<n>` skips names that already exist, so runs in the same second never overwrite each other.

### Output

Saved image paths are printed to stdout, one per line. The progress bar is written to stderr, and only when stderr is a terminal, so the output can be piped safely:

```bash
path=$(sdctl txt2img "a cat")
```

### Image format (PNG / JPEG)

Images are saved as PNG by default. The format is decided in this order:

1. `--format png|jpeg` (`jpg` is accepted too)
2. The extension of the `-o` file name: `.png` → PNG, `.jpg` / `.jpeg` → JPEG (case-insensitive). Not used when `-o` is a directory.
3. `SDCTL_FORMAT`, then `format` in the config file
4. PNG

```bash
sdctl txt2img "a cat" -o cat.jpg                 # JPEG, from the extension
sdctl txt2img "a cat" --format jpeg -o ./output/ # → ./output/output-<YYYYMMDD-HHMMSS>-<n>.jpg
sdctl txt2img "a cat" --format jpeg --quality 80
```

- `--format` that contradicts the `-o` extension (e.g. `--format jpeg -o cat.png`) is an error, reported before anything is generated.
- JPEG quality is `--quality` > `SDCTL_JPEG_QUALITY` > `jpeg_quality` > 90.
- PNG files are written exactly as returned by the WebUI. JPEG files are re-encoded by sdctl from that PNG, so **they do not keep the generation parameters (infotext) embedded in the PNG**. Keep PNG when you need PNG Info later, or for intermediate `hires` stages.

## Usage

### txt2img

```bash
sdctl txt2img "a cute cat on a window sill"
sdctl txt2img "a landscape" --steps 30 --width 768 --height 512 -o ./output/
sdctl txt2img "a portrait" --negative "blurry, low quality" --cfg-scale 8
sdctl txt2img "a cat" --batch-count 4 -o ./output/
sdctl txt2img "a cat" --batch-size 2 --batch-count 3 -o result.png
# → result.0001.png, result.0002.png, ..., result.0006.png

# Specify model checkpoint for this generation
# SDXL-family checkpoints (IL_ / Pony_) have a built-in VAE — no extra modules needed
sdctl txt2img "anime girl" --model IL_illustrij_v4

# VAE / text encoder (required for anima models, which fail with
# "You do not have VAE state dict!" without them)
# model name or full path are both accepted
sdctl txt2img "anime girl" \
  --model anima_anima-base-v1.0 \
  --vae qwen_image_vae.safetensors \
  --text-encoder qwen_3_06b_base.safetensors

# Modules are NOT restored after generation (see ADR 0016), so switching from an
# anima model back to SDXL leaves the anima modules loaded and silently produces
# a black image. Clear them with an empty list in params.yaml:
#   override_settings:
#     sd_model_checkpoint: "IL_illustrij_v4"
#     forge_additional_modules: []
sdctl txt2img "anime girl" --params clear_modules.yaml

# Using config files
sdctl txt2img --params params.yaml --prompt prompt.yaml
sdctl txt2img "override prompt" --params params.yaml

# Hires. fix (latent upscale during generation)
sdctl txt2img "anime girl" --hires-fix --hr-scale 1.25 --hr-steps 25 --hr-denoise 0.30
# Check available upscalers with: sdctl upscalers
```

### img2img

```bash
sdctl img2img "a dog" input.png
sdctl img2img "watercolor style" input.png --denoising 0.6 -o result.png
sdctl img2img "variations" input.png --batch-count 4 -o ./output/
sdctl img2img "variations" input.png --batch-count 4 -o result.png
# → result.0.png, result.1.png, result.2.png, result.3.png

# Using config files
sdctl img2img --params params.yaml --prompt prompt.yaml input.png
sdctl img2img "override prompt" --params params.yaml input.png
```

### hires

```bash
# Apply latent upscale + resampling to an existing image
# Prompt comes first, input image second
# Dimensions are scaled automatically: input × --scale → new width/height
sdctl hires "anime girl" base.png --scale 1.25 --steps 35 --denoise 0.32 -o hires1.png

# Multi-stage high-quality upscale (works for both anima and SDXL models)
sdctl txt2img "anime girl" --steps 45 -o base.png
sdctl hires "anime girl" base.png --scale 1.25 --steps 35 --denoise 0.32 -o hires1.png
sdctl hires "anime girl" hires1.png --scale 1.15 --steps 30 --denoise 0.34 -o final.png

# With model / VAE / text-encoder
sdctl hires "anime girl" base.png \
  --model anima_anima-base-v1.0 \
  --vae qwen_image_vae.safetensors \
  --text-encoder qwen_3_06b_base.safetensors \
  --scale 1.25 --steps 35 --denoise 0.32

# Using config files
sdctl hires --params params.yaml --prompt prompt.yaml input.png
sdctl hires "override prompt" --params params.yaml input.png
```

### Config file format

**Parameter file** (`params.yaml`) — generation settings and default negative prompt:

```yaml
negative_prompt: "bad quality, blurry, worst quality"
steps: 30
width: 768
height: 768
cfg_scale: 8.0
sampler: "DPM++ 2M"
scheduler: "karras"          # lowercase id from `sdctl schedulers list`
seed: -1
batch_count: 1
batch_size: 1
denoising_strength: 0.75  # img2img / hires only
enable_hr: false           # txt2img: enable Hires. fix
hr_scale: 1.25             # Hires. fix upscale factor
hr_upscaler: "Latent (nearest)"  # latent modes are not listed by `sdctl upscalers`
hr_second_pass_steps: 25   # 0 = same as steps
hr_denoise: 0.30           # Hires. fix denoising strength
override_settings:
  sd_model_checkpoint: "anima_anima-base-v1.0"  # model checkpoint (no validation)
  forge_additional_modules:                 # anima models only; use [] for SDXL
    - "qwen_image_vae.safetensors"          # model name or full path
    - "qwen_3_06b_base.safetensors"
alwayson_scripts:            # passed as-is to the API's alwayson_scripts (see below)
  ADetailer:
    args: [true, false, {ad_model: "face_yolov8n.pt"}]
```

`alwayson_scripts` is sent unchanged as the `alwayson_scripts` field of the txt2img, img2img and hires requests, so any always-on script (extension) installed in the WebUI can be configured from params. It has no CLI flag. When it is not written, the request has no `alwayson_scripts`, same as before. Only one params file is read (`--params`, otherwise `$SDCTL_PARAMS`), so `--params` replaces the whole default params, including its `alwayson_scripts`; the two files are not merged.

Example: re-draw faces with [ADetailer](https://github.com/Bing-su/adetailer) (the extension must be installed in the WebUI). The first two args are "enable ADetailer" and "skip ADetailer on img2img"; each following object is one ADetailer unit.

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

The generation info (PNG `parameters`) then contains `ADetailer model: face_yolov8n.pt`.

**Prompt file** (`prompt.yaml`) — positive prompt and optional negative prompt override:

```yaml
prompt: "a beautiful landscape, golden hour, cinematic"
negative_prompt: "ugly, distorted"  # overrides params.yaml default
```

CLI flags always take precedence over file values.

### models

```bash
sdctl models list
sdctl models set SD1_QuinceMixV2
```

### modules

```bash
sdctl modules
```

VAE とテキストエンコーダーの一覧を表示します。`--vae` / `--text-encoder` フラグや `params.yaml` の `override_settings.forge_additional_modules` に指定するファイルパスを確認できます。

### upscalers

```bash
sdctl upscalers
```

利用可能なアップスケーラーの一覧を表示します。`--hr-upscaler`（`txt2img`）や `--upscaler`（`hires`）フラグに指定する名前を確認できます。

Latent 系アップスケーラー（`Latent (nearest)` など）は `/sdapi/v1/latent-upscale-modes` エンドポイントで別途確認できます。

### Global flags

```
--config string   config file path (default "~/.config/sdctl/config.yaml")
```

### Common flags (txt2img / img2img / hires)

```
    --params string        generation parameter config file (YAML) (default: $SDCTL_PARAMS)
    --prompt string        prompt file (YAML)
-n, --negative string      negative prompt
    --steps int            sampling steps (default 20)
    --width int            image width (default 512)  # not used in hires (auto-computed)
    --height int           image height (default 512) # not used in hires (auto-computed)
    --cfg-scale float      CFG scale (default 7)
    --sampler string       sampler name (default "Euler a")
    --seed int             seed, -1 for random (default -1)
    --batch-count int      number of times to run generation (default 1)
    --batch-size int       number of images per batch (default 1)
-o, --output string        output file or directory (default: $SDCTL_OUTPUT_DIR, else current directory)
    --format string        image format: png or jpeg/jpg (default: -o extension, else $SDCTL_FORMAT, else png)
    --quality int          JPEG quality 1-100 (default: $SDCTL_JPEG_QUALITY, else 90)
    --model string         model checkpoint name (must match `sdctl models list` exactly)
    --vae string           VAE model path (sets forge_additional_modules[0])
    --text-encoder string  text encoder model path (sets forge_additional_modules[1])
```

### txt2img Hires. fix flags

```
    --hires-fix            enable Hires. fix
    --hr-scale float       upscale factor (default 1.25)
    --hr-upscaler string   upscaler name (default "Latent (nearest)", see `sdctl upscalers`)
    --hr-steps int         second pass steps (default 0 = same as --steps)
    --hr-denoise float     second pass denoising strength (default 0.30)
```

### hires flags

```
    --scale float          upscale factor applied to input image dimensions (default 1.25)
    --denoise float        denoising strength (default 0.30)
    --upscaler string      upscaler name (default "Latent (nearest)", see `sdctl upscalers`)
```

> **Note:** When `--output` is a file path (e.g. `result.png`):
> - **Single image:** saved as `result.png`. If the file already exists, saved as `result.0001.png` (4-digit zero-padded, expands as needed).
> - **Batch (`--batch-count > 1` or `--batch-size > 1`):** always saved with an index suffix starting from the next available number — e.g. `result.0001.png`, `result.0002.png`, …
>
> If `--output` is a **directory** or omitted, files are saved as `output-TIMESTAMP-N.png` (`.jpg` for JPEG).
