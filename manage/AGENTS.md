# manage/ — 管理画面 SPA

React (Vite + MUI) ベースの SPA 管理画面。`/manage/` で配信され、唯一の管理 UI として使用。旧テンプレートベースの管理画面 (V1) は削除済み。
API を提供する Go 側は `app/`（[app/AGENTS.md](../app/AGENTS.md)）を参照。
画面・API を追加するときは [manage-feature](../docs/skills/manage-feature/SKILL.md) の手順に従う。

## コマンド

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

## 構成

| ファイル/ディレクトリ | 役割 |
|---|---|
| `src/App.jsx` | React Router セットアップ (basename: `/manage`) |
| `src/components/Layout.jsx` | サイドバー + Outlet |
| `src/pages/PageList.jsx` | ページツリー一覧 (遅延展開) |
| `src/pages/PageEdit.jsx` | ページ編集フォーム・公開/非公開操作 |
| `src/api/page.js` | API クライアント (fetch wrapper) |
| `eslint.config.js` | ESLint flat config（React 19 / ESLint 10 構成） |

## ESLint 構成

`maps/` と同じく ESLint 10 の flat config を使用する。`eslint-plugin-react` は脆弱性アドバイザリと ESLint 10 非対応のため採用していない（詳細は [maps/AGENTS.md](../maps/AGENTS.md) の「ESLint 構成」を参照）。

`eslint-plugin-react-hooks` v7 の `set-state-in-effect` / `immutability` は、各ページの「マウント時にフェッチして setState」パターンが抵触するため `off`。リファクタ後に個別に有効化すること。

## MUI v9 の注意点

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

## Go 側の対応ファイル

| ファイル | 役割 |
|---|---|
| `app/handler/internal/manage_spa.go` | `_assets/manage/` の embed + 静的配信 + SPA catch-all |
| `app/handler/manage/api.go` | API ルート登録 (`/manage/api/` + `/manage/` catch-all) |
| `app/handler/manage/api_page.go` | JSON API ハンドラ (Page CRUD・公開・非公開) |

## API エンドポイント (すべて `/manage/api` 配下)

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

## 認証

- 認証は既存の Gorilla セッションをそのまま使用。`/manage/` 配下全体を `ManageHandler` が保護。
- `/manage/api/` への未認証アクセスはリダイレクトではなく JSON 401 を返す。
- CSRF 対策: セッション Cookie は `SameSite=Lax`（本番は `Secure` 付き）。`/manage/` 配下の非 GET リクエストは `ManageHandler` が Origin / Referer の同一オリジン検証を行い、不一致は 403。GIS のログイン POST（`/session`）はクロスサイトのため検証対象外。
- ログイン・管理者判定の仕組みは [app/AGENTS.md](../app/AGENTS.md) の「Authentication」を参照。
