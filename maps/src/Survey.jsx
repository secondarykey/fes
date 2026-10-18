import { useEffect, useState } from 'react';
import MapPoint from './MapPoint';
import { errorPx } from './useSurvey';

/**
 * 現地調査の操作パネル（?debug のときだけ出す）
 * @param {*} props survey: useSurvey の戻り値
 */
export const SurveyPanel = ({ survey }) => {

  const [list, setList] = useState(false);
  //「何秒前」を更新するため
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const last = survey.last;
  const p = last === null ? null : MapPoint.get(last.lat, last.lon);
  const samples = survey.samples;

  return (<>
    <div id="survey">
      {last === null ? <div>位置なし</div> : (<>
        <div>{last.lat.toFixed(7)}, {last.lon.toFixed(7)}</div>
        <div>
          ±{Math.round(last.acc)}m · {Math.max(0, Math.round((now - last.t) / 1000))}秒前 · {p === null ? "画像外" : `x ${p.x} y ${p.y}`}
        </div>
      </>)}
      {samples !== null ? (
        <div className="row">
          <span className="rec" title="今いる場所を地図でタップ">
            記録中 {samples.length}件{samples.length > 0 && ` ±${Math.round(Math.min(...samples.map((s) => s.acc)))}m`}
          </span>
          <button onClick={survey.cancel}>やめる</button>
        </div>
      ) : (
        <div className="row">
          <button onClick={survey.start} title="押してから、今いる場所を地図でタップ">記録</button>
          <button className={survey.tracking ? "on" : ""} onClick={survey.toggleTracking} title="歩いた道のりを記録">
            軌跡 {survey.track.length}
          </button>
          <button onClick={() => setList(true)}>一覧 {survey.points.length}</button>
        </div>
      )}
      {survey.saveError && <div className="error">端末に保存できません</div>}
    </div>
    {list && <SurveyList survey={survey} onClose={() => setList(false)} />}
  </>)
}

const SurveyList = ({ survey, onClose }) => {

  //ずれを m でも出す（計算した位置での縮尺）
  const errorText = (pt) => {
    const px = errorPx(pt);
    if (px === null) return "画像外";
    return `${px}px (${(px / MapPoint.pxPerMeter(pt.lat, pt.lon)).toFixed(1)}m)`;
  }

  return (
    <div id="surveyList" onClick={onClose}>
      <div className="card" onClick={(e) => e.stopPropagation()}>
        <table>
          <tbody>
            {survey.points.map((pt) => (
              <tr key={pt.no}>
                <td>{pt.no}</td>
                <td>{pt.name}</td>
                <td>±{Math.round(pt.acc)}m</td>
                <td>{errorText(pt)}</td>
                <td><button aria-label="消す" onClick={() => survey.remove(pt.no)}>✕</button></td>
              </tr>
            ))}
          </tbody>
        </table>
        <div className="row">
          <button onClick={survey.copy} title="地点を TSV でコピー">コピー</button>
          <button onClick={survey.save} title="地点と軌跡を JSON で保存・共有">保存</button>
          <button onClick={survey.clear}>全部消す</button>
          <button onClick={onClose}>閉じる</button>
        </div>
      </div>
    </div>
  )
}

/**
 * マップに重ねる表示（SVG の中身。座標はマップ画像の px）
 * 基準点（黒い点）、現在地の精度の円、記録した地点（緑: 実際の位置、赤: 計算した位置）、軌跡
 * @param {*} props survey: useSurvey の戻り値
 */
export const SurveyOverlay = ({ survey }) => {

  const line = (arr) => arr.map((p) => `${p.x},${p.y}`).join(" ");

  const last = survey.last;
  const p = last === null ? null : MapPoint.get(last.lat, last.lon);
  const k = p === null ? null : MapPoint.pxPerMeter(last.lat, last.lon);

  return (<>
    {MapPoint.controls.map((b) => (
      <circle key={b.source} className="basePoint" cx={b.x} cy={b.y} r="4">
        <title>{b.name}</title>
      </circle>
    ))}

    <polyline className="track" points={line(survey.track.filter((t) => t.x !== null))} />

    {p !== null && k !== null && <circle className="accuracy" cx={p.x} cy={p.y} r={last.acc * k} />}

    {survey.points.map((pt) => (
      <g key={pt.no}>
        {pt.calc && <line className="error" x1={pt.x} y1={pt.y} x2={pt.calc.x} y2={pt.calc.y} />}
        {pt.calc && <circle className="calc" cx={pt.calc.x} cy={pt.calc.y} r="5" />}
        <circle className="truth" cx={pt.x} cy={pt.y} r="7" />
        <text x={pt.x + 9} y={pt.y - 9}>{pt.no}</text>
      </g>
    ))}
  </>)
}
