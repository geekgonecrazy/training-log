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
