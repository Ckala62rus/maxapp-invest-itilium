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
      // To call the backend directly from the browser, keep Vite proxy disabled.
      // Previous proxy target example: VITE_BACKEND_API or http://localhost:3000
      // proxy: {
      //   '/api': {
      //     target: 'http://localhost:3000',
      //     changeOrigin: true
      //   }
      // }
    }
  }
})
