import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

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
    proxy: {
      '/manage/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/login': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/logout': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/session': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/manage/page/view': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/page': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/file': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
