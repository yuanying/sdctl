# 管理系コマンド リファレンス

## models

モデルの一覧確認または切り替えを行う。

```bash
# 一覧表示
sdctl models list

# モデル切り替え（永続的に WebUI 側を切り替える）
sdctl models set <model_name>
```

生成コマンド単位でモデルを指定する場合は `models set` ではなく `--model <model_name>` を使う。
`--model` の値は `sdctl models list` に出るモデル名と完全一致させる。

モデル名のプレフィックス（`anima_` / `IL_` / `Pony_` / `SD1_`）で系統が分かる。系統によって VAE / text encoder の要否と推奨パラメータが変わるため、一覧をユーザーに提示するときは系統ごとにまとめる（`config.md` 参照）。

## modules

VAE と text encoder の一覧を表示する。

```bash
sdctl modules
```

出力にある module name または full path は、`--vae` / `--text-encoder` および `params.yaml` の `override_settings.forge_additional_modules` に指定できる（詳細は `config.md` 参照）。

このモジュールが必要なのは Anima 系（`anima_`）のモデルだけで、SDXL 系（`IL_` / `Pony_`）や SD1 系では指定しない。

## upscalers

利用可能なアップスケーラーの一覧を表示する。

```bash
sdctl upscalers
```

ESRGAN 系など**非 latent** のアップスケーラーが表示される。
`--hr-upscaler`（txt2img の Hires. fix）や `--upscaler`（hires）のデフォルトである `Latent (nearest)` などの latent 系はこの一覧には**含まれない**。latent 系の値は `Latent`, `Latent (antialiased)`, `Latent (bicubic)`, `Latent (bicubic antialiased)`, `Latent (nearest)`, `Latent (nearest-exact)`。

## samplers / schedulers

```bash
sdctl samplers list
sdctl schedulers list
```

- sampler は `sdctl samplers list` の表示名をそのまま指定する（`ER SDE`, `DPM++ 2M`）。
- scheduler は `sdctl schedulers list` の**左カラムの小文字 ID** を指定する（`karras`, `simple`）。右カラムの表示ラベルを渡すとバリデーションエラーになる。
