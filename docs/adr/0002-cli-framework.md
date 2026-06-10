# 0002. CLI フレームワークに cobra を採用する

- Date: 2026-06-10
- Status: Accepted

## Context

sdctl は txt2img / img2img / モデル管理などの複数サブコマンドを持つ CLI ツールである。
サブコマンド構成を整理しやすいフレームワークが必要。

## Decision

CLI フレームワークとして [cobra](https://github.com/spf13/cobra) を採用する。

## Consequences

- サブコマンドの定義・ネストが容易で、将来的な機能追加がしやすい
- kubectl や gh など広く使われるツールの実績があり、UX の一貫性が保ちやすい
- 標準ライブラリのみの構成と比べて外部依存が増える
