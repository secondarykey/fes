fes is Festival Edit System.

# 概要

テンプレートを作成して、ツリー状にページを作成していきます。

| ディレクトリ | 内容 |
|---|---|
| `app/` | Go バックエンド（公開サイト・管理 API）。Google App Engine にデプロイ |
| `manage/` | 管理画面（React + Vite + MUI）。ビルドしたものを `app/` に embed して `/manage/` で配信 |
| `maps/` | 会場マップ（React + Vite）。ビルドしたものを `app/` に embed して `/maps/` で配信 |

# 開発環境での起動

ローカル（GCE の外）では自動で開発モードになり、Datastore エミュレータに接続します。
詳しい手順と、ログインできないときの確認事項は [docs/skills/fes-setup/SKILL.md](docs/skills/fes-setup/SKILL.md) にあります。

## 必要なもの

- Go（`app/go.mod` のバージョン）
- Node.js / npm（`manage/` と `maps/` のビルド）
- Google Cloud CLI と Datastore エミュレータ（エミュレータは Java が必要）

  ```
  gcloud components install cloud-datastore-emulator
  ```

## 1. environment.json を用意する

`app/handler/internal/_assets/environment.json` に値を入れます。起動時に環境変数として読み込まれます。

```json
{
  "CLIENT_ID": "xxxx.apps.googleusercontent.com",
  "CLIENT_SECRET": "",
  "SESSION_KEY": "",
  "MANAGERS": "you@example.com"
}
```

| キー | 内容 |
|---|---|
| `CLIENT_ID` | Google の OAuth クライアント ID（ログインに使う）。GCP コンソールで「承認済みの JavaScript 生成元」に `http://localhost:8080` を登録しておく |
| `CLIENT_SECRET` | 使っていない。空でよい |
| `SESSION_KEY` | セッション Cookie の署名鍵。開発時は空でも起動する（再起動のたびにログアウトする）。本番では必須 |
| `MANAGERS` | 最初の管理者のメール（カンマ区切り）。Site Settings で管理者を保存するまでの間だけ使われる |

このファイルは Git の追跡対象です。手元の値をコミットしないように、変更を無視させておきます。

```
git update-index --skip-worktree app/handler/internal/_assets/environment.json
```

テンプレート側を更新したいときだけ `git update-index --no-skip-worktree app/handler/internal/_assets/environment.json` で戻します。

## 2. 管理画面と会場マップをビルドする

`app/` は `manage/` と `maps/` のビルド成果物を embed します。成果物は Git に入っていないため、**クローン直後はこれをしないと `go build` / `go run` が通りません。**

```
cd app
go generate ./...
```

`manage/` と `maps/` でそれぞれ `npm install`（`node_modules` が無いとき）と `npm run deploy`（lint → ビルド → `app/handler/internal/_assets/` へコピー）が実行されます。
片方だけなら `go run ./_cmd/frontbuild -only manage` のように指定します。

## 3. Datastore エミュレータを起動する

```
gcloud beta emulators datastore start --host-port=localhost:8081 --project=fes
```

接続先は既定で `localhost:8081`（環境変数 `DATASTORE_EMULATOR_HOST` で変更可）、プロジェクト ID は `fes` です。
エミュレータのデータは、オプションを付けなければ再起動しても残ります。

## 4. サーバーを起動する

別のターミナルで実行します。

```
cd app
go run ./_cmd/main.go
```

http://localhost:8080 で起動します（環境変数 `PORT` で変更可）。

## 5. ログインして管理者を登録する

1. http://localhost:8080/login を開き、Google でログインする（`MANAGERS` に入れたメールで）。
2. 管理画面（http://localhost:8080/manage/）の Site Settings で、管理者のメールを登録して保存する。
   保存したあとは Site Settings の管理者だけで判定され、`MANAGERS` は使われなくなります。**自分のメールも必ず含めてください。**
3. 必要に応じてルートページなどを設定する。

本番と同じデータで確認したいときは、本番の Site Settings でバックアップを取り、ローカルの Site Settings でリストアします（リストアは全上書きです。本番では実行しないこと）。

## 変更を確かめるとき

| 変更したもの | 確かめ方 |
|---|---|
| `app/`（Go・テンプレート） | `go run ./_cmd/main.go` を起動し直す |
| `manage/` | `cd app; go run ./_cmd/frontbuild -only manage` で反映し、サーバーを起動し直す（embed しているため） |
| `maps/` | `cd maps; npm run dev` で単体で動く（http://localhost:5173/maps/）。サーバーに載せるときは `-only maps` で反映して起動し直す |

`manage/` も `npm run dev`（http://localhost:5173/manage/）で動かせます。API などは `localhost:8080` へプロキシされるので、Go のサーバーも起動しておきます。
ただし、ログインは先に http://localhost:8080 で済ませておく必要があります。また、保存などの GET 以外のリクエストは、管理 API の CSRF 対策（Origin の検証）で 403 になります。画面の見た目を確かめるとき向けです。

各ディレクトリの詳しいことは [app/AGENTS.md](app/AGENTS.md)・[manage/AGENTS.md](manage/AGENTS.md)・[maps/AGENTS.md](maps/AGENTS.md) にあります。

# Implement

- 公開日の更新方法

  指定日付と自動日付
  →  Draft機能における自動日付

- 1.20化

# Issue


- アーティスト機能
- タイムテーブル機能

- Archive機能を一部GUIで可能に

  URLを作成
  ローカルで作成してデプロイ
  ヘッドレスモードにも対応する

- キャッシュ機能

  どの程度キャッシュするか

- APIを再度形成

  - マニュアルにまとめる

- Pagingの見直し

- Form作成部分の見直し

   dialの共通化
   Version の扱い

- 管理用のURLを変更できるようにする

- マークダウン機能

- 非公開データ一括管理

  Datastoreにするか、検索にするか

- ツリーの見直し




## データの扱い

- Site
- Page
- PageData
- Content →見直し
- Children

- Top Deprecated
- Dir Deprecated

- Prev NotImplemented
- Next NotImplemented

## テンプレート

"html"
"eraseBR"
"plane"
"convertDate"
"list"
"mark"
"templateContent"
"variable"

かなり増えた

## 認証の方法

Google でログインし、Site Settings に登録したメールアドレスだけが管理画面を使えます。
設定は「[開発環境での起動](#開発環境での起動)」の 1・5 を参照してください。

## デプロイ

先に `cd app; go generate ./...` で管理画面と会場マップを反映してから、`app/` でデプロイします。
手順は [docs/skills/fes-deploy/SKILL.md](docs/skills/fes-deploy/SKILL.md) を参照してください。

```
gcloud app deploy --project small-axe-fes --bucket gs://hummingbird-deploy --no-cache app.yaml
```
