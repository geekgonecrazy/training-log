<script lang="ts">
  import { authStore } from '$lib/stores/auth';
  import { goto } from '$app/navigation';
  import Icon from './Icon.svelte';

  export let title: string = '';
  export let back: string | null = null;
  export let showLogout: boolean = false;
  // When variant === 'large', the title renders below the bar in display size.
  export let variant: 'compact' | 'large' = 'compact';

  async function logout() {
    await authStore.logout();
    goto('/login');
  }
</script>

<header class="app-bar">
  {#if back}
    <a href={back} class="btn icon ghost" aria-label="Back">
      <Icon name="chevron-left" size={22} />
    </a>
  {/if}

  {#if variant === 'compact'}
    <h1>{title}</h1>
  {/if}
  <div class="grow"></div>

  {#if showLogout}
    <button class="btn icon ghost" on:click={logout} title="Sign out" aria-label="Sign out">
      <Icon name="logout" size={20} />
    </button>
  {/if}
</header>

{#if variant === 'large' && title}
  <h1 class="title-large">{title}</h1>
{/if}
