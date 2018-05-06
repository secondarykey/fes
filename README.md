fes is Festival Edit System.

・・・って名前つけたけど、フェス専用じゃなくてただのWebCMSになった。
GoogleAppEngine上で動作します。

多分マニュアルはWikiに書いていきます。
この後のは開発中に忘れないように書いたやつで、Wikiに移します。

# デプロイ方法

開発サーバは

> dev_appserver.yaml app.yaml

デプロイは

> gcloud app deploy --project={projectid} --version={versionname} app.yaml

でOKのはず。

vendoring やってないのだけど、いろいろ試行錯誤中で、まだデプロイ方法が不確定。

## β版

まだまだデータストアへのアクセス数などを考えて設計変更していきますが、基本路線はこんな感じかな？
実際に表示しているサイトのデータとかも載せていきたいと思っています。テスト書かないとね。

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


# 管理画面へのアクセス

/manage/により管理されます。 ロールはプロジェクトに対するadminで可能です。

# 表示データの仕様

表示するデータはPageになります。

/page/{id} によりアクセス可能です
/file/{id} でファイル登録されたデータにアクセスできます


# テンプレートAPI

テンプレート上でのAPIアクセスを紹介しておきます

Xxxxx.Key.StringID による各データのIDを取得できます。
File以外のキー値はUUIDで管理しています。
※ただしページでのFile追加時はUUID

## {{ template page_template }}

サイトテンプレートにページのテンプレートを埋め込むときに使用します。
ページの設定で行うことが可能です

## Site

### Site.Name

設定しているサイトの名称が取得できます
{{ .Site.Name }} で文字列の取得が可能です

### Site.Description

サイトの説明文です。metaなどに埋め込みます

## Page

### Page.Name

ページの名称が取得できます

### Page.Description

ページの説明が取得できます

### PageData.Content | html,plane

ページデータにアクセスするにはhtmlかplaneを前につけます
{{ html .PageData.Content }}
{{ plane .PageData.Content }}
という風になります。

HTMLで返したいときはhtml、単純にテキストとして扱いたい場合はplaneです。

## File

ファイルデータへのアクセスは基本的にURLベースになります。

```
<img src="/file/{filename}">
```

ファイルのキーはfilenameを使っています。


## More

上記にそれぞれのページデータへのアクセスが可能です

### Children

そのページの子ページの情報にアクセスします
リストでそのページの以下にアクセスしたい場合は

```
<ul>
{{ range .Children }}
   <li> <a href="/page/{{.Key.StringID}}"> {{.Name}} </a> </li>
{{ end }}
</ul>
```

という風に行います。rangeではページデータが取れているのでPageで説明のあったAPIが使用可能です。
※[PageData]は使用できません

### list

{{ range list "page id" }}
{{ end }}

Childrenは自分の子のページを取得しますが、listはページIDに対する子ページのリストを返します


