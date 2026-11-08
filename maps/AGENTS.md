# maps/ — 独立した React アプリ

`maps/` は Go バックエンド（`app/`）とは独立した単体の React SPA。フェスティバルの会場マップをブラウザ上に表示し、来場者の現在地を GPS で重ねて表示する。
本番では `npm run deploy` で `dist/` を `app/handler/internal/_assets/maps/` にコピーして embed し、`/maps/` で配信する（下の「反映」）。店舗やマップの更新手順は [maps-update](../docs/skills/maps-update/SKILL.md) を参照。

## コマンド

```powershell
cd maps
npm install
npm run dev       # 開発サーバー (base path: /maps/)
npm run build     # maps/dist/ に出力
npm run lint      # ESLint (警告0件が必須)
npm run preview   # 本番ビルドのプレビュー
npm run deploy    # lint → ビルド → app の embed 用ディレクトリへコピー
```

## 反映（npm run deploy）

`npm run deploy` は lint（警告 0 件が必須）→ ビルド → `dist/` を `app/handler/internal/_assets/maps/` へコピー（中身を入れ替え）までを行う。コピーは `tools/deploy.mjs`（Node）で行い、店舗画像（`public/images/{店舗番号}/`。Git に入っていない）が無いときは警告を出す。GAE への反映は含まない（→ [fes-deploy](../docs/skills/fes-deploy/SKILL.md)）。

`app/` の `go generate ./...` からも、`manage/` とあわせて呼ばれる（→ [app/AGENTS.md](../app/AGENTS.md) の「フロントエンドの反映」）。

`_assets/maps/` は `.gitignore` 対象で、`//go:embed all:_assets/maps/*` は**ファイルが 1 つも無いとコンパイルエラー**になる。

## 構成

| ファイル | 役割 |
|---|---|
| `src/App.jsx` | ルートコンポーネント。GPS 取得・店舗パネル表示・デバッグUI を管理 |
| `src/Map.jsx` | HTML5 Canvas によるマップ描画・スケーリング・現在地ポインター |
| `src/MapPoint.jsx` | GPS 座標 → 画像座標の変換（薄板スプライン）・店舗データの読み込み |
| `src/data/gps.json` | GPS の基準点（現地調査の JSON から `tools/gps.mjs` で生成する。直接編集しない） |
| `tools/gps.mjs` | 現地調査（`/maps/?debug` の「保存」）の JSON から `src/data/gps.json` を作る |
| `src/data/shops.json` | 店舗・施設のデータ（`tools/shops` で生成する。直接編集しない） |
| `data/<開催>/shops.tsv` | 店舗の対応表（番号・店名・説明・申込との対応・タップ範囲） |
| `tools/shops/main.go` | 対応表と出店申込の CSV から `src/data/shops.json` を作る |
| `tools/images/main.go` | 申込の写真（Drive の ZIP）を店舗番号ごとに分け、webp 化したものを `public/images` に置く |
| `tools/heicrot/` | HEIC の写真の向きを EXIF に従って直す（cgo を使うため別モジュール） |
| `data/<開催>/README.md` | その開催で使ったデータと作業の記録 |
| `src/LocationHelp.jsx` | 位置情報が拒否されているときの戻し方（端末・ブラウザ別。現在地ボタンから出す） |
| `src/useSurvey.js` / `src/Survey.jsx` | 現地調査（`/maps/?debug` のときだけ）。GPS と地図上の位置の組・軌跡を記録して書き出す（使い方は `data/2026-Fall/README.md`） |
| `src/SVGButton.jsx` | SVG アイコンの汎用ボタン |
| `src/Window.jsx` | ウィンドウサイズの監視 hook |
| `eslint.config.js` | ESLint flat config（React 19 / ESLint 10 構成） |

## ESLint 構成

ESLint 10 の flat config (`eslint.config.js`) を使用する。`eslint-plugin-react` は依存する `minimatch@3` → `brace-expansion@1` に脆弱性アドバイザリが残り、かつ ESLint 10 を peer に持たないため採用していない（`eslint-plugin-react-hooks` + `eslint-plugin-react-refresh` + `@eslint/js` で構成）。

`eslint-plugin-react-hooks` v7 で追加された `set-state-in-effect` / `immutability` と `exhaustive-deps` は既存コード（`App.jsx` / `Map.jsx` / `SVGButton.jsx`）が多数抵触するため `off`。該当箇所をリファクタしたら個別に有効化すること。`static-components`（描画の中でコンポーネントを定義しない）は有効。App の中で定義したお知らせが、位置が届くたびに作り直されてフェードインを繰り返したため（2026-10-18）。

店舗名・説明文に全角スペース (U+3000) を使うため `no-irregular-whitespace` は `skipStrings` / `skipTemplates` を有効にしている。

## 座標変換の仕組み

マップ画像（2000×522px）はデフォルメされているので、現地で記録した基準点（`src/data/gps.json`。画像上の位置と緯度・経度の組）をなめらかにつなぐ薄板スプラインで変換する。`MapPoint.initTransform()` が基準点の数 + 3 元の連立一次方程式を解いて重みを作り、`MapPoint.get()` が緯度・経度から (x, y) を出す。`smoothing` を大きくすると GPS の誤差に引きずられにくくなる。基準点の外側はほぼアフィンで延長する。ハッちは `MapPoint.locate()` で、全身が見えるように端から内側に寄せて描く。画像の外でも端から 100m 以内なら端に寄せ、本当の位置の方向に赤い三角を付ける（それより遠いと出さない）。現在地ボタンを押すと、ハッちが画面の真ん中に来るように横に動かす。基準点の取り直しは [maps-update](../docs/skills/maps-update/SKILL.md) を参照。

## 店舗データの管理

店舗・施設は `src/data/shops.json` をビルド時に読み込み、`MapPoint.initShops()` でタップ範囲（`rects`）ごとに `Shop`（key, name, detail, 矩形領域 `Rect`）を作る。JSON は対応表（`data/<開催>/shops.tsv`）と出店申込の CSV から `tools/shops` で生成する（手順は [maps-update](../docs/skills/maps-update/SKILL.md)）。申込の CSV は電話番号や住所を含むため**リポジトリに置かない**。画像は `public/images/{key}/{n}.webp` に置き、`shops.json` の `images` で参照する（申込の写真からは `tools/images` で作る）。マップはアーカイブせず開催ごとに作り直すため、**店舗番号のディレクトリの画像は Git に入れない**（`.gitignore` 対象）。施設の写真（`PLAY`・`SHIP`・`WC` など。現地で撮ったもの）も同じく Git に入れず、原本と控えは `maps/work/` に置く。管理するのはアプリの素材（ハッち `BIRD`・隠れハッち `SECRET`・`noimage.png`）と、以前からある `SEA/0.webp` だけ。

## Vite 設定

`vite.config.js` で `base: "/maps/"` を設定しているため、すべてのアセットパスが `/maps/` 起点になる。
