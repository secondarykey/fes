---
name: maps-update
description: 会場マップ（maps/）の内容をフェスに合わせて更新する手順。出店申込からの店舗データ作成、店舗・ステージの追加や変更、店舗画像の差し替え、マップ画像の差し替えと GPS 基準点の調整を扱う。「出店者を追加」「店舗情報を更新」「マップを差し替え」「現在地がずれる」といった作業で使う。
---

# 会場マップの更新

店舗・施設のデータは `maps/src/data/shops.json`。対応表 `maps/data/<開催>/shops.tsv` と出店申込フォームの回答 CSV から `maps/tools/shops` で生成する。JSON は直接編集しない。

申込の CSV は電話番号や住所を含むので、**リポジトリに置かない**（デスクトップなど外に置いたまま `-apply` で指定する）。

## 店舗データの作成・更新

1. 対応表 `data/<開催>/shops.tsv`（タブ区切り、UTF-8）を編集する。1 行が 1 店舗・施設。

   | 列 | 内容 |
   |---|---|
   | `key` | マップ上の番号やステージ記号（`A` / `HQ` / `WC` など）。画像ファイル名にも使う |
   | `group` | `food` / `kitchencar` / `goods` / `facility` |
   | `name` | 表示する店名 |
   | `desc` | マップ画像の店舗一覧の説明 |
   | `app` | 申込の店名（完全一致）。申込が無ければ空 |
   | `text` | 説明文。空なら申込の販売メニュー、`-` なら `desc`、それ以外はその文（`\n` で改行） |
   | `links` | リンク。空なら申込の URL 欄から作る、`-` なら無し、それ以外は空白区切りの URL |
   | `rects` | **マップ画像（2000×522px）上**のタップ範囲 `x,y,幅,高さ`。複数あれば空白区切り |

   - 説明文は申込の販売メニューを使う。「なし」「物販希望」など主催者向けの回答は `text` を `-` にして一覧の説明にする。一部だけ主催者向けなら、整えた文を `text` に書く。
   - 申込の URL 欄の `@xxx` は Instagram のアカウントとして扱う。判断できないものは生成時に知らせるので、`links` に書く。

2. 生成する。

   ```powershell
   cd maps
   go run tools/shops/main.go -apply "<申込の CSV>" -list data/<開催>/shops.tsv -out src/data/shops.json
   ```

   標準エラーに次を出すので確認する。
   - `対応表に無い申込` … 不採用の申込か、`app` の店名が食い違っているもの
   - `URL にできない` / `リンクが複数` … `links` に正しい URL を書く

### 店舗画像

`maps/public/images/{key}/{n}.webp` に置く（n は 0 始まり）。`tools/shops` がディレクトリを見て `shops.json` の `images` に番号順で並べ、店舗パネルでは横スワイプで切り替えて表示する。画像が無い店は `noimage.png` を表示する。

申込の写真から作る手順:

1. Google Drive で申込の写真のフォルダを ZIP でダウンロードする（フォルダ名が長く Windows では解凍できないことがあるが、ツールが直接読むので解凍は不要）。
2. 店舗番号ごとに分ける。全ファイルは `work/drive`、店舗の写真は `work/images/{key}/{n}.{拡張子}` に置かれる（`work/` は Git に入れない）。

   ```powershell
   cd maps
   go run tools/images/main.go extract -zip "<ZIP>" -apply "<申込の CSV>" -list data/<開催>/shops.tsv
   ```

   ZIP には Drive の ID が無いので、アップロード時刻と回答時刻で対応づける（時差は自動で推定する）。`枚数が合わない` が出たら、その回答の写真を確認する。

3. 写真の向きを直してから webp 化する（横幅 1000。`{n}_{幅}x{高さ}.webp` ができる）。
   - webp 化では EXIF の向き（Orientation）が反映されず、スマホの写真が横に倒れる。JPEG は先に EXIF の向きを直した `{n}_rotate.jpg` を作って webp 化する（`install` は `{n}_rotate_{幅}x{高さ}.webp` を優先する）。
   - HEIC（`.heif`）も同様に向きが反映されないので、`tools/heicrot` で向きを直した `{n}_rotate.png` を作って webp 化する。EXIF の向きが実際と合っていない写真もあるので、配置後に一覧で目視確認する。
   - 中身が webp の写真（拡張子が jpg のものもある）は EXIF の読み取りに失敗するが、向きの情報を持たないのでそのままでよい。
4. 配置して JSON を生成し直す。webp が無い写真は元の画像をそのまま置く。

   ```powershell
   go run tools/images/main.go install
   go run tools/shops/main.go -apply "<申込の CSV>" -list data/<開催>/shops.tsv -out src/data/shops.json
   ```

### タップ範囲の確認

`Map.jsx` の `debug` state を一時的に `true` にすると、基準点と全店舗の矩形（青枠と key）がマップ上に描かれる。確認したら **`false` に戻してからコミットする**。

## マップ画像の差し替えと GPS 基準点

マップ画像は `maps/src/assets/map.jpg`。差し替えるときは次も合わせて直す。

1. `MapPoint.imageWidth` / `imageHeight` を新しい画像サイズにする。
2. `tops` / `bottoms` に基準点 `new Point(x, y, 緯度, 経度)` を並べる。
   - 上辺（`tops`）・下辺（`bottoms`）それぞれで、**画像上の位置と実際の GPS 座標の対応点**を複数取る。
   - 経度で隣り合う基準点の間を補間し、続いて緯度で上辺・下辺の間を補間して (x, y) を出す。基準点の外は端の区間を延長して計算し、画像の外になるときだけ現在地を表示しない。
   - 配列は **x の小さい順（経度の大きい順）**に並べる（`getBetweenPoint()` が隣接要素の経度範囲で探すため）。
3. 対応表の `rects` を新しい画像に合わせて取り直し、JSON を生成し直す。

現在地がずれるときは、ずれている区間の基準点を増やすか座標を見直す。実機では terminal アイコンのボタン（`#debugBtn`）でデバッグ表示を開くと、取得した緯度経度を確認できる。

## 確認と反映

```powershell
cd maps
npm run dev      # http://localhost:5173/maps/
npm run lint     # 警告 0 件が必須
```

本番への反映は `app/` への embed が必要（→ [fes-deploy](../fes-deploy/SKILL.md) の「会場マップを反映する」）。
