import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// npm run dev のとき、API などは Go のサーバーへ中継する。
const server = 'http://localhost:8080'

// Go 側は /manage/ 配下の GET 以外を、Origin が Host と同じときだけ受け付ける（CSRF 対策）。
// changeOrigin で Host は server に変わるが、ブラウザが付けた Origin は
// http://localhost:5173 のまま届いて 403 になるため、Origin も server に合わせる。
const toServer = {
  target: server,
  changeOrigin: true,
  configure(proxy) {
    proxy.on('proxyReq', (req) => {
      if (req.getHeader('origin')) req.setHeader('origin', server)
    })
  },
}

export default defineConfig({
  plugins: [react()],
  base: '/manage/',
  // MUI v9 では esm/ ディレクトリが廃止され、パッケージ直下の .mjs が
  // exports 経由で解決されるため、旧来の esm エイリアスは不要。
  build: {
    rollupOptions: {
      output: {
        // ライブラリごとにチャンクを分けて、アプリ更新時に vendor 側の
        // ブラウザキャッシュが無効化されないようにする。
        // @emotion は MUI と初期化順序の依存があるため同一チャンクに置く。
        manualChunks(id) {
          if (!id.includes('node_modules')) return
          if (id.includes('@mui') || id.includes('@emotion')) return 'mui'
          if (id.includes('@dnd-kit')) return 'dnd-kit'
          return 'vendor'
        },
      },
    },
  },
  server: {
    // /manage/ 配下は SPA 自身（vite が返す）なので、Go が返すパスだけを並べる。
    // 先頭一致なので '/file' は /file_cache も含む。
    proxy: Object.fromEntries(
      [
        '/manage/api',
        '/manage/page/view', // ページのプレビュー
        '/manage/file/view', // 下書きを含む画像（ページ画像など）
        '/manage/debug/pprof',
        '/login',
        '/logout',
        '/session',
        '/page',
        '/file',
      ].map((path) => [path, toServer]),
    ),
  },
})
