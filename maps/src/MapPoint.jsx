
/**
 * 画像に対する座標処理を行うクラス
 * 
 * 画像のオフセット、各４点の座標を保持し、画像内での座標を計算する
 * 
 */

import shopData from './data/shops.json';

class MapPoint {

    //元画像のサイズ
    static imageWidth = 2000;
    static imageHeight = 522;

    //現在地アイコンの大きさ
    static pointerSize = 48;

    static tops = [];
    static bottoms = [];

    static shops = [];
    /**
     * マップデータの初期化
     * 何度呼ばれても重複しないよう、毎回作り直す
     */
    static init() {
        this.tops = [];
        this.bottoms = [];
        this.shops = [];

        //上辺・下辺の基準点 (画像上の x, y, 緯度, 経度)
        //x の小さい順（経度の大きい順）に並べる
        this.tops.push(new Point(166, 152, 34.9254303, 138.3806125));
        this.tops.push(new Point(287, 152, 34.9253489, 138.380396));
        this.tops.push(new Point(490, 152, 34.9252625, 138.3801465));
        this.tops.push(new Point(685, 152, 34.9250932, 138.379661));
        this.tops.push(new Point(1049, 151, 34.9249304, 138.3792077));
        this.tops.push(new Point(1098, 151, 34.9248893, 138.3791032));
        this.tops.push(new Point(1463, 151, 34.9247155, 138.3785479));
        this.tops.push(new Point(1600, 151, 34.9246473, 138.3783819));

        this.bottoms.push(new Point(49, 496, 34.9260959, 138.3806125));
        this.bottoms.push(new Point(236, 495, 34.9259624, 138.380227));
        this.bottoms.push(new Point(592, 495, 34.92577, 138.3796955));
        this.bottoms.push(new Point(835, 495, 34.9256645, 138.3793818));
        this.bottoms.push(new Point(1127, 494, 34.9255352, 138.3790495));
        this.bottoms.push(new Point(1346, 494, 34.9254138, 138.3786924));
        this.bottoms.push(new Point(1492, 494, 34.9253461, 138.3785264));
        this.bottoms.push(new Point(1573, 494, 34.92528, 138.3783819));

        this.initShops();
    }

    /**
     * 店舗・施設の初期化
     * データは tools/shops で生成した src/data/shops.json
     */
    static initShops() {
        shopData.forEach((s) => {
            var detail = s.detail;
            if (s.links.length > 0) {
                detail += (detail === "" ? "" : "\n\n") + s.links.join("\n");
            }
            s.rects.forEach(([x, y, w, h]) => {
                this.shops.push(new Shop(s.key, s.name, detail, new Rect(x, y, w, h)));
            });
        });
    }

    static drawDebugPoint(ctx) {
        this.drawDebugArray(ctx, this.tops);
        this.drawDebugArray(ctx, this.bottoms);

        this.drawShops(ctx)
    }

    static drawDebugArray(ctx, arr) {
        arr.forEach((p) => {
            ctx.beginPath();
            ctx.arc(p.x, p.y, 8, 0, Math.PI * 2);
            ctx.fillStyle = "black";
            ctx.fill();
            ctx.stroke();
        });
    }
    static drawShops(ctx) {
        ctx.font = "14px Araial";
        ctx.strokeStyle = 'blue';
        this.shops.forEach((s) => {
            var rect = s.rect;
            ctx.beginPath();
            ctx.rect(rect.x, rect.y, rect.width, rect.height);
            ctx.stroke();
            ctx.fillText(s.key, rect.x, rect.y2())
        })
    }

    // tops / bottoms 内で経度 v を挟む2点を返す [経度の小さい点, 大きい点]
    // 配列は経度の大きい順（画像の左から右）に並んでいる
    // 基準点の範囲外は端の区間を延長して使う
    static getBetweenPoint = (arr, v) => {
        for (let idx = 1; idx < arr.length; idx++) {
            if (v >= arr[idx].longitude) {
                return [arr[idx], arr[idx - 1]];
            }
        }
        var last = arr.length - 1;
        return [arr[last], arr[last - 1]];
    }

