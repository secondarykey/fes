import { useEffect, useState } from 'react';
import MapPoint from './MapPoint';

/**
 * 現地調査（URL に ?debug を付けたときだけ使う）
 *
 * その場の GPS と「実際にいる地図上の位置」の組を記録して、基準点（src/data/gps.json。tools/gps.mjs で作る）に使う。
 * 歩いた道のり（軌跡）も記録できる。記録は端末の localStorage に残し、JSON / TSV で書き出す。
 */
export const surveyEnabled = new URLSearchParams(location.search).has("debug");

const storageKey = "maps.survey";

//記録を始めたときに、これより古い位置は使わない（ms）
const freshness = 30 * 1000;

const load = () => {
  try {
    const v = JSON.parse(localStorage.getItem(storageKey) ?? "{}");
    return { points: v.points ?? [], track: v.track ?? [] };
  } catch {
    return { points: [], track: [] };
  }
}

/**
 * 記録した地点の、計算した位置と実際の位置（タップした位置）の差（px）
 */
export const errorPx = (pt) => pt.calc ? Math.round(Math.hypot(pt.calc.x - pt.x, pt.calc.y - pt.y)) : null;

/**
 * 精度で重みを付けた平均（精度は一番よいもの）
 */
const average = (samples) => {
  let w = 0, lat = 0, lon = 0;
  samples.forEach((s) => {
    const k = 1 / Math.max(s.acc, 1) ** 2;
    w += k;
    lat += s.lat * k;
    lon += s.lon * k;
  });
  return { lat: lat / w, lon: lon / w, acc: Math.min(...samples.map((s) => s.acc)) };
}

const toTSV = (points) => {
  const head = ["no", "name", "x", "y", "lat", "lon", "acc", "n", "calc_x", "calc_y", "err_px", "time"];
  const rows = points.map((p) => [p.no, p.name, p.x, p.y, p.lat.toFixed(7), p.lon.toFixed(7), Math.round(p.acc), p.n,
    p.calc?.x ?? "", p.calc?.y ?? "", errorPx(p) ?? "", p.time]);
  return [head, ...rows].map((r) => r.join("\t")).join("\n");
}

const stamp = () => {
  const d = new Date();
  const z = (v) => String(v).padStart(2, "0");
  return `${d.getFullYear()}${z(d.getMonth() + 1)}${z(d.getDate())}-${z(d.getHours())}${z(d.getMinutes())}`;
}

/**
 * 現地調査の状態
 * @param {*} coords App が受け取った最新の位置
 */
export const useSurvey = (coords) => {

  const [data, setData] = useState(load);
  //最新の位置 { lat, lon, acc, t }
  const [last, setLast] = useState(null);
  //記録中に集めた位置（記録していないときは null）
  const [samples, setSamples] = useState(null);
  const [tracking, setTracking] = useState(false);
  const [saveError, setSaveError] = useState(false);

  useEffect(() => {
    if (!surveyEnabled || coords === undefined) return;
    const fix = { lat: coords.latitude, lon: coords.longitude, acc: coords.accuracy, t: Date.now() };
    setLast(fix);
    setSamples((s) => s === null ? null : [...s, fix]);
    if (tracking) {
      //表示用に、その時点の変換での地図上の位置も持つ
      const p = MapPoint.get(fix.lat, fix.lon);
      setData((d) => ({ ...d, track: [...d.track, { ...fix, x: p?.x ?? null, y: p?.y ?? null }] }));
    }
  }, [coords]);

  useEffect(() => {
    if (!surveyEnabled) return;
    try {
      localStorage.setItem(storageKey, JSON.stringify(data));
      setSaveError(false);
    } catch (e) {
      console.error(e);
      setSaveError(true);
    }
  }, [data]);

  //画面が消えると位置が届かなくなるので、調査中は画面を点けたままにする
  useEffect(() => {
    if (!surveyEnabled || !("wakeLock" in navigator)) return;
    let lock = null;
    const request = () => {
      navigator.wakeLock.request("screen").then((l) => { lock = l; }).catch(() => { });
    }
    const handleVisible = () => {
      if (document.visibilityState === "visible") request();
    }
    request();
    document.addEventListener("visibilitychange", handleVisible);
    return () => {
      document.removeEventListener("visibilitychange", handleVisible);
      lock?.release();
    }
  }, []);

  const start = () => {
    setSamples(last !== null && Date.now() - last.t < freshness ? [last] : []);
  }

  const cancel = () => {
    setSamples(null);
  }

  /**
   * 記録中なら、タップした位置を「実際にいる位置」として記録する
   * @returns タップを記録に使ったか（使ったときは店舗を開かない）
   */
  const tap = (x, y) => {
    if (samples === null) return false;
    if (samples.length === 0) {
      alert("位置が取れていません");
      return true;
    }
    const no = data.points.reduce((m, p) => Math.max(m, p.no), 0) + 1;
    const name = prompt("地点の名前", `P${no}`);
    if (name === null) return true;

    const a = average(samples);
    const calc = MapPoint.get(a.lat, a.lon);
    const point = {
      no, name, x, y, ...a, n: samples.length,
      calc: calc === null ? null : { x: calc.x, y: calc.y },
      time: new Date().toISOString(),
      samples,
    };
    setData((d) => ({ ...d, points: [...d.points, point] }));
    setSamples(null);
    return true;
  }

  const remove = (no) => {
    if (!confirm(`${no} を消しますか`)) return;
    setData((d) => ({ ...d, points: d.points.filter((p) => p.no !== no) }));
  }

  const clear = () => {
    if (!confirm("地点と軌跡をすべて消しますか")) return;
    setData({ points: [], track: [] });
  }

  const copy = () => {
    navigator.clipboard.writeText(toTSV(data.points))
      .then(() => alert("コピーしました"))
      .catch((e) => alert("コピーできません: " + e.message));
  }

  //共有できる端末では共有（LINE・メールなどで送る）、できなければダウンロード
  const save = async () => {
    const body = {
      image: { width: MapPoint.imageWidth, height: MapPoint.imageHeight },
      userAgent: navigator.userAgent,
      ...data,
    };
    const file = new File([JSON.stringify(body, null, 1)], `maps-survey-${stamp()}.json`, { type: "application/json" });
    if (navigator.canShare?.({ files: [file] })) {
      try {
        await navigator.share({ files: [file] });
        return;
      } catch (e) {
        if (e.name === "AbortError") return;
      }
    }
    const a = document.createElement("a");
    a.href = URL.createObjectURL(file);
    a.download = file.name;
    a.click();
    setTimeout(() => URL.revokeObjectURL(a.href), 1000);
  }

  return {
    last, samples, tracking, saveError,
    points: data.points, track: data.track,
    toggleTracking: () => setTracking((t) => !t),
    start, cancel, tap, remove, clear, copy, save,
  };
}
