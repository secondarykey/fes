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
- アップロード先バケットは Datastore の Site エンティティ `ArchiveBucket` フィールドから取得（未設定時は `config.ArchiveBucket` のデフォルト値 `"hummingbird-archives"` を使用）。

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

### ページの無効化と公開HTML

公開側の `pageView()` は `HTML` エンティティのみを参照し、`Page.Deleted` を見ない。そのため `Deleted` を立てるだけでは公開ページが残り続ける。

これを避けるため、ページを無効化する経路では `HTML` を同時に削除する。

- `PutPage()` — ページ編集画面の保存（`Deleted=true` で保存した場合）
- `PutPageSequence()` — 子ページ管理の「有効」スイッチ OFF ＋保存

いずれも `Page.Publish` / `Republish` をゼロ値に戻す。以前は「公開ページごと消えてしまう」ことを懸念して削除を見送っていた（`PutPage()` に TODO として残っていた）が、無効化＝公開停止という運用に合わせて削除する方針に変更した。一時的に隠して後で戻す場合は、再度有効化してから公開し直す必要がある。

この変更以前に無効化されたページには `HTML` が残っている。Site Settings の「サイトクリーン」（`apiSiteClean`）が孤立 HTML とあわせてこれらを削除する。

### Embedded assets

`app/handler/internal/_assets/` is embedded with `//go:embed`. It contains:
- `environment.json` — OAuth2 credentials (CLIENT_ID, CLIENT_SECRET)、セッション署名鍵 (SESSION_KEY)、初期管理者メール (MANAGERS)
- `archives/` — static archive zips（ZIP 形式の旧アーカイブ）
- `manage/` — 管理画面 SPA の Vite ビルド成果物

### GCS アーカイブ配信

`app/handler/internal/archive_gcs.go` が GCS バケットからアーカイブを配信する。

- 配信対象アーカイブ名とバケット名は Datastore の Site エンティティから取得（未設定時は `config.ArchiveNames` / `config.ArchiveBucket` をフォールバック）。管理画面の Site Settings から編集可能。
- 各アーカイブは `/{name}/` パスで配信される（例: `/2026-Spring/`）。
- キャッシュ: `storage.Reader.Attrs.LastModified` を `Last-Modified` ヘッダに設定し、`If-Modified-Since` が一致すれば 304 を返す。GCS API 呼び出しは 1 リクエストあたり 1 回。
- `Cache-Control: public, max-age=86400` を付与。
- アーカイブ名の追加は管理画面で設定後、再デプロイが必要（ハンドラ登録は起動時のみ）。

### テンプレートキャッシュの注意点

`logic/html.go` のテンプレートキャッシュは `Generator` のフィールド (`tmpCache`) としてリクエスト単位で保持する。以前はパッケージレベル変数 `cacheTemplateData` だったため、テンプレート更新後も古い内容が反映される問題と、並行リクエスト時のデータレースがあった。キャッシュを共有化する場合はこの経緯に注意すること。

### Authentication

Google Identity Services (GIS) によるログイン。`/session` に POST された ID トークンを `google.golang.org/api/idtoken` で検証する（署名・有効期限・audience=CLIENT_ID・発行者・email_verified）。CLIENT_ID は `environment.json` から環境変数経由で設定。セッションは Gorilla sessions で管理し、署名鍵は環境変数 `SESSION_KEY`（`environment.json` 由来）を使用する。本番（`DevelopMode=false`）で `SESSION_KEY` 未設定なら `manage.Register()` が起動時にエラーを返す。開発時は未設定でもランダム鍵で起動する（再起動でセッション無効化）。

管理者判定はフェイルクローズ。`Site.Managers` が設定されていればそれで判定し、未設定（初期セットアップ）時のみ `environment.json` の `MANAGERS`（カンマ区切りのメール一覧）をブートストラップ用に使用する。どちらにも該当しないメールは 403。初期管理者は `MANAGERS` に自分のメールを記述してデプロイ→ログイン後、Site Settings から他の管理者を追加する（追加後は Datastore の `Site.Managers` が優先される）。

## manage/ — 管理画面 SPA

React (Vite + MUI) ベースの SPA 管理画面。`/manage/` で配信され、唯一の管理 UI として使用。旧テンプレートベースの管理画面 (V1) は削除済み。

### コマンド

```powershell
cd manage
npm install
npm run dev       # 開発サーバー (Vite proxy 不要、Go サーバー経由で動作確認)
npm run build     # manage/dist/ に出力
```

ビルドから embed 用ディレクトリへの反映までは `go generate` で行う（robocopy を使う `npm run deploy` は廃止した）。

```powershell
cd app
go generate ./handler/internal
go build ./...
```

`app/handler/internal/manage_spa.go` の `//go:generate` が `app/_cmd/managebuild/main.go` を呼び、次を実行する。

1. `manage/node_modules` が無ければ `npm install`
2. `manage/` で `npm run build`
3. `app/handler/internal/_assets/manage/` を削除して `manage/dist/` を丸ごとコピー（旧ファイルが embed に残らないようにするため）

`-skip-build` を付けるとビルドを行わず既存の `dist/` の同期だけを行う。`-src` / `-dst` でパスを変更できる（既定値は `go generate` の実行ディレクトリ = `app/handler/internal` 基準）。

### 構成

| ファイル/ディレクトリ | 役割 |
|---|---|
| `src/App.jsx` | React Router セットアップ (basename: `/manage`) |
| `src/components/Layout.jsx` | サイドバー + Outlet |
| `src/pages/PageList.jsx` | ページツリー一覧 (遅延展開) |
| `src/pages/PageEdit.jsx` | ページ編集フォーム・公開/非公開操作 |
| `src/api/page.js` | API クライアント (fetch wrapper) |
| `eslint.config.js` | ESLint flat config（React 19 / ESLint 10 構成） |

