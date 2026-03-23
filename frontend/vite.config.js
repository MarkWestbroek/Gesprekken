import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

/**
 * Vite configuratie voor de Gesprekken frontend.
 *
 * De dev-server draait op poort 5173 en proxied alle /v1/* requests
 * naar de Go API op poort 8080, zodat er geen CORS-issues zijn
 * tijdens development.
 */
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // Alle API-aanroepen doorsturen naar de Go backend
      '/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
