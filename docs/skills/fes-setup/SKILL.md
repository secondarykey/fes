---
name: fes-setup
description: FES の開発環境・本番環境の初期セットアップ手順。environment.json（CLIENT_ID / SESSION_KEY / MANAGERS）の用意、Datastore エミュレータでのローカル起動、初回ログインと管理者の登録、Site Settings の初期設定を扱う。「新しい環境で動かす」「ローカルで起動できない」「ログインできない」「管理者を追加」といった作業で使う。
---

# 初期セットアップ

## 1. environment.json を用意する

`app/handler/internal/_assets/environment.json` は embed され、起動時にすべて環境変数へ展開される（`handler.setEnvironment()`）。

```json
{
  "CLIENT_ID": "xxxx.apps.googleusercontent.com",
  "CLIENT_SECRET": "",
  "SESSION_KEY": "<ランダムな文字列>",
  "MANAGERS": "someone@example.com,another@example.com"
}
```

| キー | 用途 |
|---|---|
| `CLIENT_ID` | Google Identity Services の OAuth クライアント ID。ログイン画面（`authentication.tmpl`）と ID トークン検証の audience に使う |
| `CLIENT_SECRET` | 現在のコードでは使っていない |
| `SESSION_KEY` | セッション Cookie の署名鍵。**本番で空だと起動時にエラー**。開発時は空でもランダム鍵で起動する（再起動でログアウト） |
| `MANAGERS` | 初期管理者のメール（カンマ区切り）。`Site.Managers` が空のときだけ使う |

- `SESSION_KEY` は例えば `[Convert]::ToBase64String((1..32 | % { Get-Random -Max 256 }))` のように 32 バイト程度のランダム値にする。
- **値はコミットしない。** このファイルは追跡対象なので、手元の値は次で隠す。
  ```powershell
  git update-index --skip-worktree app/handler/internal/_assets/environment.json
  ```
  テンプレート側を更新したいときだけ `--no-skip-worktree` で戻す。

### OAuth クライアントの設定（GCP コンソール）

「承認済みの JavaScript 生成元」に公開 URL と `http://localhost:8080`（ローカル用）を登録する。未登録の生成元ではログインボタンが動かない。

## 2. ビルドに必要な embed 対象を用意する

`.gitignore` 対象のため、クローン直後は空のディレクトリがある。

- `app/handler/internal/_assets/maps/` … **ファイルが 1 つも無いとコンパイルエラー**（`//go:embed all:_assets/maps/*`）。`maps/` をビルドしてコピーする（→ [fes-deploy](../fes-deploy/SKILL.md) の手順 2）。
- `app/handler/internal/_assets/manage/` … **ディレクトリが無いとコンパイルエラー**（`//go:embed _assets/manage`）。`cd app; go generate ./handler/internal` で生成する。

## 3. ローカルで起動する

GCE 外では `DevelopMode=true` になり、Datastore エミュレータに接続する（既定 `localhost:8081`、プロジェクト ID `fes`。`DATASTORE_EMULATOR_HOST` で変更可）。

```powershell
gcloud beta emulators datastore start --host-port=localhost:8081 --project=fes
# 別ターミナルで
cd app
go run ./_cmd/main.go      # http://localhost:8080 （PORT で変更可）
```

本番データで確認したいときは、本番の Site Settings でバックアップを取り、ローカルの Site Settings でリストアする（**リストアは全上書き。本番では実行しない**）。

## 4. 初回ログインと管理者の登録

1. `MANAGERS` に自分のメールを入れて起動（本番ならデプロイ）する。
2. `/login` から Google でログインする。`Site.Managers` が空なので `MANAGERS` で判定される。
3. 管理画面の Site Settings で管理者のメールを登録して保存する。
   - **保存した時点から `Site.Managers` だけで判定される**（`MANAGERS` は使われなくなる）。自分のメールも必ず含めること。含め忘れるとログインできなくなる。
4. Site Settings で必要に応じて次を設定する。
   - ルートページ
   - `BaseURL` … sitemap / robots.txt の URL に使う。未設定だと Host ヘッダから組み立てる。
   - `ArchiveBucket` / アーカイブ名 / 1 日のアクセス上限（→ [fes-archive](../fes-archive/SKILL.md)）

## ログインできないとき

| 症状 | 確認すること |
|---|---|
| ログインボタンが出ない・エラーになる | `CLIENT_ID` と、OAuth クライアントの JavaScript 生成元 |
| 403「管理者として登録されていません」 | `Site.Managers`（設定済みならこちらが優先）と `MANAGERS` |
| 本番で起動しない（`SESSION_KEY is not set`） | `environment.json` の `SESSION_KEY` |
| ローカルで再起動のたびにログアウトする | `SESSION_KEY` が空（開発時はランダム鍵になる） |
