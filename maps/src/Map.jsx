import { useEffect, useRef, useState } from 'react';
import map from './assets/map.jpg'
import point from './assets/point.png'
import useWindowSize from './Window';
// eslint-disable-next-line no-unused-vars -- Point はデバッグ描画で使うため残す
import MapPoint, { Point } from './MapPoint';

const iconSize = MapPoint.pointerSize;

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

  //現在地のマップ上の点（未取得・マップ外は null）
  const position = useRef(null);

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
          //p をアイコンの中心にする
          const x = p.x - iconSize / 2;
          const y = p.y - iconSize / 2;
          drawn = { x: x, y: y, back: ctx.getImageData(x, y, iconSize, iconSize) };

          ctx.globalAlpha = alpha;
          ctx.drawImage(pointer, x, y, iconSize, iconSize);
          ctx.globalAlpha = 1;
          //centering(document.querySelector("#layer"),p.x);
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
      var ctx = debug ? view.current.getContext('2d', { willReadFrequently: true }) : undefined;
      p = MapPoint.get(props.coords.latitude, props.coords.longitude, ctx);
    }
    position.current = p;
  }, [props.coords]);

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

  // eslint-disable-next-line no-unused-vars -- 現在地センタリングは未有効化のため残す
  const centering = (elm,x) => {
    
    console.log("center",x);
    console.log("scale",scale);

    elm.scrollLeft = (x * scale) - (elm.clientWidth / 2);
    //layer.scrollLeft = p.x + ( (layer.clientWidth) * scale / 2) + 100
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
    </div>
  )
}

export default Map;