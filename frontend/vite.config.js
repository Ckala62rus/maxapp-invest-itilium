import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The prototype is intentionally lightweight, so Vite only needs Vue support
// and an open host to work inside Docker and tunnels used by MAX Mini App.
export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 5173
  }
})
