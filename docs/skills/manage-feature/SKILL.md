---
name: manage-feature
description: 管理画面に機能を追加する手順。Go の JSON API（app/handler/manage/api_*.go）、API クライアント（manage/src/api/*.js）、画面（manage/src/pages/*.jsx）、ルーティングとサイドバーを一通り作る。「管理画面に〇〇の画面を追加」「管理 API を追加」「新しいエンティティを管理できるようにする」といった作業で使う。
---

# 管理画面への機能追加

既存の Variable（`api_variable.go` / `src/api/variable.js` / `VariableList.jsx` / `VariableEdit.jsx`）が一番小さい一式なので、これを手本にする。

## 1. Datastore（新しい Kind が要るときだけ）

- `app/datastore/` に Kind ごとに 1 ファイル作る。
- 一覧で読むメタ情報と本文を分ける（`Variable` + `VariableData` のように）。一覧取得で本文を読まないため。
- `datastore.NewDao()` を使う。クライアントはアプリ全体で共有されており、`Close()` は no-op。
- **バックアップ／リストアの対象に含めるか決める。** 含めるなら `datastore/backup.go` の `gob.Register()` と `CreateBackupData()` の `kinds` に追加する（今は Draft は対象外）。

## 2. Go の API ハンドラ

`app/handler/manage/api_<名前>.go` を作る。

```go
func register<Name>API(r *mux.Router) {
	r.HandleFunc("/<name>/", apiList<Name>s).Methods("GET")
	r.HandleFunc("/<name>/", apiCreate<Name>).Methods("POST")
	r.HandleFunc("/<name>/{key}", apiGet<Name>).Methods("GET")
	r.HandleFunc("/<name>/{key}", apiUpdate<Name>).Methods("POST")
	r.HandleFunc("/<name>/{key}", apiDelete<Name>).Methods("DELETE")
}
```

- `api.go` の `registerAPI()` に `register<Name>API(api)` を追加する。**SPA のキャッチオールより前**に登録すること（`api.go` の並びに足せばよい）。
- 応答は `apiJSON(w, v)`、エラーは `apiError(w, msg, code)`（`{"error": msg}` を返す）。どちらも `api_page.go` にある。
- レスポンス用の構造体（`api<Name>Res`）を定義し、JSON キーは camelCase にする。
- 更新は PUT ではなく `POST /<name>/{key}`、楽観ロックが要るものは `version` を受け取る（`api_variable.go` の `SetTargetVersion()`）。
- 認証・CSRF は `ManageHandler` が `/manage/` 全体に掛けているので、ハンドラ側では何もしなくてよい。
- 公開ページに影響する変更（ページの無効化など）は `HTML` エンティティの扱いに注意する（`app/AGENTS.md` の「ページの無効化と公開HTML」）。

## 3. API クライアント

`manage/src/api/<name>.js` を作る。各ファイルが自前の `apiFetch` を持つ作りなので、`variable.js` をコピーして `BASE` を変える。

- 401 のときは `/login` に遷移する。
- エラー時は応答の `error` をメッセージにして throw する。
- 同一オリジンの `fetch` は POST 時に `Origin` を付けるので、CSRF 検証はそのまま通る。

## 4. 画面

- `manage/src/pages/<Name>List.jsx` / `<Name>Edit.jsx` を作る。
- `src/App.jsx` に `lazy(() => import('./pages/...'))` とルートを追加する（ページ単位で遅延ロードする方針）。
- サイドバーに出すときは `src/components/Layout.jsx` のメニュー配列に `{ to, label, icon }` を追加する。アイコンは `@mui/icons-material/<Name>` から個別に import する。
- **MUI v9 の書き方にする**（`Grid size={{...}}`、`slotProps`、`Typography` は `sx` など）。v5 の書き方はビルドが通っても効かない（`manage/AGENTS.md` の「MUI v9 の注意点」）。
- 画面内に操作説明の文章は書かない（ツールチップは可）。

## 5. 確認

```powershell
cd manage
npm run lint
cd ..\app
go generate ./handler/internal
go build ./...
go run ./_cmd/main.go
```

`http://localhost:8080/manage/` で動作を確認する（ローカル起動は [fes-setup](../fes-setup/SKILL.md)）。

## 6. ドキュメント

- 新しいエンドポイントは `manage/AGENTS.md` の API 一覧に追記する。
- ページやファイル構成が増えたら `manage/AGENTS.md` の「構成」も更新する。
