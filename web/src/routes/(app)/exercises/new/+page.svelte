<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import ExerciseForm from '$lib/components/ExerciseForm.svelte';
  import { exercises, machines as machinesApi } from '$lib/api/endpoints';
  import type { Machine } from '$lib/api/types';

  let machines: Machine[] = [];

  onMount(async () => {
    const r = await machinesApi.list();
    machines = r.machines ?? [];
  });
</script>

<div class="app-shell">
  <AppHeader title="New exercise" back="/exercises" />
  <ExerciseForm
    {machines}
    submitLabel="Create"
    onSubmit={async (input) => {
      await exercises.create(input);
      goto('/exercises');
    }}
  />
</div>
