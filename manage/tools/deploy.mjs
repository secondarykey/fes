// ビルド成果物（dist）を app の embed 用ディレクトリにコピーする
//
// npm run deploy から lint・ビルドの後に呼ばれる（app の go generate からも呼ばれる）。
// GAE への反映（gcloud app deploy）は行わない。
// 旧ファイルが embed に残らないよう、コピー先の中身は入れ替える。
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const dist = path.join(root, 'dist')
const dest = path.resolve(root, '..', 'app', 'handler', 'internal', '_assets', 'manage')

if (!fs.existsSync(path.join(dist, 'index.html'))) {
  throw new Error(`ビルド成果物がありません: ${dist}`)
}

// クローン直後はコピー先が無い（.gitignore 対象のため）
fs.mkdirSync(dest, { recursive: true })
for (const name of fs.readdirSync(dest)) {
  fs.rmSync(path.join(dest, name), { recursive: true, force: true })
}
fs.cpSync(dist, dest, { recursive: true })

console.log(`${dest} にコピーしました`)
