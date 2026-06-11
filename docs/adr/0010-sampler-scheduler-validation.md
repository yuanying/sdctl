# 0010. sampler・scheduler オプション値の事前検証

- Date: 2026-06-11
- Status: Accepted

## Context

`txt2img` / `img2img` コマンドに `--sampler` と `--scheduler` フラグを追加したが、
無効な値を指定した場合でも生成 API が呼び出されてしまう。
生成処理は時間がかかるため、無効な値による無駄な待ち時間をユーザーに強いることになる。

## Decision

`--sampler` または `--scheduler` が指定された場合、生成 API 呼び出し前に
AUTOMATIC1111 API から有効な値の一覧を取得して照合する。

- 検証対象: `--sampler` と `--scheduler` のみ（数値オプションは対象外）
- タイミング: 生成 API 呼び出し前
- 未指定・デフォルト値の場合: 検証をスキップする
- エラーメッセージ: 無効な値とともに有効な値の一覧を表示する
  - 例: `invalid sampler "foo", available: Euler a, DPM++ 2M, ...`

## Consequences

- 無効な値を指定した場合、長時間の生成処理を開始する前にすぐエラーが返る
- 有効な値が明示されるためユーザーが次のアクションを取りやすい
- 検証のために API へ追加リクエストが発生するが、生成処理と比べると無視できる程度
- sampler/scheduler の一覧は実行時に取得するため、モデルや WebUI バージョンの差異に対応できる
