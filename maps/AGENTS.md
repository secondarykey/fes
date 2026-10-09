# maps/ — 独立した React アプリ

`maps/` は Go バックエンド（`app/`）とは独立した単体の React SPA。フェスティバルの会場マップをブラウザ上に表示し、来場者の現在地を GPS で重ねて表示する。
本番では `dist/` を `app/handler/internal/_assets/maps/` にコピーして embed し、`/maps/` で配信する（[fes-deploy](../docs/skills/fes-deploy/SKILL.md)）。店舗やマップの更新手順は [maps-update](../docs/skills/maps-update/SKILL.md) を参照。

## コマンド

```powershell
cd maps
npm install
npm run dev       # 開発サーバー (base path: /maps/)
npm run build     # maps/dist/ に出力
npm run lint      # ESLint (警告0件が必須)
npm run preview   # 本番ビルドのプレビュー
npm run deploy    # lint → ビルド → app の embed 用ディレクトリへコピー（GAE への反映は含まない）
```

## 構成

| ファイル | 役割 |
|---|---|
| `src/App.jsx` | ルートコンポーネント。GPS 取得・店舗パネル表示・デバッグUI を管理 |
| `src/Map.jsx` | HTML5 Canvas によるマップ描画・スケーリング・現在地ポインター |
| `src/MapPoint.jsx` | 基準点の定義・GPS 座標 ↔ 画像座標の変換ロジック・店舗データの読み込み |
| `src/data/shops.json` | 店舗・施設のデータ（`tools/shops` で生成する。直接編集しない） |
| `data/<開催>/shops.tsv` | 店舗の対応表（番号・店名・説明・申込との対応・タップ範囲） |
| `tools/shops/main.go` | 対応表と出店申込の CSV から `src/data/shops.json` を作る |
| `tools/images/main.go` | 申込の写真（Drive の ZIP）を店舗番号ごとに分け、webp 化したものを `public/images` に置く |
| `tools/heicrot/` | HEIC の写真の向きを EXIF に従って直す（cgo を使うため別モジュール） |
| `data/<開催>/README.md` | その開催で使ったデータと作業の記録 |
| `src/SVGButton.jsx` | SVG アイコンの汎用ボタン |
| `src/Window.jsx` | ウィンドウサイズの監視 hook |
| `eslint.config.js` | ESLint flat config（React 19 / ESLint 10 構成） |

## ESLint 構成

ESLint 10 の flat config (`eslint.config.js`) を使用する。`eslint-plugin-react` は依存する `minimatch@3` → `brace-expansion@1` に脆弱性アドバイザリが残り、かつ ESLint 10 を peer に持たないため採用していない（`eslint-plugin-react-hooks` + `eslint-plugin-react-refresh` + `@eslint/js` で構成）。

`eslint-plugin-react-hooks` v7 で追加された `set-state-in-effect` / `static-components` / `immutability` と `exhaustive-deps` は既存コード（`App.jsx` / `Map.jsx` / `SVGButton.jsx`）が多数抵触するため `off`。該当箇所をリファクタしたら個別に有効化すること。

店舗名・説明文に全角スペース (U+3000) を使うため `no-irregular-whitespace` は `skipStrings` / `skipTemplates` を有効にしている。

## 座標変換の仕組み

マップ画像（2000×522px）上に GPS 座標をマッピングするため、`MapPoint` クラスが画像の上辺・下辺に複数の基準点（`tops` / `bottoms`）を保持する。実際の GPS 座標から経度で基準点間を補間し、続いて緯度で上辺・下辺間を補間することで Canvas 上の (x, y) 座標を算出する。基準点の外は端の区間を延長し、画像の外になる場合は現在地を表示しない。

## 店舗データの管理

店舗・施設は `src/data/shops.json` をビルド時に読み込み、`MapPoint.initShops()` でタップ範囲（`rects`）ごとに `Shop`（key, name, detail, 矩形領域 `Rect`）を作る。JSON は対応表（`data/<開催>/shops.tsv`）と出店申込の CSV から `tools/shops` で生成する（手順は [maps-update](../docs/skills/maps-update/SKILL.md)）。申込の CSV は電話番号や住所を含むため**リポジトリに置かない**。画像は `public/images/{key}/{n}.webp` に置き、`shops.json` の `images` で参照する（申込の写真からは `tools/images` で作る）。マップはアーカイブせず開催ごとに作り直すため、**店舗番号のディレクトリの画像は Git に入れない**（`.gitignore` 対象）。`BIRD` / `SEA` / `SECRET` と `noimage.png` だけを管理する。

## Vite 設定

`vite.config.js` で `base: "/maps/"` を設定しているため、すべてのアセットパスが `/maps/` 起点になる。
