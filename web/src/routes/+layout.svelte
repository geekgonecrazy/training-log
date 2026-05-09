<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { authStore } from '$lib/stores/auth';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';

  // Root layout: only handles auth gating + global CSS. Group layouts handle
  // chrome (tabbar / fullscreen). The route file system maps:
  //   /login                    → src/routes/login/+page.svelte
  //   /, /exercises, etc.       → src/routes/(app)/...   (with tabbar)
  //   /run/:id, /:exId/do, ...  → src/routes/(runner)/... (no tabbar)

  let initialized = false;

  onMount(async () => {
    await authStore.refresh();
    initialized = true;

    // Register the service worker so the app shell is cached. Critical on iOS
    // standalone PWAs: when the OS kills the webview after backgrounding, the
    // first nav on resume can hit a flaky TLS state and Safari shows
    // "Unable to establish a secure connection". A registered SW serves the
    // cached shell instead of a hard error page.
    if ('serviceWorker' in navigator) {
      try {
        const { registerSW } = await import('virtual:pwa-register');
        registerSW({ immediate: true });
      } catch {
        // PWA module not available (dev or unsupported env) — ignore.
      }
    }
  });

  $: isLogin = $page.url.pathname === '/login';

  $: if (initialized && !$authStore.user && !isLogin) {
    goto('/login');
  }
  $: if (initialized && $authStore.user && isLogin) {
    goto('/');
  }
</script>

{#if !initialized}
  <div class="app-shell">
    <p class="muted">Loading…</p>
  </div>
{:else if $authStore.user || isLogin}
  <slot />
{/if}
