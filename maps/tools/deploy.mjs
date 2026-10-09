// ビルド成果物（dist）を app の embed 用ディレクトリにコピーする
//
// npm run deploy から lint・ビルドの後に呼ばれる。GAE への反映（gcloud app deploy）は行わない。
// app/handler/internal/_assets/maps は go:embed の対象で、空だとコンパイルエラーになるため、
// コピーできることを確かめてから中身を入れ替える。
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const dist = path.join(root, 'dist')
const dest = path.resolve(root, '..', 'app', 'handler', 'internal', '_assets', 'maps')

if (!fs.existsSync(path.join(dist, 'index.html'))) {
  throw new Error(`ビルド成果物がありません: ${dist}`)
}
if (!fs.existsSync(dest)) {
  throw new Error(`コピー先がありません: ${dest}`)
}

// 店舗画像は Git に入れていないので、無いまま出していないか知らせる
const images = path.join(dist, 'images')
const shopDirs = fs.existsSync(images) ? fs.readdirSync(images).filter((d) => /^\d+$/.test(d)) : []
if (shopDirs.length === 0) {
  console.warn('警告: 店舗画像（public/images/{店舗番号}/）がありません。全店 noimage.png になります。')
}

for (const name of fs.readdirSync(dest)) {
  fs.rmSync(path.join(dest, name), { recursive: true, force: true })
}
fs.cpSync(dist, dest, { recursive: true })

console.log(`${dest} にコピーしました（店舗画像 ${shopDirs.length} 店）`)
