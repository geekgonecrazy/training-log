<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import ExerciseForm from '$lib/components/ExerciseForm.svelte';
  import { exercises, machines as machinesApi } from '$lib/api/endpoints';
  import type { Exercise, Machine } from '$lib/api/types';

  let exercise: Exercise | null = null;
  let machines: Machine[] = [];
  let loading = true;

  $: id = $page.params.id;

  onMount(async () => {
    const [e, m] = await Promise.all([exercises.get(id), machinesApi.list()]);
    exercise = e.exercise;
    machines = m.machines ?? [];
    loading = false;
  });

  async function archive() {
    if (!confirm('Archive this exercise?')) return;
    await exercises.archive(id);
    goto('/exercises');
  }
</script>

<div class="app-shell">
  <AppHeader title="Edit exercise" back="/exercises" />

  {#if loading}
    <p class="muted">Loading…</p>
  {:else if exercise}
    <ExerciseForm
      initial={exercise}
      {machines}
      submitLabel="Save"
      onSubmit={async (input) => {
        await exercises.update(id, input);
        goto('/exercises');
      }}
    />

    <div style="margin-top: 2rem;">
      <a class="btn ghost" href="/exercises/{id}/do">Do this exercise</a>
      <button class="danger" on:click={archive} style="margin-left: 0.5rem;">Archive</button>
    </div>
  {/if}
</div>
