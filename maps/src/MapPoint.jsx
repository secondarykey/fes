
/**
 * 画像に対する座標処理を行うクラス
 *
 * GPS の緯度・経度をマップ画像上の位置に変換する。変換は現地で記録した基準点
 * （src/data/gps.json。tools/gps.mjs で現地調査の JSON から作る）をなめらかにつなぐ薄板スプライン。
 * マップはデフォルメされているので、地点ごとに伸び縮みを変えられる方式にしている。
 * 基準点の外側はほぼ平行移動・回転・拡大（アフィン）で延長される。
 */

import shopData from './data/shops.json';
import gpsData from './data/gps.json';

//緯度 1 度あたりの距離（m）
const meterPerDegree = 111320;

//薄板スプラインの基底 r² log r（r は m）
const kernel = (r) => r === 0 ? 0 : r * r * Math.log(r);

/**
 * 連立一次方程式 A X = B を解く（部分ピボット付きのガウスの消去法）
 * @param {number[][]} A n×n
 * @param {number[][]} B n×m
 * @returns {number[][]} X n×m
 */
const solve = (A, B) => {
    const n = A.length;
    const a = A.map((row, i) => [...row, ...B[i]]);
    for (let c = 0; c < n; c++) {
        let pivot = c;
        for (let r = c + 1; r < n; r++) {
            if (Math.abs(a[r][c]) > Math.abs(a[pivot][c])) pivot = r;
        }
        [a[c], a[pivot]] = [a[pivot], a[c]];
        for (let r = 0; r < n; r++) {
            if (r === c) continue;
            const f = a[r][c] / a[c][c];
            for (let k = c; k < a[r].length; k++) a[r][k] -= f * a[c][k];
        }
    }
    return a.map((row, i) => row.slice(n).map((v) => v / a[i][i]));
}

class MapPoint {

    //元画像のサイズ
    static imageWidth = 2000;
    static imageHeight = 522;

    //現在地アイコンの大きさ
    static pointerSize = 48;

    //基準点（{ name, x, y, lat, lon }）
    static controls = [];
    //基準点の平均を原点にした m 単位の座標と、スプラインの重み
    static origin = null;
    static nodes = [];
    static weights = [];

    static shops = [];
    /**
     * マップデータの初期化
     * 何度呼ばれても重複しないよう、毎回作り直す
     */
    static init() {
        this.shops = [];
        this.initTransform();
        this.initShops();
    }

    /**
     * 基準点から変換（薄板スプライン）を作る
     * 基準点の数 + 3 元の連立一次方程式を解く。smoothing が大きいほど基準点を厳密には通らず、なめらかになる
     */
    static initTransform() {
        const pts = gpsData.points;
        const n = pts.length;
        this.controls = pts;

        const lat0 = pts.reduce((s, p) => s + p.lat, 0) / n;
        const lon0 = pts.reduce((s, p) => s + p.lon, 0) / n;
        this.origin = { lat0, lon0, k: Math.cos(lat0 * Math.PI / 180) };
        this.nodes = pts.map((p) => this.toMeter(p.lat, p.lon));

        const A = Array.from({ length: n + 3 }, () => new Array(n + 3).fill(0));
        const B = Array.from({ length: n + 3 }, () => [0, 0]);
        this.nodes.forEach(([e, m], i) => {
            this.nodes.forEach(([e2, m2], j) => {
                A[i][j] = kernel(Math.hypot(e - e2, m - m2)) + (i === j ? gpsData.smoothing : 0);
            });
            A[i][n] = A[n][i] = 1;
            A[i][n + 1] = A[n + 1][i] = e;
            A[i][n + 2] = A[n + 2][i] = m;
            B[i] = [pts[i].x, pts[i].y];
        });
        this.weights = solve(A, B);
    }

    //緯度・経度を、基準点の平均を原点にした東西・南北の m に直す
    static toMeter(lat, lon) {
        const o = this.origin;
        return [(lon - o.lon0) * meterPerDegree * o.k, (lat - o.lat0) * meterPerDegree];
    }

    /**
     * 店舗・施設の初期化
     * データは tools/shops で生成した src/data/shops.json
     */
    static initShops() {
        shopData.forEach((s) => {
            s.rects.forEach(([x, y, w, h]) => {
                this.shops.push(new Shop(s.key, s.name, s.detail, new Rect(x, y, w, h), s.images, s.group, s.links));
            });
        });
    }

    static drawDebugPoint(ctx) {
        this.controls.forEach((p) => {
            ctx.beginPath();
            ctx.arc(p.x, p.y, 8, 0, Math.PI * 2);
            ctx.fillStyle = "black";
            ctx.fill();
            ctx.stroke();
        });

        this.drawShops(ctx)
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

    /**
     * マップ上の点を取得
     * @param {*} lat
     * @param {*} lon
     * @returns 画像の外になる場合は null
     */
    static get = (lat, lon) => {
        const [e, m] = this.toMeter(lat, lon);
        const w = this.weights;
        const n = this.nodes.length;

        //アフィンの部分
        let x = w[n][0] + w[n + 1][0] * e + w[n + 2][0] * m;
        let y = w[n][1] + w[n + 1][1] * e + w[n + 2][1] * m;
        //基準点ごとの曲げ
        this.nodes.forEach(([e2, m2], i) => {
            const u = kernel(Math.hypot(e - e2, m - m2));
            x += w[i][0] * u;
            y += w[i][1] * u;
        });
        x = Math.trunc(x);
        y = Math.trunc(y);

        if (x < 0 || x >= this.imageWidth || y < 0 || y >= this.imageHeight) {
            return null;
        }
        return new Point(x, y, lat, lon);
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
            new Rect(p.x - size / 2, p.y - size / 2, size, size), ["BIRD/0.webp"])
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
    images = [];
    group = "";
    links = [];
    /**
     * @param {*} images public/images からの相対パス（"{key}/0.webp" など）
     * @param {*} group food / kitchencar / goods / facility（ハッちなどは空）
     * @param {*} links 店舗のリンク（Instagram など）
     */
    constructor(key, name, detail, rect, images = [], group = "", links = []) {
        this.key = key;
        this.name = name;
        this.detail = detail;
        this.rect = rect;
        this.images = images.map((p) => `/maps/images/` + p);
        this.group = group;
        this.links = links;
    }

    inside(x, y) {
        return this.rect.inside(x, y);
    }
}

export default MapPoint;