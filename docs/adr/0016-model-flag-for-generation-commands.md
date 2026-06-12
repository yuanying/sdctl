# 0016. txt2img / img2img への --model フラグ追加

- Date: 2026-06-12
- Status: Accepted

## Context

`models set` コマンドでモデルを切り替えると、以降の生成コマンド全体に影響する。
生成コマンド単位でモデルを指定したいケースがあるため、`txt2img` / `img2img` に `--model` フラグを追加する。

また、既存の `--vae` / `--text-encoder` は `restore_afterwards: true` で生成後に元の設定へ戻していたが、
切り替えのたびにリロードが発生するオーバーヘッドがあるため、一貫して `restore_afterwards: false` に統一する。

## Decision

- `--model` フラグで生成時のモデルを指定できるようにする
- モデル名は完全一致で解決する（部分一致なし）
- `override_settings` に `sd_model_checkpoint` を設定して切り替える（`--vae` / `--text-encoder` と同じ仕組み）
- `--model` / `--vae` / `--text-encoder` すべてで `override_settings_restore_afterwards: false` とし、生成後も設定を維持する
- `--model` フラグが指定された場合は `models list` の一覧と照合してバリデーションを行い、一致しない場合はエラーを返す

## Consequences

- 生成コマンド単体でモデルを指定できるようになり、`models set` を事前に呼ぶ手間が省ける
- `restore_afterwards: false` のため、いずれかのフラグを指定した生成後はその設定が維持される（ロードのオーバーヘッドを避けるため）
- 連続して同じモデル / VAE / text-encoder を使う場合、初回ロードのみで済む
- 完全一致のみのため、モデル名のタイポ時はバリデーションエラーで検知できる
- モデル切り替えには数秒〜数十秒のロード時間が伴う場合がある
