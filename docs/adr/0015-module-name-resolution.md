# 0015. モジュール指定でモデル名とフルパスの両方を受け付ける

- Date: 2026-06-12
- Status: Accepted

## Context

`forge_additional_modules` に渡す値は WebUI の絶対パス
（例: `/mnt/data/sd-webui/models/VAE/anima/qwen_image_vae.safetensors`）だが、
ユーザーがフルパスを毎回入力するのは実用的でない。
`/sdapi/v1/sd-modules` はモデル名（`model_name`）とフルパス（`filename`）の対応表を返す。

## Decision

`--vae`・`--text-encoder` フラグおよび params.yaml の `override_settings.forge_additional_modules`
に指定された値を生成直前に解決する。

- 値が `/` で始まる場合：フルパスとしてそのまま使用する。
- それ以外の場合：`/sdapi/v1/sd-modules` から `model_name` が一致するエントリを探し、
  `filename` に置換する。
- 一致するエントリが見つからない場合：元の値をそのまま使用する（API 側の判断に委ねる）。

解決処理は `overrideSettings` が non-nil のときのみ API を呼び出す。

## Consequences

- `sdctl modules` で名前を確認してそのまま `--vae` に渡せる。
- フルパスも引き続き動作するため既存の params.yaml は変更不要。
- 生成ごとに `ListSDModules` の API 呼び出しが 1 回増える（軽微）。
- モデル名が重複している場合は先にヒットしたエントリが使われる。
