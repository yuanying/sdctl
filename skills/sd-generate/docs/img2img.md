# img2img リファレンス

既存画像をプロンプトで変換する。

## パラメータ収集順序

すでに判明している項目はスキップし、指定がない場合は `sdctl` のデフォルト値を使う。
リポジトリ内に `params.yaml` と `prompt_XX_Y.yaml` があれば `--params` / `--prompt` を優先する。
`params.yaml` に `batch_size` がない場合はユーザーが別指定しない限り `--batch-size 2` を付ける。

1. **入力画像パス** — 変換元の画像ファイルパス（必須）
2. **プロンプト** — 英語推奨
3. **ネガティブプロンプト** — 省略可
4. **設定ファイル** — `--params params.yaml` / `--prompt prompt.yaml`（あれば優先）
5. **デノイジング強度** — `0.0`（原画に近い）〜 `1.0`（大きく変換）。デフォルト: `0.75`
6. **画像サイズ** — デフォルト: `512x512`
7. **ステップ数** — デフォルト: `20`
8. **CFG scale** — デフォルト: `7`
9. **sampler** — デフォルト: `Euler a`（一覧は `sdctl samplers list`）
10. **scheduler** — 省略可（一覧は `sdctl schedulers list`）
11. **model** — 切り替えが必要な場合のみ `--model`（値は `sdctl models list` と完全一致）
12. **VAE / text encoder** — モデルに必要な場合のみ（一覧は `sdctl modules`。`config.md` 参照）
13. **seed** — デフォルト: `-1`（ランダム）
14. **batch** — 標準: `--batch-size 2 --batch-count 1`
15. **出力先** — ファイルパスを指定（`config.md` の出力命名ルール参照）

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
  --vae "<vae>" \
  --text-encoder "<text_encoder>" \
  --seed <seed> \
  --batch-count <batch_count> \
  --batch-size <batch_size> \
  -o <output_file>
```

設定ファイルを使う場合：

```bash
sdctl img2img --params params.yaml --prompt prompt.yaml <input_image> -o <output_file>
sdctl img2img "override prompt" --params params.yaml <input_image> -o <output_file>
```

省略値を使うフラグ・未指定のオプションはコマンドから省く。
