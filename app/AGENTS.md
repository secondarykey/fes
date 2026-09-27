# app/ — Go バックエンド

公開サイト・管理 API・静的アーカイブを提供する Go アプリケーション。Go モジュールは `app/go.mod`。
管理画面 SPA のフロントエンドは `manage/`（[manage/AGENTS.md](../manage/AGENTS.md)）を参照。

## Build & Run Commands

`app/` ディレクトリで実行する。

```powershell
# Run locally
go run ./_cmd/main.go

# Build binary
go build -o app ./_cmd

# Deploy to Google App Engine
gcloud app deploy app.yaml
```

ローカル（GCE 外）では `DevelopMode=true` になり、Datastore エミュレータ（既定 `localhost:8081`）に接続する。セットアップは [fes-setup](../docs/skills/fes-setup/SKILL.md)、デプロイは [fes-deploy](../docs/skills/fes-deploy/SKILL.md) を参照。

### Static Site Archive

アーカイブツール (`_cmd/archive/main.go`) は静的サイトを生成し、GCS にアップロードする。手順は [fes-archive](../docs/skills/fes-archive/SKILL.md) を参照。

## Architecture

### Backend layers

| Package | Role |
|---|---|
| `handler/` | HTTP handlers for public routes (pages, files, auth) |
| `handler/manage/` | Admin console routes |
| `handler/internal/` | Cache server, static asset serving, archive |
| `datastore/` | DAO layer — one file per Datastore kind |
| `logic/` | Business logic (static site generation, HTML rendering) |
| `api/` | Template helpers (`html/template` FuncMap, content conversion) |
| `config/` | Functional-option pattern for app configuration |

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

`handler/internal/_assets/` is embedded with `//go:embed`. It contains:
- `environment.json` — OAuth2 credentials (CLIENT_ID, CLIENT_SECRET)、セッション署名鍵 (SESSION_KEY)、初期管理者メール (MANAGERS)。`skip-worktree` で手元の値を隠している
- `archives/` — ZIP 形式の旧アーカイブ（embed は廃止済み）
- `manage/` — 管理画面 SPA の Vite ビルド成果物（`go generate ./handler/internal` で更新。手順は [manage/AGENTS.md](../manage/AGENTS.md)）
- `maps/` — 会場マップのビルド成果物（`.gitignore` 対象。空だとコンパイルエラー）
- `static/` / `templates/` — 公開側の静的ファイルとテンプレート

### GCS アーカイブ配信

`handler/internal/archive_gcs.go` が GCS バケットからアーカイブを配信する。

- 配信対象アーカイブ名・バケット名・1 日のアクセス上限は Datastore の Site エンティティ（`ArchiveNames` / `ArchiveBucket` / `ArchiveDailyLimit`）から取得する。フォールバック値は無い。管理画面の Site Settings から編集可能。
- 各アーカイブは `/{name}/` パスで配信される（例: `/2026-Spring/`）。
- キャッシュ: `storage.Reader.Attrs.LastModified` を `Last-Modified` ヘッダに設定し、`If-Modified-Since` が一致すれば 304 を返す。GCS API 呼び出しは 1 リクエストあたり 1 回。
- `Cache-Control: public, max-age=86400` を付与。
- Site Settings の保存時に `GCSArchiveRouter.Update()` が呼ばれるため、アーカイブ名の追加に再デプロイは不要。

### テンプレートキャッシュの注意点

`logic/html.go` のテンプレートキャッシュは `Generator` のフィールド (`tmpCache`) としてリクエスト単位で保持する。以前はパッケージレベル変数 `cacheTemplateData` だったため、テンプレート更新後も古い内容が反映される問題と、並行リクエスト時のデータレースがあった。キャッシュを共有化する場合はこの経緯に注意すること。

### Authentication

Google Identity Services (GIS) によるログイン。`/session` に POST された ID トークンを `google.golang.org/api/idtoken` で検証する（署名・有効期限・audience=CLIENT_ID・発行者・email_verified）。CLIENT_ID は `environment.json` から環境変数経由で設定。セッションは Gorilla sessions で管理し、署名鍵は環境変数 `SESSION_KEY`（`environment.json` 由来）を使用する。本番（`DevelopMode=false`）で `SESSION_KEY` 未設定なら `manage.Register()` が起動時にエラーを返す。開発時は未設定でもランダム鍵で起動する（再起動でセッション無効化）。

管理者判定はフェイルクローズ。`Site.Managers` が設定されていればそれで判定し、未設定（初期セットアップ）時のみ `environment.json` の `MANAGERS`（カンマ区切りのメール一覧）をブートストラップ用に使用する。どちらにも該当しないメールは 403。初期管理者は `MANAGERS` に自分のメールを記述してデプロイ→ログイン後、Site Settings から他の管理者を追加する（追加後は Datastore の `Site.Managers` が優先される）。

`/manage/` 配下の認証・CSRF 対策は [manage/AGENTS.md](../manage/AGENTS.md) の「認証」を参照。

## Conventions

- The DAO layer separates metadata entities from data entities (e.g., `Page` + `PageData`, `File` + `FileData`) to reduce read costs.
- Datastore クライアントはアプリ全体で 1 つを共有する（`datastore.Dao` はステートレスで、`Close()` は互換性のための no-op）。
- ルーティングは `app.go` の `registerHandler()` が生成する専用の `http.ServeMux` に登録する。`http.DefaultServeMux` はメインサーバーでは使用しない（pprof は `/manage/debug/pprof/` 配下・要ログイン）。
- Site エンティティの `BaseURL` を設定すると sitemap / robots.txt がその URL で生成される。未設定時はリクエストの Host ヘッダから組み立てる。
- Template rendering uses Go's `html/template` with custom functions registered in `api/helper.go`.
