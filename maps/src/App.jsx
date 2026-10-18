import { useEffect, useRef, useState } from 'react'

import Map from './Map'

import useWindowSize from './Window';
import './App.css'
import SVGButton from './SVGButton';
import MapPoint from './MapPoint';
import { surveyEnabled, useSurvey } from './useSurvey';
import { SurveyOverlay, SurveyPanel } from './Survey';
import LocationHelp from './LocationHelp';

//店舗パネルに出す区分の名前（マップ画像の店舗一覧の見出しに合わせる）
const groupLabels = {
  food: "フード・ドリンク・スイーツ",
  kitchencar: "キッチンカー",
  goods: "雑貨・ワークショップ・リラクゼーション",
  facility: "会場",
};

//番号のバッジは出店者とステージ（A〜D）だけに出す
const showShopNo = (s) => s.key !== undefined && (s.group !== "facility" || /^[A-D]$/.test(s.key)) && s.key !== "BIRD";

//会場が狭いので GPS を優先させる（ネットワーク測位だと誤差が会場の幅を超える）
const geoOptions = { enableHighAccuracy: true };

//全画面にできるブラウザだけ全画面ボタンを出す（iPhone の Safari は動画以外の全画面を許していない）
const fullscreenEnabled = document.fullscreenEnabled === true;

/**
 * 最初に出すお知らせ
 * App の中で定義すると、位置が届いて App を描き直すたびに作り直され、フェードインがやり直される
 */
const Notice = ({ onClose }) => {
  //外側（暗い部分）を押しても閉じる
  return (
    <div id="experimental" onClick={onClose}>
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
          <p>Small-Axe プロジェクトが開催する Humming Bird 2026 秋の会場マップです。</p>
          <div className="note">
            <p>
              お店の情報や配置が実際と違うことがあります。
              最新の情報は公式サイトでご確認ください。
            </p>
            <p>何か機能のご要望があれば、お問い合わせなどでお願いします。</p>
          </div>
        </div>
        <button className="close" onClick={onClose}>マップを見る</button>
        {/* マップは /maps/ で配信しているので、/ がサイトのトップになる */}
        <a className="site" href="/">サイトに戻る</a>
      </div>
    </div>
  )
}

