import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig({
  plugins: [
    sveltekit(),
    SvelteKitPWA({
      registerType: 'autoUpdate',
      manifest: {
        name: 'Training Log',
        short_name: 'Training',
        description: 'Martial arts, calisthenics, and gym training log',
        theme_color: '#00d2a8',
        background_color: '#0b1220',
        display: 'standalone',
        orientation: 'portrait',
        start_url: '/',
        icons: [
          {
            src: '/icons/icon-192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: '/icons/icon-512.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: '/icons/icon-512-maskable.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,png,webp,woff2,mp3,ogg}'],
        navigateFallback: '/',
        navigateFallbackDenylist: [/^\/v1\//],
        // adapter-static writes index.html *after* vite-pwa builds the
        // precache manifest, so the SPA shell ('/') is missing from it. The
        // navigation fallback handler then throws non-precached-url. Add it
        // explicitly with a per-build revision so the new SW re-fetches the
        // shell when we deploy.
        additionalManifestEntries: [
          { url: '/', revision: `${Date.now()}` }
        ],
        // NetworkFirst-cache GET /v1/* so reads work offline. Sessions writes
        // already go through the Dexie outbox; this layer just keeps the SPA
        // alive on iOS PWA resume when TLS is briefly unhappy. Only 200s get
        // cached, so a 401 still propagates to the client (and triggers the
        // refresh dance + login redirect when actually unauthenticated).
        runtimeCaching: [
          {
            urlPattern: /\/v1\//,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-v1',
              networkTimeoutSeconds: 3,
              expiration: { maxEntries: 200, maxAgeSeconds: 60 * 60 * 24 * 7 },
              cacheableResponse: { statuses: [200] }
            }
          }
        ]
      }
    })
  ],
  server: {
    proxy: {
      '/v1': {
        target: 'http://localhost:18080',
        changeOrigin: false
      }
    }
  }
});
