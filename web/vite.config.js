import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
// Use 127.0.0.1 (not localhost) so Windows does not prefer IPv6 [::1] when the Go server listens on IPv4 only.
const webRoot = fileURLToPath(new URL('.', import.meta.url))

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, webRoot, '')
  const apiTarget =
    env.DEV_API_ORIGIN?.trim() ||
    env.VITE_DEV_API_ORIGIN?.trim() ||
    'http://127.0.0.1:17654'

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        '/healthz': apiTarget,
        '/videos': apiTarget,
        '/tags': apiTarget,
        '/sync': apiTarget,
        '/directories': apiTarget,
        '/jav': apiTarget,
        '/config': apiTarget,
        '/collections': apiTarget,
      },
    },
  }
})
