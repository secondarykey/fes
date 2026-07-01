import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

export default defineConfig({
  plugins: [react()],
  base: '/manage/',
  resolve: {
    alias: {
      '@mui/icons-material': '@mui/icons-material/esm',
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
