import { useEffect, useRef, useState } from 'react'

import Map from './Map'

import './App.css'
import SVGButton from './SVGButton';
import MapPoint from './MapPoint';


function App() {

  const shop = useRef(undefined);
  const [release, setRelease] = useState(true);
  const [coords, setCoords] = useState(undefined);
  const [info, setInfo] = useState({});

  const handleSuccess = (position) => {
    setCoords(position.coords)
  }

  const handleError = (error) => {
    console.error(`Error: ${error.message}`);
  }

  useEffect(()=> {
    navigator.geolocation.getCurrentPosition(handleSuccess, handleError);
  },[])

  const handleHideMessage = () => {
    setRelease(false);
  }

  const Experimental = () => {

    if ( !release ) {
      return <></>;
    }

    return (
      <div id="experimental">
        <div id="title">Humming Bird 2025 春マップ（体験版）</div>
        <div id="message">
まずはこのページを見つけてくれてありがとうございます

このマップはSmall-Axeプロジェクトが開催しているHummingBird 2025 春のマップ体験版で
次回の開催から導入されるかもしれないマップアプリのテスト版です

不具合や利用したことによる問題は当方では責任をおいかねますことをご了承ください。
特に店舗情報は正しくないかもしれません。
何か機能のご要望があれば、お問い合わせかバーカウンターのスタッフにお伝え下さい。

また「現在位置の取得」をオンにすると、自分のいる場所がわかるようになっています。
※アイコンに使われている「はっち」は Humming Bird 公認の非公式キャラです。
        </div>
        <button onClick={handleHideMessage} >閉じる</button>
      </div>
    )
  }
  const handlePosition = () => {
    navigator.geolocation.getCurrentPosition(handleSuccess, handleError);
  };

const convertLinksWithText = (text) => {
    return text.replace(/(https?:\/\/[^\s]+)/g, '<a href="$1" target="_blank">Link</a>');
}

  const handleShop = (x,y) => {
    if ( shop.current === null ) return;
    var hide = false;

    if ( shop.current.classList.contains("show") ) {
      hide = true;
    } 

    var s = MapPoint.getShop(x,y);
    if ( s !== null ) {
      s.detail = convertLinksWithText(s.detail);
      setInfo(s);
      shop.current.classList.add("show");
      hide = false;
    } 

    if ( hide ) {
      shop.current.classList.remove("show");
    }
  }

  const handleImageError = (e) => {
     e.target.src = "/maps/images/noimage.png"
  }

  const handleBack = () => {
      shop.current.classList.remove("show");
  }

  return (
    <>
      <Experimental />

        <Map coords={coords} onShop={handleShop}/>

      <SVGButton id="positionBtn" name="cross" size="48px" onClick={handlePosition}/>

      <div id="shops" ref={shop}>

        <SVGButton className="icon" id="backBtn" name="back" size="48px" onClick={handleBack}/>

        <img className="thumb" src={info.image} onError={handleImageError}/>
        <div className="info">
          <h1>{info.key}:{info.name}</h1>
          <p style={{whiteSpace:"pre-wrap"}} dangerouslySetInnerHTML={{ __html: info.detail}} />
        </div>
      </div>
    </>
  )
}

export default App
