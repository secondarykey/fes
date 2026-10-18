import { useEffect, useRef, useState } from 'react';
import map from './assets/map.jpg'
import point from './assets/point.png'
import useWindowSize from './Window';
// eslint-disable-next-line no-unused-vars -- Point はデバッグ描画で使うため残す
import MapPoint, { Point } from './MapPoint';

const iconSize = MapPoint.pointerSize;
//鳥と方向の三角を描く範囲（中心からの半分の幅）。前のフレームの分を背景で消すため
const drawHalf = MapPoint.pointerMargin;

//この距離（px）以上動かしたらクリックではなくドラッグとみなす
const dragThreshold = 5;

const Map = (props) => {

  // eslint-disable-next-line no-unused-vars -- setDebug はデバッグ表示切り替え用に残す
  const [debug, setDebug] = useState(false);

  const [fit, setFit] = useState(false);
  const [scale, setScale] = useState(1);
  const [mapStyle, setMapStyle] = useState({});

  const winSize = useWindowSize();
  const view = useRef(undefined);

  //現在地のマップ上の点（MapPoint.locate の戻り値。未取得・遠すぎるときは null）
  const position = useRef(null);
  //現在地ボタンが押された。次に位置が届いたら、鳥が画面の真ん中に来るように動かす
  const centerPending = useRef(false);

  //マウスでドラッグ中の状態（開始位置、開始時のスクロール位置、動かしたか）
  const drag = useRef(null);

  const fitAll = () => {
    setFit(!fit);
  }

  useEffect(() => {

    MapPoint.init();

    const current = view.current;
    const ctx = current.getContext('2d', { willReadFrequently: true });

    let canceled = false;
    let frame = 0;

    const img = new Image();
    img.src = map;
    img.onload = () => {
      if (canceled) return;
      current.width = img.width;
      current.height = img.height;
      ctx.drawImage(img, 0, 0, img.width, img.height);
      if (debug) {
        MapPoint.drawDebugPoint(ctx)
      }
      const pointer = new Image();
      pointer.src = point;
      pointer.onload = () => {
        if (canceled) return;
        drawPointer(ctx, pointer);
      }
    }

    /**
     * 現在位置の描画
     * 毎フレーム position を参照するので、画像の読み込み前に届いた位置も反映される
     * @param {*} ctx
     * @param {*} pointer
     */
    const drawPointer = (ctx, pointer) => {

      //前のフレームで鳥を描いた位置と、その下の背景
      let drawn = null;

      let fadeIn = true;
      const odds = 0.008;
      const min = 0.4;
      let alpha = min;
      const fade = () => {

        ctx.globalAlpha = 1;
        if (drawn !== null) {
          ctx.putImageData(drawn.back, drawn.x, drawn.y);
          drawn = null;
        }

        const p = position.current;
        if (p !== null) {
          drawn = { x: p.x - drawHalf, y: p.y - drawHalf, back: ctx.getImageData(p.x - drawHalf, p.y - drawHalf, drawHalf * 2, drawHalf * 2) };

          //p をアイコンの中心にする
          ctx.globalAlpha = alpha;
          ctx.drawImage(pointer, p.x - iconSize / 2, p.y - iconSize / 2, iconSize, iconSize);
          ctx.globalAlpha = 1;
          if (p.outside) {
            drawArrow(ctx, p);
          }
        }

        if (fadeIn) {
          alpha += odds;
          if (alpha >= 1) {
            fadeIn = false;
          }
        } else {
          alpha -= odds;
          if (alpha <= min) {
            fadeIn = true;
          }
        }
        frame = requestAnimationFrame(fade);
      }
      fade();
    }

    /**
     * マップの外にいるとき、本当の位置の方向に三角を描く
     */
    const drawArrow = (ctx, p) => {
      const r = MapPoint.arrowGap;
      ctx.save();
      ctx.translate(p.x, p.y);
      ctx.rotate(p.angle);
      ctx.beginPath();
      ctx.moveTo(r + MapPoint.arrowLength, 0);
      ctx.lineTo(r, -9);
      ctx.lineTo(r, 9);
      ctx.closePath();
      ctx.fillStyle = "#d32f2f";
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 2;
      ctx.fill();
      ctx.stroke();
      ctx.restore();
    }

    fitAll();

    return () => {
      canceled = true;
      cancelAnimationFrame(frame);
    }
  }, []);

  useEffect(() => {
    fitAll();
  }, [winSize.height]);

  useEffect(() => {
    var p = null;
    if (props.coords !== undefined) {
      p = MapPoint.locate(props.coords.latitude, props.coords.longitude);
    }
    position.current = p;
    if (p !== null && centerPending.current) {
      centerPending.current = false;
      centering(p.x);
    }
  }, [props.coords]);

  //現在地ボタン：今の位置ですぐ寄せ、新しい位置が届いたらもう一度寄せる
  useEffect(() => {
    if (props.center === 0) return;
    if (position.current !== null) {
      centering(position.current.x);
    }
    centerPending.current = true;
  }, [props.center]);

  /**
   * 表示切り替えイベント
   */
  useEffect(() => {
    const current = view.current;
    if (current === null) return;

    var parent = current.parentElement;
    const scaleY = (parent.clientHeight) / MapPoint.imageHeight;

    setScale(scaleY);

    setMapStyle({
      transform: `scale(${scaleY})`,
      transformOrigin: 'left top',
    });

  }, [fit]);

  /**
   * マップクリック処理
   * @param {*} e 
   */
  const handleClick = (e) => {
    //ドラッグで動かした後のクリックは店舗を開かない
    var d = drag.current;
    drag.current = null;
    if (d !== null && d.moved) {
      return;
    }

    var dx = document.querySelector("#layer").scrollLeft;
    var x = (dx + e.pageX) / scale;
    var y = e.pageY / scale;
    props.onShop(Math.trunc(x), Math.trunc(y));
  }

  /**
   * 画像上の x が画面の真ん中に来るように横に動かす
   */
  const centering = (x) => {
    const elm = document.querySelector("#layer");
    elm.scrollTo({ left: (x * scale) - (elm.clientWidth / 2), behavior: "smooth" });
  }

  /**
   * マウスのドラッグで横に動かす（スクロールバーを出していないため）
   * タッチはブラウザのスクロールに任せる
   */
  const handlePointerDown = (e) => {
    if (e.pointerType !== "mouse" || e.button !== 0) return;
    drag.current = { x: e.clientX, left: e.currentTarget.scrollLeft, moved: false, active: true };
    e.currentTarget.setPointerCapture(e.pointerId);
  }

  const handlePointerMove = (e) => {
    var d = drag.current;
    if (d === null || !d.active) return;
    var dx = e.clientX - d.x;
    if (Math.abs(dx) > dragThreshold) {
      d.moved = true;
      e.currentTarget.classList.add("dragging");
    }
    if (d.moved) {
      e.currentTarget.scrollLeft = d.left - dx;
    }
  }

  const handlePointerUp = (e) => {
    var d = drag.current;
    if (d === null) return;
    d.active = false;
    e.currentTarget.classList.remove("dragging");
  }

  //マウスホイールの縦の回転で横に動かす
  const handleWheel = (e) => {
    if (Math.abs(e.deltaY) > Math.abs(e.deltaX)) {
      e.currentTarget.scrollLeft += e.deltaY;
    }
  }

  return (
    <div id="layer" onClick={handleClick}
      onPointerDown={handlePointerDown} onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp} onPointerCancel={handlePointerUp}
      onWheel={handleWheel}>
      <canvas id="map" ref={view} style={mapStyle} />
      {props.overlay && (
        <svg id="overlay" width={MapPoint.imageWidth} height={MapPoint.imageHeight} style={mapStyle}>
          {props.overlay}
        </svg>
      )}
    </div>
  )
}

export default Map;