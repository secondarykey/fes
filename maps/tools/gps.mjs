// 現地調査（/maps/?debug の「保存」）の JSON から、GPS の基準点 src/data/gps.json を作る
//
//   node tools/gps.mjs data/2026-Fall/survey-20261018.json [他の調査…] [--smoothing 100] [--exclude 3,7]
//
// 地点ごとの「タップした画像上の位置（x, y）」と「平均した緯度・経度」を基準点にする。
// 複数の調査を渡すと、まとめて 1 つの基準点にする。--exclude は「ファイル番号:地点番号」か
// 「地点番号」（1 つ目のファイル）で、使わない地点を指定する。
// smoothing は変換（MapPoint の薄板スプライン）のならしの強さ。0 だと全地点をぴったり通る。
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const out = path.join(root, 'src', 'data', 'gps.json')

const files = []
let smoothing = 100
const exclude = new Set()
const args = process.argv.slice(2)
for (let i = 0; i < args.length; i++) {
  if (args[i] === '--smoothing') {
    smoothing = Number(args[++i])
  } else if (args[i] === '--exclude') {
    args[++i].split(',').forEach((v) => exclude.add(v.includes(':') ? v : `1:${v}`))
  } else {
    files.push(args[i])
  }
}
if (files.length === 0 || Number.isNaN(smoothing)) {
  console.error('使い方: node tools/gps.mjs <調査の JSON>... [--smoothing 100] [--exclude 3,7]')
  process.exit(1)
}

const points = []
files.forEach((file, i) => {
  const survey = JSON.parse(fs.readFileSync(file, 'utf8'))
  survey.points.forEach((p) => {
    if (exclude.has(`${i + 1}:${p.no}`)) return
    points.push({
      //全角スペースは半角にそろえる
      name: p.name.replace(/\s+/g, ' '),
      x: p.x,
      y: p.y,
      lat: Number(p.lat.toFixed(7)),
      lon: Number(p.lon.toFixed(7)),
      source: `${path.basename(file)}#${p.no}`,
    })
  })
})

fs.writeFileSync(out, JSON.stringify({ smoothing, points }, null, 2) + '\n')
console.log(`${points.length} 地点を ${path.relative(root, out)} に書き出しました（smoothing ${smoothing}）`)
