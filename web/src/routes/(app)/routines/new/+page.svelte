<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import RoutineForm from '$lib/components/RoutineForm.svelte';
  import { exercises as exApi, routines } from '$lib/api/endpoints';
  import type { Exercise } from '$lib/api/types';

  let allExercises: Exercise[] = [];

  onMount(async () => {
    const r = await exApi.list();
    allExercises = r.exercises ?? [];
  });
</script>

<div class="app-shell">
  <AppHeader title="New routine" back="/routines" />
  <RoutineForm
    {allExercises}
    submitLabel="Create"
    onSubmit={async ({ name, exerciseIds, alternateSets }) => {
      await routines.create(name, exerciseIds, alternateSets);
      goto('/routines');
    }}
  />
</div>
