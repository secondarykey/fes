import { useEffect, useRef, useState } from 'react';
import map from './assets/map.jpg'
import point from './assets/point.png'
import useWindowSize from './Window';
import MapPoint, { Point } from './MapPoint';

const iconSize = 48;

const Map = (props) => {

  const [debug, setDebug] = useState(false);

  const [fit, setFit] = useState(false);
  const [scale, setScale] = useState(1);
  const [mapStyle, setMapStyle] = useState({});

  const winSize = useWindowSize();
  const view = useRef(undefined);

  const fitAll = () => {
    setFit(!fit);
  }

  useEffect(() => {

    MapPoint.init();

    const current = view.current;
    const ctx = current.getContext('2d', { willReadFrequently: true });

    const img = new Image();
    img.src = map;
    img.onload = () => {
      current.width = img.width;
      current.height = img.height;
      ctx.drawImage(img, 0, 0, img.width, img.height);
      const pointer = new Image();
      pointer.src = point;
      pointer.onload = () => {
        drawPointer(img, pointer);
      }
    }

    fitAll();
  }, []);

  useEffect(() => {
    fitAll();
  }, [winSize.height]);


  /**
   * 現在位置の描画
   * @param {*} img 
   * @param {*} pointer 
   */
  const drawPointer = (img, pointer) => {

    var layer = document.querySelector("#layer");

    const current = view.current;
    const ctx = current.getContext('2d', { willReadFrequently: true });

    let fadeIn = true;
    if (debug) {
      MapPoint.drawDebugPoint(ctx)
    }

    var p = null;
    if (props.coords !== undefined) {
      p = MapPoint.get(props.coords.latitude, props.coords.longitude, debug ? ctx : undefined);
    }

    if (p === null) {
      p = MapPoint.get(34.9248673, 138.3785824);
    } else {
      //TODO 広野公園内でしか動作しない
    }

    var back = ctx.getImageData(p.x, p.y, iconSize, iconSize);
    //centering(layer,p.x);

    current.addEventListener("change.pointer", function (e) {

      ctx.globalAlpha = 1;
      ctx.drawImage(img, 0, 0);

      var po = e.coords;
      var wk = MapPoint.get(po.latitude, po.longitude, debug ? ctx : undefined);
      if (wk !== null) {
        p = wk;
        back = ctx.getImageData(p.x, p.y, iconSize, iconSize);
        //centering(layer,p.x);
      } else {
        //TODO 広野公園内でしか動作しない
      }
    });

    const odds = 0.008;
    const min = 0.4;
    var alpha = min;
    const fade = () => {

      ctx.globalAlpha = 1;
      ctx.putImageData(back, p.x, p.y);

      ctx.globalAlpha = alpha;
      ctx.drawImage(pointer, p.x, p.y, iconSize, iconSize);
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
      if (true) {
        requestAnimationFrame(fade);
      }
    }
    fade();
  }

  useEffect(() => {

    //変更イベントを発行
    const current = view.current;
    if (current !== null) {
      var e = new Event("change.pointer");
      e.coords = props.coords;
      current.dispatchEvent(e)
    }

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
    var dx = document.querySelector("#layer").scrollLeft;
    var x = (dx + e.pageX) / scale;
    var y = e.pageY / scale;
    props.onShop(Math.trunc(x), Math.trunc(y));
  }

  const centering = (elm,x) => {
    
    console.log("center",x);
    console.log("scale",scale);

    elm.scrollLeft = (x * scale) - (elm.clientWidth / 2);
    //layer.scrollLeft = p.x + ( (layer.clientWidth) * scale / 2) + 100
  }

  return (
    <div id="layer" onClick={handleClick}>
      <canvas id="map" ref={view} style={mapStyle} />
    </div>
  )
}

export default Map;