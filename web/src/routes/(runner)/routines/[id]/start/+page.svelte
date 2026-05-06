<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { routines } from '$lib/api/endpoints';

  let error = '';

  onMount(async () => {
    const id = $page.params.id;
    try {
      const { run } = await routines.start(id);
      goto(`/run/${run.id}?routineId=${id}`, { replaceState: true });
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    }
  });
</script>

<div class="app-shell">
  {#if error}
    <p class="error">Failed to start: {error}</p>
    <a class="btn ghost" href="/routines">Back to routines</a>
  {:else}
    <p class="muted">Starting…</p>
  {/if}
</div>
