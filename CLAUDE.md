# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FES (Festival Edit System) is a Japanese CMS for managing festival/event websites. The main application lives in `app/`. The `maps/` directory is a separate standalone React application and is developed/deployed independently.

## Build & Run Commands

### Backend (Go — `app/`)

```powershell
# Run locally
go run ./app/_cmd/main.go

# Build binary
go build -o app ./app/_cmd

# Deploy to Google App Engine
gcloud app deploy app/app.yaml
```

### Static Site Archive

アーカイブツール (`app/_cmd/archive/main.go`) は静的サイトを生成し、GCS にアップロードする。

```powershell
# ローカル確認（サーバーを起動して確認）
go run ./app/_cmd/archive/main.go -local 2026-Spring

# GCS にアップロード（y/N 確認あり）
go run ./app/_cmd/archive/main.go -upload 2026-Spring
```

- 第1引数はアーカイブ名（必須）。既存ディレクトリがあっても上書きする。
- `-local`: ローカルサーバーで表示確認。アップロードしない。
- `-upload`: バケット名を表示して y/N 確認後、GCS にアップロード。
- アップロード先バケットは `config.ArchiveBucket`（現在: `"hummingbird-archives"`）で定義。

## Architecture

### Backend layers (`app/`)

| Package | Role |
|---|---|
| `app/handler/` | HTTP handlers for public routes (pages, files, auth) |
| `app/handler/manage/` | Admin console routes |
| `app/handler/internal/` | Cache server, static asset serving, archive |
| `app/datastore/` | DAO layer — one file per Datastore kind |
| `app/logic/` | Business logic (static site generation, HTML rendering) |
| `app/api/` | Template helpers (`html/template` FuncMap, content conversion) |
| `app/config/` | Functional-option pattern for app configuration |

Routes are registered in `handler.Register()` (public) and `manage.Register()` (admin) via Gorilla Mux.

### Data model (Google Cloud Datastore)

Core kinds: `Page` / `PageData`, `HTML` (cached render), `File` / `FileData`, `Template`, `Variable`, `Site`, `Draft`, `Meta`.

Pages form a parent-child tree. `HTML` stores pre-rendered output to avoid re-rendering on each request.

### Embedded assets

`app/handler/internal/_assets/` is embedded with `//go:embed`. It contains:
- `environment.json` — OAuth2 credentials (CLIENT_ID, CLIENT_SECRET)
- `archives/` — static archive zips（ZIP 形式の旧アーカイブ）
- `manage-v2/` — V2 管理画面の Vite ビルド成果物

### GCS アーカイブ配信

`app/handler/internal/archive_gcs.go` が GCS バケットからアーカイブを配信する。

- 配信対象アーカイブ名は `config.ArchiveNames` (string スライス) に列挙する。**起動時に GCS API を呼ばない**設計なので起動遅延は発生しない。
- 各アーカイブは `/{name}/` パスで配信される（例: `/2026-Spring/`）。
- キャッシュ: `storage.Reader.Attrs.LastModified` を `Last-Modified` ヘッダに設定し、`If-Modified-Since` が一致すれば 304 を返す。GCS API 呼び出しは 1 リクエストあたり 1 回。
- `Cache-Control: public, max-age=86400` を付与。
- アーカイブ名の追加は `config/option.go` の `ArchiveNames` のみ変更すればよい。

### テンプレートキャッシュの注意点

`logic/html.go` の `PutHTMLs()` はページ HTML を一括生成する。テンプレートの参照先ページを更新した後に HTML 公開を行っても古い内容が反映される問題があったため、`PutHTMLs()` の先頭で `startTemplateCache()` を呼び出してキャッシュをクリアするようにしている。パッケージレベル変数 `cacheTemplateData` が原因だった。

### Authentication

Google OAuth2 via credentials embedded in `environment.json`. Sessions managed with Gorilla sessions. JWT used for token handling.

## manage-v2/ — 管理画面 SPA (開発中)

既存の `/manage/` をそのまま残しつつ `/manage/v2/` で新しい SPA ベースの管理画面を並行稼働させる構成。

### コマンド

