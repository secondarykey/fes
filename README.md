fes is Festival Edit System.

# 概要

テンプレートを作成して、ツリー状にページを作成していきます。

# Issue

- 認証ボタンを真ん中に
- 認証機能をGUIで可能に
- Archive機能をGUIで可能に
- アーティスト機能
- キャッシュ機能
- URL機能
- エラーページ
- datastoreからRequest削除
- 以下をマニュアルに記載

## データの扱い

- Site

- Page

- PageData

- Content

- Children

- Top

- Dir

- Prev

- Next

## テンプレート

"html"
"eraseBR"
"plane"
"convertDate"
"list"
"mark"
"templateContent"
"variable"



## 認証の方法


```
    "CLIENT_ID":"",
    "CLIENT_SECRET":""
```

git update-index --skip-worktree app/handler/internal/\_assets/environment.json

git update-index --no-skip-worktree 
