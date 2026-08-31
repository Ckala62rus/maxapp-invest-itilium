import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The prototype is intentionally lightweight, so Vite only needs Vue support
// and an open host to work inside Docker and tunnels used by MAX Mini App.
export default defineConfig(() => {
  // В Docker Compose backend доступен как `backend-dev`; на хосте — `127.0.0.1:3000`.
  const apiProxyTarget = process.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:3000'
  const hmrClientPort = Number(process.env.VITE_HMR_CLIENT_PORT || 0)
  const hmrHost = process.env.VITE_HMR_HOST || '127.0.0.1'
  const disableHmr =
    process.env.VITE_DISABLE_HMR === 'true' || process.env.VITE_DISABLE_HMR === '1'

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      strictPort: true,
      // В Docker за nginx HMR часто рвёт WebSocket → @vite/client делает location.reload().
      // Для стабильного демо отключаем HMR (VITE_DISABLE_HMR); обновление — Ctrl+F5.
      hmr: disableHmr
        ? false
        : hmrClientPort > 0
          ? {
              host: hmrHost,
              clientPort: hmrClientPort,
              protocol: 'ws'
            }
          : true,
      ...(disableHmr
        ? {
            watch: {
              ignored: ['**/*']
            }
          }
        : {}),
      // Same-origin `/api/*` → локальный backend (см. `frontend/src/api/axios.js` и `documentation/local_development.md`).
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true
        },
        '/healthz': {
          target: apiProxyTarget,
          changeOrigin: true
        },
        '/readyz': {
          target: apiProxyTarget,
          changeOrigin: true
        },
        '/metrics': {
          target: apiProxyTarget,
          changeOrigin: true
        }
      },
      // Tuna issues dynamic subdomains, so allow the tunnel suffix in development.
      allowedHosts: ['localhost', '127.0.0.1', '.ru.tuna.am']
    }
  }
})
