# 0012. Batch Output Filename Pattern for File Path Specification

- Date: 2026-06-11
- Status: Superseded by [0013](0013-output-filename-dedup-index.md)

## Context

ADR 0009 では `--output` にファイルパスが指定され、かつ生成枚数が2枚以上になる場合はエラーとしていた。しかし、スクリプトや自動化フローでは `--output result.png` のようにベース名だけ指定しつつ複数枚生成したいケースがある。エラーで弾くより、自動的に連番ファイルを生成するほうが利便性が高い。

## Decision

- `--output xxx.png` が指定されており生成枚数が2枚以上の場合、エラーとせずに連番ファイルを生成する。
- ファイル名パターン: `<stem>.<N><ext>`（例: `xxx.0.png`, `xxx.1.png`）
- 番号は **0始まり**、総枚数に応じた **ゼロパディング** を行う（例: 10枚なら `xxx.00.png`〜`xxx.09.png`）。
- 生成枚数が **1枚のとき** は番号サフィックスを付けず `xxx.png` のまま保存する。
- `txt2img` と `img2img` の両コマンドに同様に適用する。
- `--output` がディレクトリの場合の動作は ADR 0009 のまま変更しない。

## Consequences

- ファイルパス指定でバッチ生成が行えるようになり、スクリプトからの利用が容易になる。
- 1枚生成時はこれまでと同じファイル名になるため後方互換性を維持できる。
- `validateOutputForBatch` によるエラーチェックは不要になり、削除または簡略化できる。
