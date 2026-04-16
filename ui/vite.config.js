import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const capacitor = env.VITE_CAPACITOR === '1'
  return {
    plugins: [vue()],
    // Capacitor loads the bundle from file/https app origin; relative URLs are required.
    base: capacitor ? './' : '/',
  }
})