//デバッグ表示（ボタンはコメントアウト中）
const Debug = ({ text, onClose }) => {
  return (
    <div id="debugText">
      <span>Debug</span>
      <div>{text}</div>
      <button onClick={onClose} >閉じる</button>
    </div>
  )
}

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

  //位置情報が拒否されている（ブラウザが覚えるので、設定を直すまでダイアログは出ない）
  const [denied, setDenied] = useState(false);
  const deniedRef = useRef(false);
  //拒否されているときの戻し方のカード
  const [help, setHelp] = useState(false);
  //現在地ボタンを押した回数（押すたびにハッちを画面の真ん中に寄せる）
  const [center, setCenter] = useState(0);
  //全画面表示中か（ブラウザの戻る操作などで抜けることもあるので、イベントで受け取る）
  const [fullscreen, setFullscreen] = useState(false);
  const watch = useRef(null);

  const [shopStyle, setShopStyle] = useState({left:"100vw",width:"80%",transition:"0.5s"});

  //現地調査（?debug のときだけ）
  const survey = useSurvey(coords);

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

  const setDeniedState = (v) => {
    deniedRef.current = v;
    setDenied(v);
  }

  const handleSuccess = (position) => {
    var coords = position.coords;
    addDebugText("handleSuccess:"+coords.latitude + " - " + coords.longitude)
    setCoords(coords)
    setLock(false);
    //拒否から許可に戻った。拒否で止まった追いかけをやり直す
    if ( deniedRef.current ) {
      setDeniedState(false);
      setHelp(false);
      startWatch();
    }
  }

  const handleError = (error) => {
    addDebugText("handleError:"+error.message)
    console.error(`Error: ${error.message}`);
    setLock(false);
    if ( error.code === error.PERMISSION_DENIED ) {
      setDeniedState(true);
    }
  }

  const startWatch = () => {
    if ( watch.current !== null ) {
      navigator.geolocation.clearWatch(watch.current);
    }
    addDebugText("watch:");
    watch.current = navigator.geolocation.watchPosition(handleSuccess, handleError, geoOptions);
  }

  //拒否されたままかを確かめる（許可に戻っていれば handleSuccess で追いかけ直す）
  const recheck = (onDenied) => {
    navigator.geolocation.getCurrentPosition(handleSuccess, (e) => {
      handleError(e);
      if ( e.code === e.PERMISSION_DENIED && onDenied ) onDenied();
    }, geoOptions);
  }

  useEffect(() => {
    startWatch();
    setLock(true);

    //許可の状態が変わったら追いかけ直す（変化を通知しないブラウザもあるので、画面に戻ったときも確かめる）
    let status = null;
    navigator.permissions?.query({ name: "geolocation" }).then((s) => {
      status = s;
      if ( s.state === "denied" ) setDeniedState(true);
      s.onchange = () => {
        if ( s.state === "denied" ) {
          setDeniedState(true);
        } else if ( deniedRef.current ) {
          startWatch();
        }
      };
    }).catch(() => {});

    const handleVisible = () => {
      if ( document.visibilityState === "visible" && deniedRef.current ) recheck();
    }
    document.addEventListener("visibilitychange", handleVisible);

    return () => {
      if ( status !== null ) status.onchange = null;
      document.removeEventListener("visibilitychange", handleVisible);
      navigator.geolocation.clearWatch(watch.current);
      watch.current = null;
    }
  }, [])

  useEffect(() => {
    createShopStyle();
  },[winSize.width,winSize.height]);

  useEffect(() => {
    const handleChange = () => setFullscreen(document.fullscreenElement !== null);
    document.addEventListener("fullscreenchange", handleChange);
    return () => document.removeEventListener("fullscreenchange", handleChange);
  }, []);

  //ブラウザのバーを消してマップを画面いっぱいにする（タップからしか全画面にできない）
  const handleFullscreen = () => {
    if ( document.fullscreenElement ) {
      document.exitFullscreen().catch(() => {});
    } else {
      document.documentElement.requestFullscreen({ navigationUI: "hide" }).catch((e) => console.error(e));
    }
  }


  const handleHideMessage = () => {
    setRelease(false);
  }

  const handleHideDebug = () => {
    setDebugDisplay(false);
  }

  const handlePosition = () => {
    if ( denied ) {
      recheck(() => setHelp(true));
      return;
    }
    setCenter((c) => c + 1);
    if ( !lock ) {
      addDebugText("currentPosition:");
      navigator.geolocation.getCurrentPosition(handleSuccess, handleError, geoOptions);
      setLock(true);
    }
  };

  /**
   * リンクの表示名（Instagram はサービス名、それ以外はドメイン）
   * @param {*} url
   * @returns
   */
  const linkLabel = (url) => {
    try {
      var host = new URL(url).hostname.replace(/^www\./, "");
      return host.endsWith("instagram.com") ? "Instagram" : host;
    } catch {
      return url;
    }
  }

  /**
   * 店舗表示
   * @param {*} x 
   * @param {*} y 
   * @returns 
   */
  const handleShop = (x, y) => {
    if (shop.current === null) return;

    //記録中のタップは今いる場所として使う
    if ( survey.tap(x, y) ) return;

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
      {release && <Notice onClose={handleHideMessage} />}
      {debugDisplay && <Debug text={debugText} onClose={handleHideDebug} />}

      <Map coords={coords} center={center} onShop={handleShop}
        overlay={surveyEnabled ? <SurveyOverlay survey={survey} /> : undefined} />
      {surveyEnabled && <SurveyPanel survey={survey} />}
{/**
      <SVGButton id="debugBtn" name="terminal" size="16px" onClick={handleDebug} />
 */}
      <SVGButton id="updateBtn" name="update" size="16px" onClick={handleReload} />
      {fullscreenEnabled && (
        <SVGButton id="fullscreenBtn" name={fullscreen ? "fullscreen-exit" : "fullscreen"} size="16px"
          onClick={handleFullscreen} title={fullscreen ? "全画面をやめる" : "全画面"} />
      )}

      <SVGButton id="positionBtn" name="cross" size="32px" onClick={handlePosition}
        className={denied ? "denied" : undefined} title={denied ? "位置情報がオフです" : undefined} />
      {help && <LocationHelp onClose={() => setHelp(false)} />}

      <div id="shops" ref={shop} style={shopStyle}>

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
          <button className="panelClose" aria-label="閉じる" onClick={handleBack}>✕</button>
        </div>
        <div className={"info " + (info.group || "")}>
          {groupLabels[info.group] && <div className="group">{groupLabels[info.group]}</div>}
          <div className="title">
            {showShopNo(info) && <span className="no">{info.key}</span>}
            <h1>{info.name}</h1>
          </div>
          {info.detail && <p className="detail">{info.detail}</p>}
          {info.links && info.links.length > 0 && (
            <div className="links">
              {info.links.map((url) => (
                <a key={url} href={url} target="_blank" rel="noopener noreferrer">
                  {linkLabel(url)}
                  <svg viewBox="0 0 24 24" width="12" height="12" aria-hidden="true">
                    <path fill="currentColor" d="M14 3h7v7h-2V6.41l-9.29 9.3-1.42-1.42L17.59 5H14V3zM5 5h6v2H7v10h10v-4h2v6H5V5z" />
                  </svg>
                </a>
              ))}
            </div>
          )}
        </div>

        <button className="back" onClick={handleBack}>マップに戻る</button>
      </div>
    </>
  )
}

export default App
