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

## modules

VAE と text encoder の一覧を表示する。

```bash
sdctl modules
```

出力にある module name または full path は、`--vae` / `--text-encoder` および `params.yaml` の `override_settings.forge_additional_modules` に指定できる。

## upscalers

利用可能なアップスケーラーの一覧を表示する。

```bash
sdctl upscalers
```

`--hr-upscaler`（txt2img の Hires. fix）や `--upscaler`（hires）に指定する名前を確認できる。

## samplers / schedulers

```bash
sdctl samplers list
sdctl schedulers list
```
