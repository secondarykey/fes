import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

export default defineConfig({
  plugins: [react()],
  base: '/manage/',
  // MUI v9 では esm/ ディレクトリが廃止され、パッケージ直下の .mjs が
  // exports 経由で解決されるため、旧来の esm エイリアスは不要。
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
