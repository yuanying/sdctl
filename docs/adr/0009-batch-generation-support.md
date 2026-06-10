# 0009. Batch Generation Support

- Date: 2026-06-10
- Status: Accepted

## Context

txt2img / img2img コマンドで同一プロンプトから複数枚の画像を生成したいケースがある。AUTOMATIC1111 WebUI API は `batch_count`（生成を繰り返す回数）と `batch_size`（1回の forward pass で並列生成する枚数）の2つのパラメータを持つ。

## Decision

- `txt2img` および `img2img` コマンドの両方に `--batch-count` と `--batch-size` フラグを追加する。
- 複数枚生成時の出力ファイルは連番サフィックスを付与する（例: `output-20260610-120000-1.png`）。
- `--output` にファイルパスが指定されており、かつ生成枚数が2枚以上になる場合はエラーとする。
- `--output` にディレクトリが指定された場合、またはデフォルト（カレントディレクトリ）の場合は連番ファイルとして保存する。

## Consequences

- 複数枚生成時のファイル衝突を防げる。
- `batch_count` と `batch_size` を両方サポートすることで SD API の機能を最大限に活用できる。
- `--output` にファイルパス指定 × 複数枚生成の組み合わせはエラーになるため、スクリプトからの利用時に注意が必要。
