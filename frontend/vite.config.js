import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The prototype is intentionally lightweight, so Vite only needs Vue support
// and an open host to work inside Docker and tunnels used by MAX Mini App.
export default defineConfig(() => {
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
      // Same-origin `/api/*` → локальный backend (см. `frontend/src/api/axios.js` и `documentation/local_development.md`).
      proxy: {
        '/api': {
          target: 'http://127.0.0.1:3000',
          changeOrigin: true
        },
        '/healthz': {
          target: 'http://127.0.0.1:3000',
          changeOrigin: true
        },
        '/readyz': {
          target: 'http://127.0.0.1:3000',
          changeOrigin: true
        },
        '/metrics': {
          target: 'http://127.0.0.1:3000',
          changeOrigin: true
        }
      },
      // Tuna issues dynamic subdomains, so allow the tunnel suffix in development.
      allowedHosts: ['localhost', '127.0.0.1', '.ru.tuna.am']
    }
  }
})
