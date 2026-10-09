---
name: fes-deploy
description: FES を Google App Engine に本番デプロイする手順。管理画面（manage/）と会場マップ（maps/）のビルド成果物を app/ の embed に反映し、ビルド確認のうえ gcloud app deploy する。「デプロイ」「本番に反映」「リリース」「GAE に上げる」といった作業で使う。
---

# 本番デプロイ

`manage/` と `maps/` はどちらも **ビルド成果物を `app/handler/internal/_assets/` に置き、Go バイナリに embed して配信する**。そのため、フロントエンドを変更したら embed に反映してから `app/` をデプロイする。

## 手順

### 1. 管理画面（manage/）を反映する

`manage/` に変更があるとき。

```powershell
cd app
go generate ./handler/internal
```

`app/_cmd/managebuild` が `npm install`（`node_modules` が無いときのみ）→ `npm run build` → `_assets/manage/` を削除して `manage/dist/` を丸ごとコピーする。

### 2. 会場マップ（maps/）を反映する

`maps/` に変更があるとき。専用ツールは無いので手でコピーする。

```powershell
cd maps
npm run lint     # 警告 0 件が必須
npm run build
Remove-Item -Recurse -Force ..\app\handler\internal\_assets\maps\*
Copy-Item -Recurse dist\* ..\app\handler\internal\_assets\maps\
```

- `_assets/maps/` は `.gitignore` 対象で、リポジトリには入っていない。
- 店舗画像（`maps/public/images/{店舗番号}/`）も Git に入っていない。ビルドする環境に店舗画像があることを確かめる（無いと全店 `noimage.png` になる）。作り方は [maps-update](../maps-update/SKILL.md) の「店舗画像」。
- `//go:embed all:_assets/maps/*` は**ファイルが 1 つも無いとコンパイルエラー**になる。削除だけして止めないこと。

### 3. ビルドを確認する

```powershell
cd app
go build ./...
```

### 4. デプロイする

**実行前にユーザーの確認を取ること**（本番に即時反映される）。

```powershell
cd app
gcloud app deploy --project small-axe-fes --bucket gs://hummingbird-deploy --no-cache app.yaml
```

- `app.yaml` は `runtime: go125`、`main: _cmd`、F1 / 最大 1 インスタンス。
- 本番では `environment.json` の `SESSION_KEY` が空だと起動時にエラーになる（→ [fes-setup](../fes-setup/SKILL.md)）。

### 5. 確認する

公開トップ、`/manage/`（ログインできるか）、`/maps/` を開いて確認する。

## 注意

- `environment.json` は `git update-index --skip-worktree` で手元の値を隠している。デプロイには手元の値がそのまま embed される。
- デプロイ後、コミットしていない変更が本番に出ていないか `git status` で確認しておく。
