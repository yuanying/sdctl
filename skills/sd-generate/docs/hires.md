# hires リファレンス

> **共通設定**（モデル系統・YAML形式・出力命名）→ [`config.md`](config.md)

既存画像にlatentアップスケールを適用する。入力画像のサイズが `--scale` 倍に自動計算される（`--width` / `--height` は使わない）。

引数の順序は **プロンプトが先、入力画像が後**（`sdctl hires "<prompt>" <input_image>`）。`--prompt` ファイルを使う場合は入力画像を最後の引数に置く。

## パラメータ収集順序

すでに判明している項目はスキップし、指定がない場合は `sdctl` のデフォルト値を使う。

1. **入力画像パス** — アップスケール対象の画像ファイルパス（必須）
2. **モデル系統** — 元画像を生成したモデルに合わせる。系統が変わると絵柄が変わる（`config.md` 参照）
3. **プロンプト** — 再サンプリング時のプロンプト。生成時と同じものを推奨。
4. **スケール倍率** — `--scale`。デフォルト: `1.25`
5. **ステップ数** — `--steps`。デフォルト: `20`
6. **デノイジング強度** — `--denoise`。デフォルト: `0.30`。`0.30`〜`0.35` が Anima 系・SDXL 系ともに扱いやすい（上げるほど構図が変化する）
7. **アップスケーラー** — `--upscaler`。デフォルト: `Latent (nearest)`（下記参照）
8. **model / VAE / text encoder** — Anima 系は `--vae` / `--text-encoder` が必須。SDXL 系は `--model` のみで、直前が Anima 系だった場合は残留モジュールのクリアが要る（`config.md` 参照）
9. **設定ファイル** — `--params params.yaml` / `--prompt prompt.yaml`
10. **出力先** — ファイルパスを指定（`config.md` の出力命名ルール参照）

latent 系のアップスケーラー名は `Latent`, `Latent (antialiased)`, `Latent (bicubic)`, `Latent (bicubic antialiased)`, `Latent (nearest)`, `Latent (nearest-exact)`。
これらは `sdctl upscalers` の一覧には**含まれない**（`sdctl upscalers` は ESRGAN 系など非 latent のアップスケーラーを表示する）。

## コマンド

```bash
sdctl hires "<prompt>" <input_image> \
  --scale <scale> \
  --steps <steps> \
  --denoise <denoise> \
  --upscaler "<upscaler>" \
  --model "<model_name>" \
  -o <output_file>
```

Anima 系ではこれに `--vae` / `--text-encoder` を加える：

```bash
sdctl hires "<prompt>" base.png \
  --model anima_anima-base-v1.0 \
  --vae qwen_image_vae.safetensors \
  --text-encoder qwen_3_06b_base.safetensors \
  --scale 1.25 --steps 35 --denoise 0.32 \
  -o hires1.png
```

SDXL 系は `--model` のみ：

```bash
sdctl hires "<prompt>" base.png \
  --model IL_illustrij_v4 \
  --scale 1.25 --steps 30 --denoise 0.32 \
  -o hires1.png
```

設定ファイルを使う場合（入力画像は最後の引数）：

```bash
sdctl hires --params params.yaml --prompt prompt.yaml <input_image> -o <output_file>
sdctl hires "override prompt" --params params.yaml <input_image> -o <output_file>
```

## 多段アップスケールワークフロー

高品質な仕上がりが必要な場合、倍率を小さく分けて段階的にアップスケールする。Anima 系・SDXL 系のどちらでも同じ手順を使う。
段ごとに倍率とデノイジング強度をわずかに変えて、ディテールを足しつつ構図の崩れを抑える。

Anima 系：

```bash
sdctl txt2img "<prompt>" --model anima_anima-base-v1.0 \
  --vae qwen_image_vae.safetensors --text-encoder qwen_3_06b_base.safetensors \
  --steps 45 -o base.png
sdctl hires "<prompt>" base.png --scale 1.25 --steps 35 --denoise 0.32 -o hires1.png
sdctl hires "<prompt>" hires1.png --scale 1.15 --steps 30 --denoise 0.34 -o final.png
```

SDXL 系（`--model` のみ。直前が Anima 系なら最初のコマンドでモジュールをクリアする）：

```bash
sdctl txt2img "<prompt>" --model IL_illustrij_v4 --steps 30 -o base.png
sdctl hires "<prompt>" base.png --scale 1.25 --steps 30 --denoise 0.32 -o hires1.png
sdctl hires "<prompt>" hires1.png --scale 1.15 --steps 28 --denoise 0.34 -o final.png
```

2段目以降は `--model` などを再指定しなくても直前の設定が維持される（`restore_afterwards: false`）。
出力ファイルには段階が分かる名前を付ける（例: `base.png` → `hires1.png` → `final.png`）。
