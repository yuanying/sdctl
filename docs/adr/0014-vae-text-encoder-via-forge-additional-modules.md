# 0014. VAE / text encoder の指定に forge_additional_modules を使用する

- Date: 2026-06-12
- Status: Superseded by [0016]

## Context

anima のような一部モデルは VAE と text encoder を明示的に指定しないと API がエラーを返す。
Forge WebUI の API では `override_settings.forge_additional_modules` にパスの配列を渡すことで
生成リクエスト単位でモジュールを指定できる。
一方 `sd_vae` などの既存フィールドはモデルタイプをまたいだ汎用 VAE 指定であり、
anima/flux などの新しいアーキテクチャには対応していない。

また `override_settings` はデフォルトで `restore_afterwards: true` だが、
Forge では設定が永続化されるケースがあるため、明示的に送信する必要がある。

## Decision

- `txt2img` / `img2img` に `--vae` と `--text-encoder` フラグを追加し、
  値を `override_settings.forge_additional_modules` の配列に格納して送信する。
- params.yaml の `override_settings` フィールドは API の `override_settings` に直接マッピングする
  （独自フィールド `vae:` / `text_encoder:` は設けない）。
- `override_settings` が存在する場合は常に `override_settings_restore_afterwards: true` を送信する。
- `sdctl modules` サブコマンドを追加し、`/sdapi/v1/sd-modules` から利用可能な
  VAE・text encoder の一覧を表示できるようにする。

## Consequences

- WebUI の設定状態が生成後に汚染されない（restore_afterwards による保護）。
- params.yaml の構造が API レスポンスと一致するため、WebUI の設定をそのまま流用しやすい。
- `--vae` / `--text-encoder` はどちらか一方だけ指定することもできる。
- IL (SDXL) モデルに anima 用の text encoder を誤って指定しても Forge 側はエラーを返さず
  そのまま読み込むため、ユーザーが誤指定に気づきにくい（Forge の挙動）。
