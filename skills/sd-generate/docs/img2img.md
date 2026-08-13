# img2img リファレンス

> **共通設定**（モデル系統・YAML形式・出力命名）→ [`config.md`](config.md)

既存画像をプロンプトで変換する。

## パラメータ収集順序

すでに判明している項目はスキップし、指定がない場合は `sdctl` のデフォルト値を使う。
リポジトリ内に `params.yaml` と `prompt_XX_Y.yaml` があれば `--params` / `--prompt` を優先する。
`params.yaml` に `batch_size` がない場合はユーザーが別指定しない限り `--batch-size 2` を付ける。

1. **入力画像パス** — 変換元の画像ファイルパス（必須）
2. **モデル系統** — Anima 系（`anima_`）か SDXL 系（`IL_` / `Pony_`）か。元画像を生成したモデルに合わせるのが基本（`config.md` 参照）
3. **プロンプト** — 英語推奨。Anima 系はタグ構造が独自なので `anima-prompt` スキルを使う
4. **ネガティブプロンプト** — 省略可
5. **設定ファイル** — `--params params.yaml` / `--prompt prompt.yaml`（あれば優先）
6. **デノイジング強度** — `0.0`（原画に近い）〜 `1.0`（大きく変換）。デフォルト: `0.75`
7. **画像サイズ** — `sdctl` のデフォルトは `512x512`。総画素 約1MP（1024×1024 相当）を推奨
8. **ステップ数** — `sdctl` のデフォルトは `20`。推奨は Anima 30〜50 / SDXL 25〜35
9. **CFG scale** — `sdctl` のデフォルトは `7`。推奨は Anima 4〜5 / SDXL 5〜7
10. **sampler** — `sdctl` のデフォルトは `Euler a`。Anima は `ER SDE` を推奨（一覧は `sdctl samplers list`）
11. **scheduler** — 省略可。Anima は `simple`、SDXL は `karras` を推奨（一覧は `sdctl schedulers list`）
12. **model** — `--model`（値は `sdctl models list` と完全一致）
13. **VAE / text encoder** — Anima 系のみ必須。SDXL 系では付けない。SDXL 系に切り替える際は残留モジュールのクリアが要る（`config.md` 参照）
14. **seed** — デフォルト: `-1`（ランダム）
15. **batch** — 標準: `--batch-size 2 --batch-count 1`
16. **出力先** — ファイルパスを指定（`config.md` の出力命名ルール参照）

## コマンド

```bash
sdctl img2img "<prompt>" <input_image> \
  -n "<negative_prompt>" \
  --denoising <denoising_strength> \
  --width <width> --height <height> \
  --steps <steps> \
  --cfg-scale <cfg_scale> \
  --sampler "<sampler>" \
  --scheduler "<scheduler>" \
  --model "<model_name>" \
  --seed <seed> \
  --batch-count <batch_count> \
  --batch-size <batch_size> \
  -o <output_file>
```

Anima 系ではこれに `--vae qwen_image_vae.safetensors --text-encoder qwen_3_06b_base.safetensors` を加える。SDXL 系は `--model` のみ。

設定ファイルを使う場合：

```bash
sdctl img2img --params params.yaml --prompt prompt.yaml <input_image> -o <output_file>
sdctl img2img "override prompt" --params params.yaml <input_image> -o <output_file>
```

省略値を使うフラグ・未指定のオプションはコマンドから省く。
