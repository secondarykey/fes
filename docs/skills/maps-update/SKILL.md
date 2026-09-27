---
name: maps-update
description: 会場マップ（maps/）の内容をフェスに合わせて更新する手順。店舗・ステージの追加や変更、店舗画像の差し替え、マップ画像の差し替えと GPS 基準点の調整を扱う。「出店者を追加」「店舗情報を更新」「マップを差し替え」「現在地がずれる」といった作業で使う。
---

# 会場マップの更新

データはすべて `maps/src/MapPoint.jsx` にハードコードされている。外部 API や CSV は無い。

## 店舗・ステージの追加・変更

`MapPoint.init()`（ステージ・本部・トイレなどの施設）または `MapPoint.initShops()`（出店者）に `Shop` を追加・編集する。

```js
this.shops.push(new Shop("23", "店名", `説明文
https://www.instagram.com/xxx/`,
    new Rect(407, 330, 30, 55)));
```

- `key` … 番号やステージ記号。画像ファイル名にも使う。
- `detail` … テンプレートリテラルで改行可。全角スペース（U+3000）も使える（ESLint の `no-irregular-whitespace` は文字列内を除外済み）。
- `Rect(x, y, width, height)` … **元画像（2395×647px）上のピクセル座標**。タップ判定に使う。
- 1 つの店舗が L 字などで複数の矩形にまたがるときは、**同じ key の `Shop` を複数追加**する（既存の `B` / `C` / `99` などと同じ）。
- 既存の並びはブロック（上部・キッチンカー・A〜L・下・右ライン・特殊）ごとにコメントで区切っている。同じ区切りに追加する。

### 店舗画像

`maps/public/images/{key}_0.webp` に置く（`/maps/images/{key}_0.webp` として参照される）。形式は webp。

### 矩形の位置合わせ

`Map.jsx` の `debug` state を一時的に `true` にすると、基準点と全店舗の矩形（青枠と key）がマップ上に描かれる。位置を合わせたら **`false` に戻してからコミットする**。

## マップ画像の差し替えと GPS 基準点

マップ画像は `maps/src/assets/map.jpg`。差し替えるときは次も合わせて直す。

1. `MapPoint.imageWidth` / `imageHeight` を新しい画像サイズにする。
2. `offsetTop` / `offsetBottom`（座標計算に使う上辺・下辺の y 座標）を決める。
3. `tops` / `bottoms` に基準点 `new Point(x, y, 緯度, 経度)` を並べる。
   - 上辺（`tops`）・下辺（`bottoms`）それぞれで、**画像上の x と実際の GPS 座標の対応点**を複数取る。
   - 経度で隣り合う基準点の間を補間し、続いて緯度で上辺・下辺の間を補間して (x, y) を出す。基準点の経度範囲の外にいると現在地は表示されない。
   - 配列は既存と同じく **x の小さい順（経度の大きい順）**に並べる（`getBetweenPoint()` が隣接要素の経度範囲で探すため）。
4. 全店舗の `Rect` を新しい画像に合わせて取り直す。

現在地がずれるときは、ずれている区間の基準点を増やすか座標を見直す。実機では terminal アイコンのボタン（`#debugBtn`）でデバッグ表示を開くと、取得した緯度経度を確認できる。

## 確認と反映

```powershell
cd maps
npm run dev      # http://localhost:5173/maps/
npm run lint     # 警告 0 件が必須
```

本番への反映は `app/` への embed が必要（→ [fes-deploy](../fes-deploy/SKILL.md) の「会場マップを反映する」）。
