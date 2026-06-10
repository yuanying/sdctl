# 0001. 実装言語に Go を採用する

- Date: 2026-06-10
- Status: Accepted

## Context

sdctl は Stable Diffusion Web UI の API を操作する CLI ツールである。
配布のしやすさと CLI ツールとしての実用性が求められる。

## Decision

実装言語として Go を採用する。

## Consequences

- シングルバイナリとしてビルド・配布できるため、依存関係の管理が不要
- クロスコンパイルが容易で、Linux / macOS / Windows 向けバイナリを生成できる
- Python 製の Stable Diffusion エコシステムのライブラリは直接利用できない（API 経由での連携に限定される）