```powershell
cd manage-v2
npm install
npm run dev       # 開発サーバー (Vite proxy 不要、Go サーバー経由で動作確認)
npm run build     # manage-v2/dist/ に出力

# ビルド後、Go embed 用にコピーしてから go build する
Copy-Item -Recurse -Force manage-v2\dist\* app\handler\internal\_assets\manage-v2\
cd app; go build ./...
```

### 構成

| ファイル/ディレクトリ | 役割 |
|---|---|
| `src/App.jsx` | React Router セットアップ (basename: `/manage/v2`) |
| `src/components/Layout.jsx` | サイドバー + Outlet |
| `src/pages/PageList.jsx` | ページツリー一覧 (遅延展開) |
| `src/pages/PageEdit.jsx` | ページ編集フォーム・公開/非公開操作 |
| `src/api/page.js` | API クライアント (fetch wrapper) |

### Go 側の対応ファイル

| ファイル | 役割 |
|---|---|
| `app/handler/internal/manage_v2.go` | `_assets/manage-v2/` の embed + 静的配信 + SPA catch-all |
| `app/handler/manage/v2.go` | v2 ルート登録 (`/manage/v2/api/` + `/manage/v2/` catch-all) |
| `app/handler/manage/v2_page.go` | JSON API ハンドラ (Page CRUD・公開・非公開) |

### API エンドポイント (すべて `/manage/v2/api` 配下)

```
GET    /page/                   ルートページ + 子一覧
GET    /page/new/{parentKey}    新規ページスキャフォールド (キー生成)
GET    /page/children/{key}     子ページ一覧
GET    /page/{key}              ページ詳細 + テンプレート一覧
POST   /page/{key}              ページ保存 (新規・更新兼用)
DELETE /page/{key}              ページ削除
POST   /html/publish/{key}      HTML 公開
POST   /html/unpublish/{key}    HTML 非公開
```

### 認証・切り替え方針

- 認証は既存の Gorilla セッションをそのまま使用。`/manage/` 配下全体を `ManageHandler` が保護。
- `/manage/v2/api/` への未認証アクセスはリダイレクトではなく JSON 401 を返す。
- 将来 v2 に完全移行する際は `/manage/` を `/manage/v2/` へリダイレクトするだけでよい。

## maps/ — 独立した React アプリ

`maps/` は Go バックエンドとは独立した単体の React SPA。フェスティバルの会場マップをブラウザ上に表示し、来場者の現在地を GPS で重ねて表示する。

### コマンド

```powershell
cd maps
npm install
npm run dev       # 開発サーバー (base path: /maps/)
npm run build     # maps/dist/ に出力
npm run lint      # ESLint (警告0件が必須)
npm run preview   # 本番ビルドのプレビュー
```

### 構成

| ファイル | 役割 |
|---|---|
| `src/App.jsx` | ルートコンポーネント。GPS 取得・店舗パネル表示・デバッグUI を管理 |
| `src/Map.jsx` | HTML5 Canvas によるマップ描画・スケーリング・現在地ポインター |
| `src/MapPoint.jsx` | 店舗データ定義・GPS 座標 ↔ 画像座標の変換ロジック |
| `src/SVGButton.jsx` | SVG アイコンの汎用ボタン |
| `src/Window.jsx` | ウィンドウサイズの監視 hook |

### 座標変換の仕組み

マップ画像（2395×647px）上に GPS 座標をマッピングするため、`MapPoint` クラスが画像の上辺・下辺に複数の基準点（`tops` / `bottoms`）を保持する。実際の GPS 座標から経度で基準点間を補間し、続いて緯度で上辺・下辺間を補間することで Canvas 上の (x, y) 座標を算出する。

### 店舗データの管理

店舗情報は `MapPoint.initShops()` にハードコードされており、外部 API や CSV ファイルからは読み込まない。各店舗は `Shop` クラス（key, name, detail, 矩形領域 `Rect`）で表現される。画像は `/maps/images/{key}_0.webp` を参照する。

### Vite 設定

`vite.config.js` で `base: "/maps/"` を設定しているため、すべてのアセットパスが `/maps/` 起点になる。

## Key Conventions

- Management UI, comments, and commit messages are in Japanese.
- The DAO layer separates metadata entities from data entities (e.g., `Page` + `PageData`, `File` + `FileData`) to reduce read costs.
- Template rendering uses Go's `html/template` with custom functions registered in `api/helper.go`.
- No test files exist in the repo; manual/integration testing is the current approach.
