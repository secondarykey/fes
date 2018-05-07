fes is Festival Edit System.

・・・って名前つけたけど、フェス専用じゃなくてただのWebCMSになった。
GoogleAppEngine上で動作します。

まだまだデータストアへのアクセス数などを考えて設計変更していきますが、基本路線はこんな感じかな？
実際に表示しているサイトのデータとかも載せていきたいと思っています。テスト書かないとね。

多分マニュアルはWikiに書いていきます。
一部を[Wiki](https://github.com/secondarykey/fes/wiki)に移動

# デプロイ方法

開発サーバは

> dev_appserver.yaml app.yaml

デプロイは

> gcloud app deploy --project={projectid} --version={versionname} app.yaml

vendoring やってないのだけど、いろいろ試行錯誤中で、まだデプロイ方法が不確定。

# 管理画面へのアクセス

/manage/により管理されます。 ロールはプロジェクトに対するadminで可能です。

# データ構造

Site
Tetmplate -> TetmplateData
Page -> PageData
File -> FileData

になります。
XxxxData はバイナリデータが格納されています。

# 表示フロー

ページに設定したテンプレート（サイト、ページ）により、基本的なページの部分が表示されます。
サイトのテンプレート内にページのテンプレートを埋め込み、
ページのテンプレートでページのデータを構成していきます。

APIは後述しますが、

```
<html>
  <head>
    <title>{{ .Site.Name }} {{ .Page.Name }}</title>
  </head>
  <body>
  {{ template page_template}}
  </body>
</html>
```

とサイトテンプレートを作成
ページテンプレートとして

```
{{ .Page.Description }}
{{ html .PageData.Content }}
```

を登録するとそのページの情報が書き込まれる仕組みです。


# 表示データの仕様

表示するデータはPageになります。

/page/{id} によりアクセス可能です
/file/{id} でファイル登録されたデータにアクセスできます

