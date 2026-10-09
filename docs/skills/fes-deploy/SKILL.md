---
name: fes-deploy
description: FES を Google App Engine に本番デプロイする手順。管理画面（manage/）と会場マップ（maps/）のビルド成果物を app/ の embed に反映し、コンパイルを確かめて gcloud app deploy する。「デプロイ」「本番に反映」「リリース」「GAE に上げる」といった作業で使う。
---

# 本番デプロイ

`manage/` と `maps/` はどちらも **ビルド成果物を `app/handler/internal/_assets/` に置き、Go バイナリに embed して配信する**。そのため、フロントエンドを変更したら embed に反映してから `app/` をデプロイする。

## 手順

### 1. 管理画面（manage/）と会場マップ（maps/）を反映する

```powershell
cd app
go generate ./handler/internal
```

`app/handler/internal/frontend.go` の `go:generate` が `app/_cmd/frontbuild` を呼び、`manage/` と `maps/` のそれぞれで `npm run deploy` を実行する（`node_modules` が無ければ先に `npm install`）。

`npm run deploy` は lint（警告 0 件が必須）→ ビルド → `dist/` を `app/handler/internal/_assets/<manage|maps>/` へコピー（中身を入れ替え）までを行う（各ディレクトリの `tools/deploy.mjs`）。GAE への反映（下の 3.）は含まない。片方だけ反映するときは、そのディレクトリで `npm run deploy` を実行するか、`app/` で `go run ./_cmd/frontbuild -root .. -only maps` のように指定する。

- `_assets/manage/` と `_assets/maps/` は `.gitignore` 対象で、リポジトリには入っていない。
- `node_modules` が古いと lint やビルドが失敗する。`package.json` を更新したら、それぞれのディレクトリで `npm install` する。
- 店舗画像（`maps/public/images/{店舗番号}/`）も Git に入っていない。ビルドする環境に店舗画像があることを確かめる（無いと全店 `noimage.png` になる）。作り方は [maps-update](../maps-update/SKILL.md) の「店舗画像」。
- `//go:embed all:_assets/maps/*` は**ファイルが 1 つも無いとコンパイルエラー**になる。削除だけして止めないこと。

### 2. コンパイルが通るか確かめる（任意）

```powershell
cd app
go vet ./...
```

GAE は**アップロードしたソースをクラウド側でビルドする**ので、手元のビルドはデプロイに使われない。ここは事前の確認で、embed 対象が空でコンパイルエラーになる、などのミスに数秒で気づける（デプロイで失敗すると、アップロードとクラウドのビルドを待つことになる）。`go vet` は型チェックまで行い、ファイルを出力しない。

一方、**1. の反映は必ず手元で行う**。クラウドのビルドは Go だけで npm は動かないので、`_assets/manage/` と `_assets/maps/` はアップロード時点で揃っている必要がある（Git には入っていないが `.gcloudignore` の対象ではないので送られる）。

### 3. デプロイする

**実行前にユーザーの確認を取ること**（本番に即時反映される）。

```powershell
cd app
gcloud app deploy --project small-axe-fes --bucket gs://hummingbird-deploy --no-cache app.yaml
```

- `app.yaml` は `runtime: go125`、`main: _cmd`、F1 / 最大 1 インスタンス。
- 本番では `environment.json` の `SESSION_KEY` が空だと起動時にエラーになる（→ [fes-setup](../fes-setup/SKILL.md)）。

### 4. 確認する

公開トップ、`/manage/`（ログインできるか）、`/maps/` を開いて確認する。

## 注意

- `environment.json` は `git update-index --skip-worktree` で手元の値を隠している。デプロイには手元の値がそのまま embed される。
- デプロイ後、コミットしていない変更が本番に出ていないか `git status` で確認しておく。