    // 上辺の点 p1 と下辺の点 p2 の間を緯度 v で補間する（範囲外は延長）
    static createBetweenLatitudePoint = (p1, p2, v) => {

        var v1 = p1.latitude;
        var v2 = p2.latitude;

        var distance = v2 - v1;
        var remain = v - v1;
        var odds = remain / distance;
        var dY = Math.abs(p1.y - p2.y);
        var y = Math.trunc(p1.y + (dY * odds));

        var dX = Math.abs(p1.x - p2.x);
        var num = Math.trunc(dX * odds);
        if (p1.x > p2.x) {
            num *= -1;
        }
        var x = p1.x + num

        //Y を割り出して、直線の点としてXを割り出す
        //var m = (p2.y - p1.y) / (p2.x - p2.x);
        //var x = ((y - p2.y) / m) + p2.x;

        return new Point(x, y, v, p1.longitude);
    }

    static createBetweenLongitudePoint = (p1, p2, v) => {

        var v1 = p1.longitude;
        var v2 = p2.longitude;
        var distance = v2 - v1;
        var remain = v - v1;

        //位置の割合を取得
        var odds = remain / distance;

        var d = p1.x - p2.x;
        //逆側からの値でXは算出
        var x = Math.trunc(p2.x + ((1 - odds) * d));
        //Yは必ず同じ
        var y = p1.y

        //同じ倍率の値を取得
        var latD = p2.latitude - p1.latitude;
        var lat = p1.latitude + (latD * odds);
        // 求めている位置なのでそのまま利用
        var lon = v;
        return new Point(x, y, lat, lon);
    }

    // 緯度の割合を出す
    static getLongitudePoint = (v) => {
        var [p1, p2] = this.getBetweenPoint(this.tops, v);
        var tp = this.createBetweenLongitudePoint(p1, p2, v)

        var [p3, p4] = this.getBetweenPoint(this.bottoms, v);
        var bp = this.createBetweenLongitudePoint(p3, p4, v)
        return [tp, bp];
    }

    /**
     * マップ上の点を取得
     * @param {*} lat
     * @param {*} lon
     * @returns 画像の外になる場合は null
     */
    static get = (lat, lon, ctx) => {

        //経度から定点を利用して、マップの上部、下部の点を算出
        var [p1, p2] = this.getLongitudePoint(lon)

        if (ctx !== undefined) {
            ctx.beginPath();
            ctx.arc(p1.x, p1.y, 4, 0, Math.PI * 2);
            ctx.fillStyle = "red";
            ctx.fill();
            ctx.stroke();
            ctx.beginPath();
            ctx.arc(p2.x, p2.y, 4, 0, Math.PI * 2);
            ctx.fillStyle = "red";
            ctx.fill();
            ctx.stroke();
        }

        //上部と下部の線から緯度を割り出す
        var p = this.createBetweenLatitudePoint(p1, p2, lat)
        console.debug("point", p);

        if (p.x < 0 || p.x >= this.imageWidth || p.y < 0 || p.y >= this.imageHeight) {
            return null;
        }
        return p;
    }

    static getShop = (x, y) => {
        var shop = null;
        this.shops.forEach((s) => {
            if (s.inside(x, y)) {
                shop = s;
            }
        })
        return shop;
    }
    
    static getBird = (coords) => {
        if ( coords === undefined ) return null;
      var p = MapPoint.get(coords.latitude,coords.longitude);
      if ( p === null ) {
        return null;
      }
      //アイコンは p を中心に描画している
      var size = MapPoint.pointerSize;
      return new Shop("BIRD", "ハッち", `Humming Bird 非公式キャラ のハッちです。`,
            new Rect(p.x - size / 2, p.y - size / 2, size, size))
    }

}

export class Point {
    x = 0
    y = 0
    latitude = 0
    longitude = 0
    constructor(x, y, latitude, longitude) {
        this.x = x;
        this.y = y;
        this.latitude = latitude;
        this.longitude = longitude;
    }
}

export class Rect {
    x = 0
    y = 0
    width = 0
    height = 0
    constructor(x, y, w, h) {
        this.x = x;
        this.y = y;
        this.width = w;
        this.height = h;
    }
    x2() {
        return this.x + this.width;
    }
    y2() {
        return this.y + this.height;
    }

    inside(x, y) {
        if (x < this.x || y < this.y) {
            return false;
        }
        if (x <= this.x2() && y <= this.y2()) {
            return true;
        }
        return false;
    }
}

export class Shop {
    key = ""
    name = "No Name"
    detail = "";
    rect = undefined;
    constructor(key, name, detail, rect) {
        this.key = key;
        this.name = name;
        this.detail = detail;
        this.rect = rect;

        this.image = this.url();
    }

    url() {
        var url = `/maps/images/` + this.key + "_0.webp";
        return url
    }

    inside(x, y) {
        return this.rect.inside(x, y);
    }
}

export default MapPoint;