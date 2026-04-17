import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const capacitor = env.VITE_CAPACITOR === '1'
  return {
    plugins: [
      vue(),
      VitePWA({
        registerType: 'autoUpdate',
        manifest: false,
        workbox: {
          globPatterns: ['**/*.{js,css,html,ico,png,svg,webmanifest}'],
          runtimeCaching: [
            {
              urlPattern: /^\/(api|commons|farms|plants|sensors|devices|actuators|schedules|tasks|alerts|zones|costs|profile|naturalfarming|fertigation|crop-cycles|auth|organizations|units|file-attachments)\//,
              handler: 'NetworkFirst',
              options: { cacheName: 'gr33n-api', expiration: { maxEntries: 200, maxAgeSeconds: 3600 } },
            },
            {
              urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp|ico)$/,
              handler: 'StaleWhileRevalidate',
              options: { cacheName: 'gr33n-images', expiration: { maxEntries: 60, maxAgeSeconds: 86400 * 30 } },
            },
            {
              urlPattern: /\.(?:js|css)$/,
              handler: 'StaleWhileRevalidate',
              options: { cacheName: 'gr33n-static', expiration: { maxEntries: 60, maxAgeSeconds: 86400 * 30 } },
            },
          ],
        },
      }),
    ],
    base: capacitor ? './' : '/',
  }
})
