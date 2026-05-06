<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import RoutineForm from '$lib/components/RoutineForm.svelte';
  import { exercises as exApi, routines } from '$lib/api/endpoints';
  import type { Exercise, Routine } from '$lib/api/types';

  let routine: Routine | null = null;
  let allExercises: Exercise[] = [];
  let loading = true;

  $: id = $page.params.id;

  onMount(async () => {
    const [r, ex] = await Promise.all([routines.get(id), exApi.list()]);
    routine = r.routine;
    allExercises = ex.exercises ?? [];
    loading = false;
  });

  async function archive() {
    if (!confirm('Archive this routine?')) return;
    await routines.archive(id);
    goto('/routines');
  }
</script>

<div class="app-shell">
  <AppHeader title="Edit routine" back="/routines" />

  {#if loading}
    <p class="muted">Loading…</p>
  {:else if routine}
    <RoutineForm
      initial={routine}
      {allExercises}
      submitLabel="Save"
      onSubmit={async ({ name, exerciseIds, alternateSets }) => {
        await routines.update(id, name, exerciseIds, alternateSets);
        goto('/routines');
      }}
    />

    <div style="margin-top: 2rem;">
      <a class="btn primary" href="/routines/{id}/start">Start routine</a>
      <button class="danger" on:click={archive} style="margin-left: 0.5rem;">Archive</button>
    </div>
  {/if}
</div>