### ESLint 構成

`maps/` と同じく ESLint 10 の flat config を使用する。`eslint-plugin-react` は脆弱性アドバイザリと ESLint 10 非対応のため採用していない（詳細は maps 側の「ESLint 構成」を参照）。

`eslint-plugin-react-hooks` v7 の `set-state-in-effect` / `immutability` は、各ページの「マウント時にフェッチして setState」パターンが抵触するため `off`。リファクタ後に個別に有効化すること。

### MUI v9 の注意点

MUI は v9 系を使用する。v5 から移行済みで、以下は v5 の書き方をすると**ビルドは通るが実行時に無視される**ため注意すること。

| 旧 (v5) | 新 (v9) |
|---|---|
| `<Grid item xs={12} sm={6}>` | `<Grid size={{ xs: 12, sm: 6 }}>` |
| `<TextField InputProps={{...}}>` | `<TextField slotProps={{ input: {...} }}>` |
| `<TextField inputProps={{...}}>` | `<TextField slotProps={{ htmlInput: {...} }}>` |
| `<ListItemText primaryTypographyProps={{...}}>` | `<ListItemText slotProps={{ primary: {...} }}>` |
| `<ListItem button onClick=...>` | `<ListItem disablePadding><ListItemButton onClick=...>` |
| `<Typography fontWeight={600}>` | `<Typography sx={{ fontWeight: 600 }}>` |

特に `Typography` は v9 で `extendSxProp` が外れ、`fontWeight` / `fontFamily` などのシステムプロパティを直接渡すと**スタイルが当たらず不正な HTML 属性として出力される**（`Box` は従来どおりシステムプロパティ対応）。`sx` を使うこと。

`vite.config.js` にあった `@mui/icons-material` → `@mui/icons-material/esm` のエイリアスは、v9 で `esm/` ディレクトリが廃止されたため削除済み。復活させるとビルドが壊れる。

### Go 側の対応ファイル

| ファイル | 役割 |
|---|---|
| `app/handler/internal/manage_spa.go` | `_assets/manage/` の embed + 静的配信 + SPA catch-all |
| `app/handler/manage/api.go` | API ルート登録 (`/manage/api/` + `/manage/` catch-all) |
| `app/handler/manage/api_page.go` | JSON API ハンドラ (Page CRUD・公開・非公開) |

### API エンドポイント (すべて `/manage/api` 配下)

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

### 認証

- 認証は既存の Gorilla セッションをそのまま使用。`/manage/` 配下全体を `ManageHandler` が保護。
- `/manage/api/` への未認証アクセスはリダイレクトではなく JSON 401 を返す。
- CSRF 対策: セッション Cookie は `SameSite=Lax`（本番は `Secure` 付き）。`/manage/` 配下の非 GET リクエストは `ManageHandler` が Origin / Referer の同一オリジン検証を行い、不一致は 403。GIS のログイン POST（`/session`）はクロスサイトのため検証対象外。

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
| `eslint.config.js` | ESLint flat config（React 19 / ESLint 10 構成） |

### ESLint 構成

ESLint 10 の flat config (`eslint.config.js`) を使用する。`eslint-plugin-react` は依存する `minimatch@3` → `brace-expansion@1` に脆弱性アドバイザリが残り、かつ ESLint 10 を peer に持たないため採用していない（`eslint-plugin-react-hooks` + `eslint-plugin-react-refresh` + `@eslint/js` で構成）。

`eslint-plugin-react-hooks` v7 で追加された `set-state-in-effect` / `static-components` / `immutability` と `exhaustive-deps` は既存コード（`App.jsx` / `Map.jsx` / `SVGButton.jsx`）が多数抵触するため `off`。該当箇所をリファクタしたら個別に有効化すること。

店舗名・説明文に全角スペース (U+3000) を使うため `no-irregular-whitespace` は `skipStrings` / `skipTemplates` を有効にしている。

### 座標変換の仕組み

マップ画像（2395×647px）上に GPS 座標をマッピングするため、`MapPoint` クラスが画像の上辺・下辺に複数の基準点（`tops` / `bottoms`）を保持する。実際の GPS 座標から経度で基準点間を補間し、続いて緯度で上辺・下辺間を補間することで Canvas 上の (x, y) 座標を算出する。

### 店舗データの管理

店舗情報は `MapPoint.initShops()` にハードコードされており、外部 API や CSV ファイルからは読み込まない。各店舗は `Shop` クラス（key, name, detail, 矩形領域 `Rect`）で表現される。画像は `/maps/images/{key}_0.webp` を参照する。

### Vite 設定

`vite.config.js` で `base: "/maps/"` を設定しているため、すべてのアセットパスが `/maps/` 起点になる。

## Key Conventions

- Management UI, comments, and commit messages are in Japanese.
- The DAO layer separates metadata entities from data entities (e.g., `Page` + `PageData`, `File` + `FileData`) to reduce read costs.
- Datastore クライアントはアプリ全体で 1 つを共有する（`datastore.Dao` はステートレスで、`Close()` は互換性のための no-op）。
- ルーティングは `app.go` の `registerHandler()` が生成する専用の `http.ServeMux` に登録する。`http.DefaultServeMux` はメインサーバーでは使用しない（pprof は `/manage/debug/pprof/` 配下・要ログイン）。
- Site エンティティの `BaseURL` を設定すると sitemap / robots.txt がその URL で生成される。未設定時はリクエストの Host ヘッダから組み立てる。
- Template rendering uses Go's `html/template` with custom functions registered in `api/helper.go`.
- No test files exist in the repo; manual/integration testing is the current approach.
