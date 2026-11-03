# 0017. 認証付きプロキシの後ろの WebUI に Bearer トークンで接続する

- Date: 2026-11-03
- Status: Accepted

## Context

WebUI をログイン付きのリバースプロキシ（GitHub ログインなど）の後ろで公開すると、
sdctl のリクエストはログインページへリダイレクトされ、返ってきた HTML を JSON として
読もうとして `invalid character '<'` で失敗する。プロキシはブラウザを持たない
クライアント向けに `Authorization: Bearer <token>` を受け付けるが、sdctl には
ヘッダーを付ける手段が無かった。

トークンは秘密情報なので、設定ファイルに直接書きたくない。

## Decision

- 設定に `token`（環境変数 `SDCTL_TOKEN`）を足し、空でなければ全リクエストに
  `Authorization: Bearer <token>` を付ける。優先順位は 0005 のとおり環境変数 > 設定ファイル。
  `SDCTL_TOKEN=`（空）は他のキーと同じく設定ファイルの値を無視する。
- 設定ファイルの文字列の値（`url` / `token` / `params` / `output_dir` / `format`）に
  書いた `$VAR` / `${VAR}` は、読み込むときに `os.ExpandEnv` で環境変数に置き換える。
  `token: $SD_WEBUI_TOKEN` と書けば、トークン自体はファイルに残らない。
- 別ホストへのリダイレクトは追わず、`API error 302: redirected to <URL>` として報告する。
  同じホスト内のリダイレクトは従来どおり追い、トークンも付けたままにする。

## Consequences

- ログイン付きのプロキシの後ろにある WebUI を、トークンを渡すだけで使える
- 設定ファイルの値に `$` をそのまま書く手段は無い（エスケープは用意しない）。
  未定義の変数は空文字列になる
- WebUI 自身の `--api-auth`（Basic 認証）には対応しない。必要になったら別に扱う
