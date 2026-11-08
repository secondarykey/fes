/**
 * 位置情報が拒否されているときの戻し方（現在地ボタンを押したときに出す）
 *
 * 一度拒否すると、ブラウザはそのサイトに許可のダイアログを出し直さない。
 * 利用者がブラウザ・端末の設定を直して、再読み込みする必要がある。
 */

//アプリ内ブラウザ（会場では QR から LINE などで開く人が多い）
const inAppBrowsers = [
  { name: "LINE", re: / Line\//i },
  { name: "Instagram", re: /Instagram/ },
  { name: "Facebook", re: /FBAN|FBAV/ },
];

//iOS の Safari 以外のブラウザ
const iosBrowsers = [
  { name: "Chrome", re: /CriOS/ },
  { name: "Firefox", re: /FxiOS/ },
  { name: "Edge", re: /EdgiOS/ },
];

/**
 * 端末・ブラウザに合わせた手順
 * @returns { steps: 手順, more: それでも使えないとき（無ければ undefined） }
 */
const guide = () => {
  const ua = navigator.userAgent;
  //iPadOS の Safari は Mac と同じ UA になる
  const ios = /iPhone|iPad|iPod/.test(ua) || (/Macintosh/.test(ua) && navigator.maxTouchPoints > 1);
  const android = /Android/.test(ua);

  const app = inAppBrowsers.find((b) => b.re.test(ua));
  if (app !== undefined) {
    return {
      steps: [
        "メニュー（… など）から「ブラウザで開く」を選ぶ",
        "開いたブラウザで、位置情報を「許可」する",
      ],
      more: ios
        ? `このまま使うときは、端末の「設定」→「${app.name}」→「位置情報」を「このAppの使用中」にして、再読み込みする`
        : `このまま使うときは、端末の「設定」→「アプリ」→「${app.name}」→「権限」→「位置情報」を許可して、再読み込みする`,
    };
  }

  if (ios) {
    const other = iosBrowsers.find((b) => b.re.test(ua));
    if (other !== undefined) {
      return {
        steps: [
          `端末の「設定」→「${other.name}」→「位置情報」を「このAppの使用中」にする`,
          "下の「再読み込み」を押す",
        ],
      };
    }
    return {
      steps: [
        "アドレスバーの「ぁあ」（ページメニュー）→「Webサイトの設定」→「位置情報」を「許可」にする",
        "下の「再読み込み」を押す",
      ],
      more: "それでも使えないときは、端末の「設定」→「プライバシーとセキュリティ」→「位置情報サービス」をオンにし、「Safari Webサイト」を「このAppの使用中」にする",
    };
  }

  if (android) {
    return {
      steps: [
        "アドレスバー左のアイコン →「権限」→「位置情報」を許可する",
        "下の「再読み込み」を押す",
      ],
      more: "それでも使えないときは、端末の位置情報（クイック設定の「位置情報」）をオンにする",
    };
  }

  return {
    steps: [
      "アドレスバー左のアイコンから、位置情報を許可する",
      "下の「再読み込み」を押す",
    ],
  };
}

const LocationHelp = ({ onClose }) => {

  const { steps, more } = guide();

  return (
    <div id="locationHelp" onClick={onClose}>
      <div className="card" role="dialog" aria-modal="true" aria-labelledby="locationHelpTitle"
        onClick={(e) => e.stopPropagation()}>
        <div className="title" id="locationHelpTitle">現在地を使うには</div>
        <p className="lead">位置情報の利用がオフになっています。</p>
        <ol>
          {steps.map((s) => <li key={s}>{s}</li>)}
        </ol>
        {more && <p className="more">{more}</p>}
        <button className="reload" onClick={() => location.reload()}>再読み込み</button>
        <button className="close" onClick={onClose}>閉じる</button>
      </div>
    </div>
  )
}

export default LocationHelp;
