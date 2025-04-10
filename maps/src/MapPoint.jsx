
/**
 * 画像に対する座標処理を行うクラス
 * 
 * 画像のオフセット、各４点の座標を保持し、画像内での座標を計算する
 * 
 */

class MapPoint {

    //元画像のサイズ
    static imageWidth = 2395;
    static imageHeight = 647;
    //座標算定しない部分
    static offsetTop = 170;
    static offsetBottom = 600;

    static tops = [];
    static bottoms = [];

    static shops = [];
    /**
     * マップデータの初期化
     */
    static init() {
        this.tops.push(new Point(170, this.offsetTop, 34.9254303, 138.3806125));
        this.tops.push(new Point(320, this.offsetTop, 34.9253489, 138.380396));
        this.tops.push(new Point(570, this.offsetTop, 34.9252625, 138.3801465));
        this.tops.push(new Point(810, this.offsetTop, 34.9250932, 138.379661));
        this.tops.push(new Point(1260, this.offsetTop, 34.9249304, 138.3792077));
        this.tops.push(new Point(1320, this.offsetTop, 34.9248893, 138.3791032));
        this.tops.push(new Point(1770, this.offsetTop, 34.9247155, 138.3785479));
        this.tops.push(new Point(1940, this.offsetTop, 34.9246473, 138.3783819));

        this.bottoms.push(new Point(0, this.offsetBottom, 34.9260959, 138.3806125));
        this.bottoms.push(new Point(230, this.offsetBottom, 34.9259624, 138.380227));
        this.bottoms.push(new Point(670, this.offsetBottom, 34.92577, 138.3796955));
        this.bottoms.push(new Point(970, this.offsetBottom, 34.9256645, 138.3793818));
        this.bottoms.push(new Point(1330, this.offsetBottom, 34.9255352, 138.3790495));
        this.bottoms.push(new Point(1600, this.offsetBottom, 34.9254138, 138.3786924));
        this.bottoms.push(new Point(1780, this.offsetBottom, 34.9253461, 138.3785264));
        this.bottoms.push(new Point(1880, this.offsetBottom, 34.92528, 138.3783819));


        this.shops.push(new Shop("GATE", "入退場ゲート案内所", ``,
            new Rect(1760, 260, 180, 90)));
        this.shops.push(new Shop("A", "SEASIDE STAGE", `
`,
            new Rect(1325, 155, 300, 100)));
        this.shops.push(new Shop("B", "BAND STAGE", `
`,
            new Rect(880, 420, 55, 60)));
        this.shops.push(new Shop("B", "BAND STAGE", `
`,
            new Rect(770, 470, 160, 80)));
        this.shops.push(new Shop("C", "SOUND STAGE", `
`,
            new Rect(340, 190, 140, 70)));
        this.shops.push(new Shop("C", "SOUND STAGE", `
`,
            new Rect(330, 260, 60, 40)));

        this.shops.push(new Shop("D", "RECORD STAGE", `
`,
            new Rect(1130, 510, 60, 50)));
        this.shops.push(new Shop("D", "RECORD STAGE", `
`,
            new Rect(1100, 550, 150, 80)));

        this.shops.push(new Shop("HQ", "本部", ``,
            new Rect(930, 510, 90, 50)));
        this.shops.push(new Shop("BEER", "ドリンク", `ビールのみのお客様はこちらでご注文を承ります
ソフトドリンク、カクテルなどはBARの方で承ります。`,
            new Rect(937, 450, 80, 60)));
        this.shops.push(new Shop("WC", "トイレ", `紙がない時は本部までお知らせください`,
            new Rect(200, 430, 100, 120)));

        this.initShops();
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

    //横のリミット
    //Longitude 
    // 138.3783819 - 138.3806125
    //  この経度上に居ない場合、nullを返す

    // tops 内でどこにいるかを検索 （緯度が決定）
    // bottoms 内でどこにいるかを検索 (緯度が決定)
    // 双方の緯度の間にいない場合、nullを返す
    static getBetweenPoint = (arr, v) => {
        var rtn = [null, null];
        for (let idx = arr.length - 1; idx > 0; idx--) {
            let p1 = arr[idx];
            let p2 = arr[idx - 1];

            if (v >= p1.longitude && v <= p2.longitude) {
                rtn[0] = p1;
                rtn[1] = p2;
            }
        }
        return rtn;
    }

    static createBetweenLatitudePoint = (p1, p2, v) => {

        var v1 = p1.latitude;
        var v2 = p2.latitude;
        if (v > v2 || v < v1) {
            return null;
        }

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
        if (p1 === null) {
            return [null, null]
        }
        var tp = this.createBetweenLongitudePoint(p1, p2, v)

        var [p3, p4] = this.getBetweenPoint(this.bottoms, v);
        if (p3 === null) {
            return [null, null]
        }
        var bp = this.createBetweenLongitudePoint(p3, p4, v)
        return [tp, bp];
    }

    /**
     * マップ上の点を取得
     * @param {*} lat 
     * @param {*} lon 
     * @returns 
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

        if (p1 === null || p2 === null) {
            return null;
        }

        //上部と下部の線から緯度を割り出す
        var p = this.createBetweenLatitudePoint(p1, p2, lat)
        console.debug("point", p);

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

    static initShops() {

        var p;

        // フード 上部
        this.shops.push(new Shop("2", "キャプテンジャーク", `ジャークチキン ライスアンドピーズ ジャークチキンオーバーライス`,
            new Rect(565, 190, 30, 55)));

        this.shops.push(new Shop("3", "GREEDY TACOS", `TACOS、チキンオーバーライス、ナチョス、
https://www.instagram.com/greedy_tacos57?igsh=a3hvd3RldHJ4OWJm`,
            new Rect(600, 194, 30, 55)));

        this.shops.push(new Shop("4", "Enishi", `大串焼き　唐揚げ　フライドポテト　オムそば　ホルモン焼きそば　海鮮焼き　シャーピン`,
            new Rect(635, 197, 30, 55)));

        this.shops.push(new Shop("5", "TACOCAFE", `こだわりのたこやき
TACOCAFE.jp`,
            new Rect(668, 200, 30, 55)));

        this.shops.push(new Shop("6", "3 mile Cafe", `ロングポテト　チュロス等`,
            new Rect(700, 205, 30, 55)));

        this.shops.push(new Shop("7", "晴レノ日ノ", `バイン・ミー　フォー　無農薬自家菜園ホーリーバジルのガパオライス
@harenohi.no`,
            new Rect(730, 210, 30, 55)));

        this.shops.push(new Shop("8", "うまいもんでぶ屋", `魔法の粉をかけたフライドポテト、ふりふりポテト、肉巻きおにぎり、でぶサンド、チュロス、かき氷、トロピカルジュース等`,
            new Rect(764, 211, 30, 55)));

        this.shops.push(new Shop("9", "篠原商店", `鶏のからあげ、さといものからあげ、とり皮のからあげ
https://www.instagram.com/wasouzaishinoharasyouten/`,
            new Rect(795, 214, 30, 55)));

        this.shops.push(new Shop("10", "Mr.Potato Farm", `フレンチフライ・サツマイモチップス・大学芋
Instagram @mr.potatofarm`,
            new Rect(828, 220, 30, 55)));

        this.shops.push(new Shop("11", "テキーラダイナー", `ハンバーガー、ホットドッグ、ステーキ
https://tequilasdiner.owst.jp/`,
            new Rect(861, 220, 30, 55)));

        this.shops.push(new Shop("12", "B.B.garden", `カレー`,
            new Rect(893, 220, 30, 55)));

        this.shops.push(new Shop("13", "29CHORI veg+", `静岡もつ焼きそば、ホルモン焼き
https://www.instagram.com/gido_29chori`,
            new Rect(925, 220, 30, 55)));

        this.shops.push(new Shop("14", "スレイジースタンド", `ポテトボール（新感覚　揚げじゃがバター）、オリジナルドリンク（フルーツエイド他）
https://www.instagram.com/srazystand?igsh=MWxwcWgyOW96dmVrYw==`,
            new Rect(959, 220, 30, 55)));

        this.shops.push(new Shop("15", "MIKAKOFFEE", `朝霧牛乳のカフェラテ、コーヒー、自家製シロップのドリンク各種、自家製ティラミス、朝霧牛乳のプリン各種、焼き菓子各種、自家焙煎コーヒー豆
https://www.instagram.com/mikakoffee/profilecard/?igsh=MTRoNWs5cGg0eGNneA==`,
            new Rect(994, 220, 30, 55)));

        this.shops.push(new Shop("16", "安寧", `長岡式酵素玄米、オーガニックドリンク、燻製各種
https://www.instagram.com/annei_shizuoka?igsh=MThsOWMyZDg1c3VyaA%3D%3D&utm_source=qr`,
            new Rect(1028, 220, 30, 55)));

        this.shops.push(new Shop("17", "BLOOM", `コーヒー、ハンバーガーなど
bloom_work.fitness`,
            new Rect(1062, 220, 30, 55)));
        this.shops.push(new Shop("18", "串焼亭", `静岡県産黒毛和牛牛串、はらみ串、牛タン串、かき氷`,
            new Rect(1096, 218, 30, 55)));
        this.shops.push(new Shop("19", "T’s green", `揚げ団子、かりんとう饅頭、抹茶ラテ等
@ts.green_2013`,
            new Rect(1130, 215, 30, 55)));

        this.shops.push(new Shop("20", "とらさんキッチン", `富士山焼きそば、マーボー焼きそば、燻製セット、炊き込みご飯、惣菜各種
torasankitchen78@instagram.jp`,
            new Rect(1165, 213, 30, 55)));
        this.shops.push(new Shop("21", "caffe kofka", `ドリップコーヒー、朝霧牛乳のカフェラテ、季節のフルーツシロップドリンク、ソフトドリンク
https://www.instagram.com/caffe_kofka/`,
            new Rect(1198, 212, 30, 55)));
        this.shops.push(new Shop("22", "ふわいろ（元わたじろう）", `わたあめ（わたあめは許可証は必要ないそうです）
https://www.instagram.com/kiyumi0013`,
            new Rect(1230, 210, 30, 55)));





        //キッチンカー
        this.shops.push(new Shop("42", "LA BALENA VOLANTE", `薪窯で焼くピッツァ、コーヒーその他ソフトドリンク類
https://www.instagram.com/tobekujira/`,
            new Rect(270, 295, 55, 60)));

        this.shops.push(new Shop("43", "池めん清水町店", `台湾まぜそば、豚骨らーめんなど`,
            new Rect(230, 355, 55, 60)));


        this.shops.push(new Shop("44", "ハワイ食堂LeaLea", `静岡わさびバーガー、他バーガー、モチコチキン（ハワイのからあげ）、ポテト等
https://lealea.crayonsite.com/`,
            new Rect(295, 370, 70, 50)));

        this.shops.push(new Shop("45", "Pomaikai", `クレープ`,
            new Rect(310, 420, 70, 40)));

        this.shops.push(new Shop("46", "たこ助", `チョコバナナ　串焼　ポテト　チュロス　丼`,
            new Rect(330, 460, 70, 40)));

        this.shops.push(new Shop("47", "いちご実　加藤農園", `いちごスイーツ
ichigomi.com`,
            new Rect(330, 510, 70, 50)));

        this.shops.push(new Shop("48", "48DINER", `スライダーバーガー、チキンオーバーライス、プルコギドッグ
https://www.instagram.com/48diner`,
            new Rect(405, 470, 40, 70)));

        this.shops.push(new Shop("49", "tomosfactory", `ロングポテト チュリトス カキ氷`,
            new Rect(450, 470, 40, 70)));

        this.shops.push(new Shop("50", "中嶋養蜂", `蜂蜜　ハニーレモネード　チュロス　ポテト
@y.nakajimayoho`,
            new Rect(495, 470, 40, 70)));


        this.shops.push(new Shop("51", "enwa縁和", `ローストビーフサンド・ローストポークサンド・スモーキーチキンサンド・ダブルハムチーズサンドのホットサンド フライドポテト かき氷 静岡県産レモンをじっくり煮込んだ レモネード&レモンスカッシュ クリーミーいちごミルク キャラメルマキアート ウインナーコーヒー等`,
            new Rect(540, 470, 40, 70)));


        this.shops.push(new Shop("52", "手羽先おおむら", `焼き鳥（手羽先、鶏もも焼き）モツ煮
https://www.instagram.com/tebasaki_oomura/`,
            new Rect(585, 470, 40, 70)));


        p = MapPoint.get(34.9256546, 138.3796935);
        this.shops.push(new Shop("53", "JAH KITCHEN", `ジャマイカンフード、おつまみ`,
            new Rect(630, 490, 75, 60)));






        //A ブロック
        this.shops.push(new Shop("23", "ポテイト王国", `ロングポテト　チュロス　チーズボール`,
            new Rect(407, 330, 30, 55)));

        this.shops.push(new Shop("24", "ぽっぽ屋", `広島焼き オムそば ロングポテト 唐揚げ フランク
https://instagram.com/umaimono.yatai.poppoya?igshid=MzRlODBiNWFlZA==`,
            new Rect(442, 330, 30, 55)));

        this.shops.push(new Shop("25", "ふわとろたこ焼き　あじたま", `たこ焼き
https://www.ajitama-takoyaki.com/`,
            new Rect(407, 388, 30, 55)));

        this.shops.push(new Shop("26", "保科精肉店", `からあげ 炭火焼きチャーシュー串 ロングポテト ゆず蜜ドリンク
https://www.instagram.com/hoshina_seinikuten/profilecard/?igsh=MXhic3RiYjhoNTFjcw==`,
            new Rect(442, 388, 30, 55)));


        //B ブロック
        this.shops.push(new Shop("27", "一新食堂", `ラーメン、お弁当、お惣菜、さつまいもチップス、レモンネード、おでん`,
            new Rect(506, 330, 30, 55)));

        this.shops.push(new Shop("28", "10DER HOUSE", `スモークチキンレッグ・ガーリックシュリンプ・ノンアルコールカクテル
https://www.instagram.com/10derhouse/profilecard/?igsh=MWJ6d2k3dHluenMxcA==`,
            new Rect(541, 330, 30, 55)));

        this.shops.push(new Shop("29", "ル・ジャルダン・ゴロワ", `ハーブグリルソーセージ、ハムチーズクレープ、ワイン、レモネード、キッシュ、チーズケーキ、タルト、シューケットラスク、マシュマロ`,
            new Rect(506, 385, 30, 55)));

        this.shops.push(new Shop("30", "マサイの風", `京都の山里で自家焙煎した珈琲、メインブレンドの「マサイの風」、他7種類の自家焙煎珈琲。
https://www.instagram.com/danchou.lion/`,
            new Rect(541, 385, 30, 55)));

        //C ブロック
        this.shops.push(new Shop("54", "Tequila Cream Trade Mark", `アパレルブランド
https://www.tequila-cream.com`,
            new Rect(605, 290, 30, 55)));

        this.shops.push(new Shop("55", "Shop standard", `オールハンドメイドのBAGやCAP、ウェアを制作しています！今回は刺繍用のミシンを使い、その場で刺繍を施すWORK
SHOPも行います。
https://www.instagram.com/standard_81/profilecard/?igsh=NXF3MDJjeDlhdnNi`,
            new Rect(640, 290, 30, 55)));

        this.shops.push(new Shop("56", "ETHNIC TOKYO", `アパレル・アフリカンバスケット・カレンシルバーなど
https://ethnictokyo.com`,
            new Rect(605, 345, 30, 55)));

        this.shops.push(new Shop("57", "Toko Ramai", `アウトドアウェア、アウトドアグッズ、雑貨、ガラポン
https://www.instagram.com/tokoramai2`,
            new Rect(640, 345, 30, 55)));


        //D ブロック
        this.shops.push(new Shop("58", "ATELIER / m-no", `リメイクアパレル
https://instagram.com/atelier_mno/`,
            new Rect(700, 290, 30, 55)));

        this.shops.push(new Shop("59", "ｅｎ", `オーガニックウエアー販売
https://www.instagram.com/en_shizuoka/`,
            new Rect(733, 290, 30, 55)));

        this.shops.push(new Shop("60", "流木ジョニー／古着ジェリー", `流木雑貨、古着とレゲエイベント用エスニック雑貨などの販売。
https://@ryubokujohnny`,
            new Rect(700, 345, 30, 55)));

        this.shops.push(new Shop("61", "HAE. vintage with Lab. Paper Dripper", `海外古着とコーヒーと焼き菓子
https://www.instagram.com/hae.vintage/
https://www.instagram.com/lab.paperdripper/`,
            new Rect(733, 345, 30, 55)));




        //E ブロック
        this.shops.push(new Shop("31", "ひので(青葉おでん街)", `静岡おでん(盛り合わせ5種) 特製おでん出汁おにぎり ホルモン煮込み 炭火焼鶏 牛ハラミ串 フランクフルト 静岡茶 コカコーラ`,
            new Rect(997, 338, 30, 55)));

        this.shops.push(new Shop("32", "最高餃子", `餃子、ホットク
https://www.instagram.com/saikou.gyouza315?igsh=MTg2ZGg1dGdzeHk3NA%3D%3D&utm_source=qr`,
            new Rect(1030, 338, 30, 55)));

        this.shops.push(new Shop("33", "八ヶ岳スモーク", `スモーク全部盛り（スモークチキン、スモークベーコン、ソーセージ）`,
            new Rect(997, 395, 30, 55)));

        this.shops.push(new Shop("34", "Fragrant", `コーヒー
https://www.instagram.com/fragrant_coffee_by_tls`,
            new Rect(1030, 395, 30, 55)));



        //F ブロック
        this.shops.push(new Shop("35", "グランプールテラス静波", `ホットサンド、ポテト、レモネード
https://www.instagram.com/glampool_shizunami`,
            new Rect(1100, 292, 59, 35)));

        this.shops.push(new Shop("36", "オールドソーコ", `ペンネアラビアータ 鴨肉のスモーク 鴨マヨピザ ガーリックトースト スーコン6種 沼津レモネード、沼津レモンスカッシュ、アイスコーヒー、チャイ、ラッシー、いちごミルク
old.soco`,
            new Rect(1162, 272, 30, 55)));

        this.shops.push(new Shop("37", "ベルラッテ", `豚角煮のちまき、藤枝高級抹茶のクレープ、ドリンク、らくがきせんべい`,
            new Rect(1100, 330, 59, 35)));

        this.shops.push(new Shop("38", "mario coffee labo", `わくわくするこだわりのドリップコーヒーです。ご注文を受けてその場で抽出いたします。少量の焼き豆とドリップバック。`,
            new Rect(1162, 330, 30, 55)));


        //G ブロック
        this.shops.push(new Shop("62", "アトリエクエスト", `オリジナル木材雑貨、マクラメ作品、七宝作品、パラコード作品、レジン作品等
https://www.create-life.jp/`,
            new Rect(1227, 272, 30, 55)));

        this.shops.push(new Shop("63", "ROTTENING", `レザー、レジン製アクセサリー、着物小物、キッズアクセサリー
https://www.instagram.com/rotten_ing/`,
            new Rect(1260, 272, 30, 55)));

        this.shops.push(new Shop("64", "yuni", `アクセサリー（ピアス、イヤリング）
https://www.instagram.com/yuni__0223/profilecard/?igsh=MXM3aDR3NndlamsxZw==`,
            new Rect(1227, 330, 30, 55)));

        this.shops.push(new Shop("65", "sail accessory", `ミニチュアのドリンクカップに毛糸や好きなパーツを１つ１つ選んで詰め込んで、オリジナルのキーホルダーを作ります
https://www.instagram.com/sail.accessory?igsh=YnBld21pZ2VsZnl2&utm_souらrce=qr`,
            new Rect(1260, 330, 30, 55)));



        //H ブロック
        this.shops.push(new Shop("66", "Freedom&Hope", `羊毛フェルト・アフリカ布雑貨他
https://www.instagram.com/freedom_3_3_hope`,
            new Rect(1094, 372, 30, 55)));

        this.shops.push(new Shop("67", "AYANO knitting", `あみぐるみ・ニット小物・雑貨
https://ayanoknitting.com/`,
            new Rect(1127, 372, 30, 55)));

        this.shops.push(new Shop("68", "MUSUBI★Store", `麻たわし、麻紐雑貨、精麻飾、わらじ、石笛等
https://musubistore.thebase.in/`,
            new Rect(1094, 428, 30, 55)));

        this.shops.push(new Shop("69", "YOPOOHRU", `オリジナルイラストプリントウェア＆グッズ・カッティングステッカー
https://www.instagram.com/yopoohru/`,
            new Rect(1127, 428, 30, 55)));



        // Iブロック
        this.shops.push(new Shop("70", "yatri", `インド伝統染めの手ぬぐい、スカーフ、ターバンなど布雑貨
https://www.instagram.com/yatri_silkroad/`,
            new Rect(1180, 380, 38, 49)));

        this.shops.push(new Shop("71", "武器あります。", `中世ヨーロッパで実際使用されていた武器のモチーフ、ドラゴンやグ+リフォンなど幻、(獣のモチーフシルバーアクセサリーです。
爪や牙など細部まで表現した作品
天然石水晶を使用した世界に一つの一点ものも
ドラゴンベビーや猫やコウモリのイヤーカフなど女性にも人気です。
https://minne.com/@m2sick　Xフォロワー数4000以上 https://x.com/m2sicksilver`,
            new Rect(1220, 384, 55, 45)));

        this.shops.push(new Shop("72", "HIMALAI TRiP MARKET & Sala Mehndi", `タイダイ曼荼羅染め&オリジナルデザインTシャツ、雑貨販売。メヘンディ(ヘナタトゥー)
            https://himalai-trip-market.stores.jp/`,
            new Rect(1188, 430, 30, 55)));

        this.shops.push(new Shop("73", "Cherie CoCo〜シェリーココ〜", `トルコモザイクガラスワークショップと
            うぉーたーびーずアクアリウムワークショップ
            ハンドメイド雑貨販売
            https://www.instagram.com/cherie.coco77?igsh=MWo0dGZlbjhhaHkyMg==`,
            new Rect(1220, 430, 34, 55)));




        //J ブロック
        this.shops.push(new Shop("74", "c.blue", `オーシャンレジン、マクラメ編みのインテリア雑貨、アクセサリー販売
            https://www.instagram.com/d.s.k.666.masonlysurf`,
            new Rect(1290, 372, 75, 55)));

        this.shops.push(new Shop("75", "タソガレリリー・and g", `カラーセラピー、占い、ハンドメイドアクセサリー`,
            new Rect(1295, 429, 30, 55)));

        this.shops.push(new Shop("76", "Qyuncandle(キュンキャンドル", `グラデーションキャンドル作り体験
            https://www.instagram.com/qyuncandle`,
            new Rect(1328, 429, 30, 55)));



        //Kブロック
        this.shops.push(new Shop("77", "KSL RESIN ART", `海のレジンアート販売、ベアアートのワークショップ
            https://www.instagram.com/ksl_resinart/profilecard/?igsh=MXRnMGtiYXZkaDU0eg==`,
            new Rect(1390, 368, 32, 55)));

        this.shops.push(new Shop("78", "MOMONA Mermaid", `ハワイアン雑貨、ハワイアンファブリックのハンドメイド布小物`,
            new Rect(1425, 368, 32, 55)));

        this.shops.push(new Shop("79", "Merry", `アクアビーズドームワークショップ　あみぐるみワークショップ
            @merry.miyuki1007shop`,
            new Rect(1390, 426, 32, 55)));

        this.shops.push(new Shop("80", "花十色-mimi-", `風鈴、ドアベル、キーホルダーのワークショップ
            https://instagram.com/hanatoiro_mimi_`,
            new Rect(1425, 426, 32, 55)));



        // L ブロック
        this.shops.push(new Shop("81", "coconuts", `イニシャルキーホルダー、ワイヤーリングキーホルダーワークショップ
            @coconuts.handmade66`,
            new Rect(1490, 368, 32, 55)));

        this.shops.push(new Shop("82", "maeum candle", `アロマジェルキャンドル作り、フロートキャンドル作り、キャンドルすくい
            https://www.instagram.com/maeum.candle`,
            new Rect(1525, 368, 32, 55)));

        this.shops.push(new Shop("83", "トコトコっち", `スクイーズのデコレーション体験と端材を用いたロボットオブジェ作り体験
            https://www.instagram.com/tocotocotti/`,
            new Rect(1490, 424, 32, 55)));

        this.shops.push(new Shop("84", "庵寿玲海Amy", `ミニ水族館ボトリウム`,
            new Rect(1525, 424, 32, 55)));






        //下 ブロック
        this.shops.push(new Shop("39", "極東アンティール", `ジーラポーク、豆カレーorジャークチキン（他店様とのカブリの少ない方）
https://far-east-antilles.com`,
            new Rect(1030, 470, 30, 60)));

        this.shops.push(new Shop("40", "SHANTY", `jamaica料理`,
            new Rect(1065, 485, 30, 60)));

        this.shops.push(new Shop("85", "HOME BEAT", `ミニサウンド・システム販売`,
            new Rect(1098, 495, 30, 60)));

        this.shops.push(new Shop("86", "CORNERSHOP", `レコード、CD販売`,
            new Rect(1193, 495, 30, 60)));

        //RECORD STAGE

        this.shops.push(new Shop("41", "Ailie の野生コーヒー", `自家焙煎コーヒー`,
            new Rect(1229, 495, 30, 60)));

        this.shops.push(new Shop("87", "ウクレレ移動販売＆出張ワークショップ【イコレレ】", `ウクレレや小物楽器の販売　ウクレレ演奏のワークショップ
https://note.com/iko15ico`,
            new Rect(1263, 495, 30, 60)));

        this.shops.push(new Shop("88", "陶工房修", `陶器、ワークショップ`,
            new Rect(1297, 492, 30, 60)));

        this.shops.push(new Shop("89", "Sachiのメタルクリエイト", `アイアン製品(インテリア雑貨)、スツール等アイアン製品の販売。アイアンアートの展示
https://www.instagram.com/sach_imetal/profilecard/?igsh=MXB6aXYxbG1yeG02eA==`,
            new Rect(1330, 490, 30, 60)));

        this.shops.push(new Shop("90", "かなせっけん", `せっけん、豆菓子、木工製品、お茶などの販売
https://kananoie.shop-pro.jp/`,
            new Rect(1362, 490, 30, 60)));

        this.shops.push(new Shop("91", "meChill整体salon", `マッサージと足指裏マッサージ
https://www.instagram.com/mechill_sangobiyou.seitai/profilecard/?igsh=MWNsdmNla3Z2NmF1MA==`,
            new Rect(1395, 490, 30, 60)));

        this.shops.push(new Shop("92", "sowaca", `首・肩マッサージ、フットマッサージ、ベットでマッサージ
https://www.instagram.com/p/Cy2Oc_vyvGB/?igsh=MWRldHFxZHcwN2ViMg==`,
            new Rect(1425, 487, 30, 60)));

        this.shops.push(new Shop("93", "spika永田町", `ハンドトリート20分、お肌の分析30分
https://www.instagram.com/pola_spika/profilecard/?igsh=MWg3bjF1Zzh6bTc5eg==`,
            new Rect(1458, 485, 30, 60)));

        this.shops.push(new Shop("94", "足流療術", `足流療術`,
            new Rect(1492, 485, 30, 60)));

        this.shops.push(new Shop("95", "FUMIHEALING:GROUP", `深リンパケア/スピリチュアル鑑定
ALL10分1000円
https://www.instagram.com/fumihealing_deeplymph_school?igsh=NXVkd3BxZW4zYWtk`,
            new Rect(1524, 485, 30, 60)));

        this.shops.push(new Shop("96", "精神と時の部屋", `テントサウナ`,
            new Rect(1560, 485, 30, 60)));



        //
        //右ライン
        //
        this.shops.push(new Shop("97", "明治安田", `健康チェックブース`,
            new Rect(1590, 462, 56, 43)));

        this.shops.push(new Shop("98", "toooa+", `うぉーたーぷにぷにワークショップ
ぷにぷに石鹸ワークショップ
ぷにぷにすくい
光るおもちゃくじ
@toooa.plus`,
            new Rect(1597, 432, 60, 35)));

        this.shops.push(new Shop("99", "MRK Caricatures", `その場で似顔絵を描きます
https://instagram.com/mrk_caricatures`,
            new Rect(1610, 390, 50, 40)));
        this.shops.push(new Shop("99", "MRK Caricatures", `その場で似顔絵を描きます
https://instagram.com/mrk_caricatures`,
            new Rect(1660, 410, 20, 20)));

        this.shops.push(new Shop("100", "バルーンゆっぴー", `バルーン販売`,
            new Rect(1620, 365, 40, 25)));
        this.shops.push(new Shop("100", "バルーンゆっぴー", `バルーン販売`,
            new Rect(1660, 372, 35, 35)));

        this.shops.push(new Shop("101", "用宗を楽しくする会", `用宗＆広野PRブース、食品晩梅`,
            new Rect(1650, 340, 40, 25)));
        this.shops.push(new Shop("101", "用宗を楽しくする会", `用宗＆広野PRブース、食品晩梅`,
            new Rect(1690, 350, 35, 40)));

        this.shops.push(new Shop("102", "縁日処", `射的、スーパーボールすくい
https://www.instagram.com/ennichi_docoro?igsh=cGlmZ2V2YmljMWpm&utm_source=qr`,
            new Rect(1675, 315, 40, 25)));
        this.shops.push(new Shop("102", "縁日処", `射的、スーパーボールすくい
https://www.instagram.com/ennichi_docoro?igsh=cGlmZ2V2YmljMWpm&utm_source=qr`,
            new Rect(1715, 315, 45, 55)));




        //
        //特殊
        //
        this.shops.push(new Shop("1", "official BAR", `ビール、アルコール類`,
            new Rect(735, 420, 145, 50)));

        this.shops.push(new Shop("103", "フィリップモリス　アイコス", `アイコス紹介/販促`,
            new Rect(170, 200, 110, 90)));
        this.shops.push(new Shop("104", "4541 Village", `杉本青果 (野菜販売)
デルクス 39 (衣類販売)
運動するところ (マッサージ)
大黒屋 (駄菓子販売)
RAI 4 (ベーコン等)
4541 (雑貨販売)`,
            new Rect(620, 400, 110, 90)));
        this.shops.push(new Shop("105", "Harbor Art Billage", `よいこ共和国 (無農薬栽培コーヒー、 豆乳チャイ)
栄養灯台堂 (キャンドルワークショップ、販売)
黄桜じーじのガラス玉とけん玉屋さん (けん玉販売やWS)
Green Island (おやつ、雑貨販売)
KUKU (信州産りんごパイ、 焼き菓子、 三年番茶)
東静岡アートラボ (お絵描きワークショップ)
HAPPY MONSTA (リサイクルペットボトル食器アート)
お祭り【酒-sai-】 (射的、 型抜き、ふんどし販売)`,
            new Rect(1580, 225, 180, 90)));
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