# txt2img リファレンス

テキストプロンプトから画像を生成する。

## パラメータ収集順序

すでに判明している項目はスキップし、指定がない場合は `sdctl` のデフォルト値を使う。
リポジトリ内に `params.yaml` と `prompt_XX_Y.yaml` があれば `--params` / `--prompt` を優先する。
`params.yaml` に `batch_size` がない場合はユーザーが別指定しない限り `--batch-size 2` を付ける。

1. **プロンプト** — 英語推奨
2. **ネガティブプロンプト** — 省略可
3. **設定ファイル** — `--params params.yaml` / `--prompt prompt.yaml`（あれば優先）
4. **画像サイズ** — デフォルト: `512x512`
5. **ステップ数** — デフォルト: `20`
6. **CFG scale** — デフォルト: `7`
7. **sampler** — デフォルト: `Euler a`（一覧は `sdctl samplers list`）
8. **scheduler** — 省略可（一覧は `sdctl schedulers list`）
9. **model** — 切り替えが必要な場合のみ `--model`（値は `sdctl models list` と完全一致）
10. **VAE / text encoder** — モデルに必要な場合のみ（一覧は `sdctl modules`。`config.md` 参照）
11. **Hires. fix** — 生成と同時にアップスケールする場合のみ有効にする（下記参照）
12. **seed** — デフォルト: `-1`（ランダム）
13. **batch** — 標準: `--batch-size 2 --batch-count 1`
14. **出力先** — ファイルパスを指定（`config.md` の出力命名ルール参照）

### Hires. fix フラグ

`--hires-fix` を有効にした場合のみ以下を収集する：

| フラグ | デフォルト | 説明 |
|---|---|---|
| `--hr-scale` | `1.25` | アップスケール倍率 |
| `--hr-upscaler` | `Latent (nearest)` | アップスケーラー名（一覧は `sdctl upscalers`） |
| `--hr-steps` | `0`（=`--steps` と同じ） | セカンドパスのステップ数 |
| `--hr-denoise` | `0.30` | セカンドパスのデノイジング強度 |

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
  --vae "<vae>" \
  --text-encoder "<text_encoder>" \
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

設定ファイルを使う場合：

```bash
sdctl txt2img --params params.yaml --prompt prompt.yaml -o <output_file>
sdctl txt2img "override prompt" --params params.yaml -o <output_file>
```

省略値を使うフラグ・未指定のオプションはコマンドから省く。
