# 0003. 対象 Web UI を AUTOMATIC1111 に絞る

- Date: 2026-06-10
- Status: Accepted

## Context

Stable Diffusion の Web UI 実装は複数存在する（AUTOMATIC1111、ComfyUI 等）。
まず明確なターゲットを定め、API 設計を具体化する必要がある。

## Decision

対象 Web UI を [AUTOMATIC1111 stable-diffusion-webui](https://github.com/AUTOMATIC1111/stable-diffusion-webui) に絞る。

## Consequences

- 最も普及した実装であり、ユーザー数・ドキュメント・コミュニティが充実している
- `/sdapi/v1/` の REST API が安定しており、実装の見通しが立てやすい
- ComfyUI などのワークフローベースの UI には対応しない（別途対応する場合は新規 ADR で決定）
