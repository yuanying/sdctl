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
```

**Environment variable** (takes priority over config file):

```bash
export SDCTL_URL=http://myserver:7860
```

## Usage

### txt2img

```bash
sdctl txt2img "a cute cat on a window sill"
sdctl txt2img "a landscape" --steps 30 --width 768 --height 512 -o ./output/
sdctl txt2img "a portrait" --negative "blurry, low quality" --cfg-scale 8
sdctl txt2img "a cat" --batch-count 4 -o ./output/
sdctl txt2img "a cat" --batch-size 2 --batch-count 3 -o result.png
# → result.00.png, result.01.png, ..., result.05.png

# Using config files
sdctl txt2img --params params.yaml --prompt prompt.yaml
sdctl txt2img "override prompt" --params params.yaml
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

### Config file format

**Parameter file** (`params.yaml`) — generation settings and default negative prompt:

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
batch_size: 1
denoising_strength: 0.75  # img2img only
```

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

### Global flags

```
--config string   config file path (default "~/.config/sdctl/config.yaml")
```

### Common flags (txt2img / img2img)

```
    --params string     generation parameter config file (YAML)
    --prompt string     prompt file (YAML)
-n, --negative string   negative prompt
    --steps int         sampling steps (default 20)
    --width int         image width (default 512)
    --height int        image height (default 512)
    --cfg-scale float   CFG scale (default 7)
    --sampler string    sampler name (default "Euler a")
    --seed int          seed, -1 for random (default -1)
    --batch-count int   number of times to run generation (default 1)
    --batch-size int    number of images per batch (default 1)
-o, --output string     output file or directory (default: current directory)
```

> **Note:** When generating multiple images (`--batch-count > 1` or `--batch-size > 1`):
> - If `--output` is a **file path** (e.g. `result.png`), files are saved as `result.0.png`, `result.1.png`, … (zero-padded when 10+ images).
> - If `--output` is a **directory** or omitted, files are saved as `output-TIMESTAMP-N.png` in that directory.
