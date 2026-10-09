---
name: fes-archive
description: 開催が終わったサイトを静的アーカイブとして書き出し、GCS にアップロードして /{アーカイブ名}/ で公開する手順。「アーカイブを作る」「2026-Spring をアーカイブ」「GCS に上げる」「アーカイブを公開・追加する」といった作業で使う。
---

# 静的アーカイブの作成と公開

`app/_cmd/archive/main.go` で Datastore のページ・ファイルから静的サイトを生成し、GCS に上げ、公開サイトの `/{アーカイブ名}/` で配信する。

## 前提

- アーカイブツールは **`app/` で実行する**（Go モジュールが `app/go.mod` のため）。
- ローカル実行時（GCE 外）は `DevelopMode=true` になり、**Datastore エミュレータ（既定 `localhost:8081`、プロジェクト ID `fes`）を読む**。本番 Datastore は直接読まない。
  - 別のホストを使うときは `DATASTORE_EMULATOR_HOST` を設定する。
- 出力先は**カレントディレクトリ直下の `<アーカイブ名>/`**（`app/<アーカイブ名>/`）。`.gitignore`（`app/[0-9][0-9][0-9][0-9]-*/`）と `app/.gcloudignore` の対象なので、コミットにもデプロイにも含まれない。アーカイブ名は `2026-Spring` のように「西暦4桁-」で始める。
- 以前は `app/_cmd/archive/` の下にアーカイブや zip がコミットされ、リポジトリが大きくなっていた（履歴から削除済み）。`app/` 直下と `app/_cmd/archive/` の zip も同じく除外している。

## 手順

### 1. 本番データをローカルに用意する

1. 本番の管理画面 Site Settings の「バックアップ」で zip をダウンロードする（`GET /manage/api/tool/backup`）。
2. Datastore エミュレータを起動する（例: `gcloud beta emulators datastore start --host-port=localhost:8081 --project=fes`）。
3. `app/` で `go run ./_cmd/main.go` を起動し、`http://localhost:8080/manage/` の Site Settings の「リストア」で zip を読み込む。
   - **リストアは全データを上書きする。** 必ずローカルで行い、本番では実行しない。
   - バックアップ対象は Site / File / Variable / Template / Page / HTML（と各 Data）で、**Draft は含まれない**。

### 2. 生成してローカルで確認する

```powershell
cd app
go run ./_cmd/archive/main.go -local 2026-Spring
```

- `http://localhost:3000/2026-Spring/` で表示を確認する。
- **ディレクトリが既にあると、`-local` / `-upload` は生成をスキップして既存の内容を使う。** 作り直すときは先に `app/2026-Spring/` を削除する。
- フラグなしで実行すると「生成 → アップロード確認」になる。この場合、ディレクトリが既にあると `os.Mkdir` がエラーになる。

### 3. GCS にアップロードする

```powershell
gcloud auth application-default login   # 未ログイン時のみ
go run ./_cmd/archive/main.go -upload 2026-Spring
```

- アップロード先は **Datastore の `Site.ArchiveBucket`**（リストアした本番値）。未設定だと `log.Fatal` で終了する（フォールバックは無い）。
- バケット名が表示され `Upload? [y/N]` を聞かれる。**実行はユーザーの確認を取ってから**行う。
- オブジェクトは `gs://<bucket>/<アーカイブ名>/...` に置かれる。

### 4. 公開する

本番の管理画面 Site Settings の「Archive」にアーカイブ名を追加して保存する。

- 保存時に `GCSArchiveRouter.Update()` が呼ばれ、**再デプロイなしで `/{アーカイブ名}/` が有効になる**（起動時に GCS クライアントの初期化に失敗していた場合のみ再起動が必要）。
- 同じ画面で、各アーカイブのストレージクラス・サイズ・ファイル数と、1 日あたりのアクセス上限（`ArchiveDailyLimit`）を確認・設定できる。
- `robots.txt` には登録済みアーカイブが `Disallow:/{name}/` として自動で追加される。

### 5. 後片付け

`app/<アーカイブ名>/` を削除する（Git にもデプロイにも含まれないが、手元に残す必要は無い）。

## 関連ファイル

| ファイル | 役割 |
|---|---|
| `app/_cmd/archive/main.go` | CLI 本体 |
| `app/logic/generate.go` | `CreateStaticSite()` 静的サイト生成 |
| `app/logic/gcs.go` | `UploadToGCS()` |
| `app/handler/internal/archive_gcs.go` | 公開側の配信（`Last-Modified` / 304、`Cache-Control: max-age=86400`、日次アクセス上限） |
| `app/handler/manage/api_site.go` | Site 保存時のルーター更新 |
