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
  //表示中の写真の番号
  const [photo, setPhoto] = useState(0);

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

    //外側（暗い部分）を押しても閉じる
    return (
      <div id="experimental" onClick={handleHideMessage}>
        <div className="card" role="dialog" aria-modal="true" aria-labelledby="experimentalTitle"
          onClick={(e) => e.stopPropagation()}>
          <div className="head">
            <img src="/maps/images/BIRD/0.webp" alt="" />
            <div>
              <div className="event">Humming Bird 2026 秋</div>
              <div className="title" id="experimentalTitle">会場マップ<span className="badge">体験版</span></div>
            </div>
          </div>
          <div className="body">
            <p>まずはこのページを見つけてくれてありがとうございます。</p>
            <p>
              このマップは Small-Axe プロジェクトが開催している HummingBird 2026 秋のマップ体験版で、
              次回の開催から導入されるかもしれないマップアプリのテスト版です。
            </p>
            <div className="note">
              <p>
                不具合や利用したことによる問題は、当方では責任を負いかねますことをご了承ください。
                特に店舗情報は正しくないかもしれません。
              </p>
              <p>何か機能のご要望があれば、お問い合わせなどでお願いします。</p>
            </div>
          </div>
          <button className="close" onClick={handleHideMessage}>マップを見る</button>
        </div>
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
      setPhoto(0);

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

  const photos = (info.images === undefined || info.images.length === 0) ? ["/maps/images/noimage.png"] : info.images;
  const photoSrc = photos[photo % photos.length];

  //前後の写真へ（端では反対側に回る）
  const movePhoto = (d) => {
    setPhoto((p) => (p + d + photos.length) % photos.length);
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

        <div className="thumbs">
          {/* 写真ごとに作り直して、縦長かどうかを判定し直す */}
          <div className="thumb" key={photoSrc}>
            <img className="backdrop" src={photoSrc} alt="" aria-hidden="true" onError={handleImageError} />
            <img className="photo" src={photoSrc} onLoad={handleImageLoad} onError={handleImageError} />
          </div>
          {photos.length > 1 && (<>
            <button className="photoBtn prev" aria-label="前の写真" onClick={() => movePhoto(-1)}>‹</button>
            <button className="photoBtn next" aria-label="次の写真" onClick={() => movePhoto(1)}>›</button>
            <div className="dots">
              {photos.map((src, i) => <span key={src} className={i === photo ? "on" : ""} />)}
            </div>
          </>)}
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
