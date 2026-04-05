import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/healthz': 'http://localhost:8080',
      '/videos': 'http://localhost:8080',
      '/tags': 'http://localhost:8080',
      '/sync': 'http://localhost:8080',
      '/directories': 'http://localhost:8080',
      '/jav': 'http://localhost:8080',
      '/config': 'http://localhost:8080',
    },
  },
})
