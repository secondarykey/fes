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
})
