fes is Festival Edit System.

# 概要

テンプレートを作成して、ツリー状にページを作成していきます。

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

app/handler/internal/\_assets/environment.json の値を編集します。

```
    "CLIENT_ID":"",
    "CLIENT_SECRET":""
```

この値はGCPのプロジェクトからAPIのOAuth2を設定して、設定します。

この値により管理機能にGoogle認証が追加されます。
認証を許すメールアドレスを管理画面上で設定します。

git update-index --skip-worktree app/handler/internal/\_assets/environment.json

git update-index --no-skip-worktree 
