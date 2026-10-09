import { useEffect, useRef, useState } from 'react'

import Map from './Map'

import useWindowSize from './Window';
import './App.css'
import SVGButton from './SVGButton';
import MapPoint from './MapPoint';

//会場が狭いので GPS を優先させる（ネットワーク測位だと誤差が会場の幅を超える）
const geoOptions = { enableHighAccuracy: true };

function App() {

  const shop = useRef(undefined);
  const winSize = useWindowSize();

  const [release, setRelease] = useState(true);
  const [coords, setCoords] = useState(undefined);
  const [info, setInfo] = useState({});

  const [debugDisplay,setDebugDisplay] = useState(false);
  const [debugText,setDebugText] = useState("");
  const [lock,setLock] = useState(false);

  const [shopStyle, setShopStyle] = useState({left:"100vw",width:"80%",transition:"0.5s"});

  const addDebugText = (line) => {
    setDebugText((text) => text + "\n" + line);
  }

  const createShopStyle = (show) => {

    var style = {};
    var width = 80;

    if ( show === undefined ) {
      show = (shopStyle.left !== "100vw");
    }

    if ( show === true ) {
      var odds = winSize.width / winSize.height;
      if ( odds > 1 ) {
        width = 40;
      }
      style.left = (100 - width) + "%";
    } else {
      style.left = "100vw"
    }
    style.width = width + "%";
    style.transition = "0.5s"
    setShopStyle(style);
  }

  const handleSuccess = (position) => {
    var coords = position.coords;
    addDebugText("handleSuccess:"+coords.latitude + " - " + coords.longitude)
    setCoords(coords)
    setLock(false);
  }

  const handleError = (error) => {
    addDebugText("handleError:"+error.message)
    console.error(`Error: ${error.message}`);
    setLock(false);
  }

  useEffect(() => {
    addDebugText("watch:");
    const id = navigator.geolocation.watchPosition(handleSuccess, handleError, geoOptions);
    setLock(true);
    return () => {
      navigator.geolocation.clearWatch(id);
    }
  }, [])

  useEffect(() => {
    createShopStyle();
  },[winSize.width,winSize.height]);


  const handleHideMessage = () => {
    setRelease(false);
  }

  const handleHideDebug = () => {
    setDebugDisplay(false);
  }

  const Experimental = () => {

    if (!release) {
      return <></>;
    }

    return (
      <div id="experimental">
        <div id="title">Humming Bird 2026 秋マップ（体験版）</div>
        <div id="message">
          まずはこのページを見つけてくれてありがとうございます

          このマップはSmall-Axeプロジェクトが開催しているHummingBird 2026 秋のマップ体験版で
          次回の開催から導入されるかもしれないマップアプリのテスト版です

          不具合や利用したことによる問題は当方では責任をおいかねますことをご了承ください。
          特に店舗情報は正しくないかもしれません。
          何か機能のご要望があれば、お問い合わせなどでお願いします。

        </div>
        <button onClick={handleHideMessage} >閉じる</button>
      </div>
    )
  }

  const Debug = () => {
    if ( !debugDisplay ) return (<></>);
    return (<>
      <div id="debugText">
        <span>Debug</span>
        <div>{debugText}</div>
        <button onClick={handleHideDebug} >閉じる</button>
      </div>
    </>)

  }


  const handlePosition = () => {
    if ( !lock ) {
      addDebugText("currentPosition:");
      navigator.geolocation.getCurrentPosition(handleSuccess, handleError, geoOptions);
      setLock(true);
    }
  };

  /**
   * リンク変換
   * @param {*} text 
   * @returns 
   */
  const convertLinksWithText = (text) => {
    return text.replace(/(https?:\/\/[^\s]+)/g, '<a href="$1" target="_blank">Link</a>');
  }

  /**
   * 店舗表示
   * @param {*} x 
   * @param {*} y 
   * @returns 
   */
  const handleShop = (x, y) => {
    if (shop.current === null) return;

    var hide = false;
    //すでに表示状態かを取得
    if ( shopStyle.left !== "100vw" ) {
      hide = true;
    }

    //座標による店舗を取得
    var s = MapPoint.getShop(x, y);
    if ( s === null ) {
      var bird = MapPoint.getBird(coords)
      if ( bird !== null ) {
        if ( bird.inside(x,y) ) {
          s = bird;
        }
      }
    }

    if (s !== null) {
      s.html = convertLinksWithText(s.detail);
      setInfo(s);

      createShopStyle(true)
      hide = false;
    }

    //店舗がない場合、非表示
    if ( hide ) {
      createShopStyle(false)
    }
  }

  const handleImageError = (e) => {
    e.target.src = "/maps/images/noimage.png"
  }

  //縦長の写真は表示の仕方を変える
  const handleImageLoad = (e) => {
    var img = e.currentTarget;
    img.parentElement.classList.toggle("tall", img.naturalHeight > img.naturalWidth);
  }

  const handleBack = () => {
    createShopStyle(false)
  }

  const handleReload = () => {
    location.reload();
  }

  // eslint-disable-next-line no-unused-vars -- デバッグボタン (コメントアウト中) 用に残す
  const handleDebug = () => {
    setDebugDisplay(true);
  }

  return (
    <>
      <Experimental />
      <Debug />

      <Map coords={coords} onShop={handleShop} />
{/**
      <SVGButton id="debugBtn" name="terminal" size="16px" onClick={handleDebug} />
 */}
      <SVGButton id="updateBtn" name="update" size="16px" onClick={handleReload} />

      <SVGButton id="positionBtn" name="cross" size="32px" onClick={handlePosition} />

      <div id="shops" ref={shop} style={shopStyle}>

        <SVGButton className="icon" id="backBtn" name="back" size="48px" onClick={handleBack} />

        {/* 店舗を切り替えたら先頭の画像に戻すため key を付ける */}
        <div className="thumbs" key={info.key}>
          {(info.images === undefined || info.images.length === 0 ? ["/maps/images/noimage.png"] : info.images).map((src) => (
            <div className="thumb" key={src}>
              <img className="backdrop" src={src} alt="" aria-hidden="true" onError={handleImageError} />
              <img className="photo" src={src} onLoad={handleImageLoad} onError={handleImageError} />
            </div>
          ))}
        </div>
        <div className="info">
          <h1>{info.key}:{info.name}</h1>
          <p style={{ whiteSpace: "pre-wrap" }} dangerouslySetInnerHTML={{ __html: info.html }} />
        </div>

        <SVGButton className="icon" id="backBBtn" name="back" size="48px" onClick={handleBack} />
      </div>
    </>
  )
}

export default App
