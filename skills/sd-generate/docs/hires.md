# hires リファレンス

既存画像にlatentアップスケールを適用する。入力画像のサイズが `--scale` 倍に自動計算される（`--width` / `--height` は使わない）。

## パラメータ収集順序

すでに判明している項目はスキップし、指定がない場合は `sdctl` のデフォルト値を使う。

1. **入力画像パス** — アップスケール対象の画像ファイルパス（必須）
2. **プロンプト** — 再サンプリング時のプロンプト。生成時と同じものを推奨。
3. **スケール倍率** — `--scale`。デフォルト: `1.25`
4. **ステップ数** — `--steps`。デフォルト: `20`
5. **デノイジング強度** — `--denoise`。デフォルト: `0.30`
6. **アップスケーラー** — `--upscaler`。デフォルト: `Latent (nearest)`（一覧は `sdctl upscalers`）
7. **model / VAE / text encoder** — 生成時と同じモデルを使う場合のみ（`config.md` 参照）
8. **設定ファイル** — `--params params.yaml` / `--prompt prompt.yaml`
9. **出力先** — ファイルパスを指定（`config.md` の出力命名ルール参照）

## コマンド

```bash
sdctl hires <input_image> "<prompt>" \
  --scale <scale> \
  --steps <steps> \
  --denoise <denoise> \
  --upscaler "<upscaler>" \
  --model "<model_name>" \
  --vae "<vae>" \
  --text-encoder "<text_encoder>" \
  -o <output_file>
```

設定ファイルを使う場合：

```bash
sdctl hires --params params.yaml --prompt prompt.yaml <input_image> -o <output_file>
sdctl hires "override prompt" --params params.yaml <input_image> -o <output_file>
```

## 多段アップスケールワークフロー（Anima Latent Upscale）

高品質な仕上がりが必要な場合、段階的にアップスケールする：

```bash
sdctl txt2img "anime girl" --steps 45 -o base.png
sdctl hires base.png "anime girl" --scale 1.25 --steps 35 --denoise 0.32 -o hires1.png
sdctl hires hires1.png "anime girl" --scale 1.15 --steps 30 --denoise 0.34 -o final.png
```

出力ファイルには段階が分かる名前を付ける（例: `base.png` → `hires1.png` → `final.png`）。
