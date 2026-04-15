/* gr33n PWA — minimal offline shell; API calls stay network-first (tasks slice can be extended later). */
const CACHE = 'gr33n-shell-v1'

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll(['/', '/index.html', '/manifest.webmanifest']))
  )
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim())
})

self.addEventListener('fetch', (event) => {
  const req = event.request
  if (req.method !== 'GET') return
  const url = new URL(req.url)
  if (url.origin !== self.location.origin) return
  if (url.pathname.startsWith('/src/') || url.pathname.startsWith('/@')) return

  event.respondWith(
    fetch(req)
      .then((res) => {
        if (res.ok && (req.mode === 'navigate' || url.pathname === '/')) {
          const copy = res.clone()
          caches.open(CACHE).then((c) => c.put('/', copy))
        }
        return res
      })
      .catch(() => caches.match('/').then((c) => c || new Response('Offline', { status: 503 })))
  )
})
