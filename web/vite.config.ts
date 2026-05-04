import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiUrl = env.VITE_API_URL || 'http://localhost:8080/v1'
  const target = apiUrl.replace(/\/v1\/?$/, '')
  const allowedHosts = env.VITE_ALLOWED_HOSTS ? env.VITE_ALLOWED_HOSTS.split(',') : []

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
      },
    },
    server: {
      port: 3023,
      host: '0.0.0.0',
      allowedHosts,
      proxy: {
        '/v1': {
          target,
          changeOrigin: true,
          ws: true,
        },
      },
    },
  }
})
