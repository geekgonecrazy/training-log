<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import { exercises as api } from '$lib/api/endpoints';
  import type { Exercise } from '$lib/api/types';
  import { exerciseGoalSummary, kindIcon } from '$lib/util/format';

  let items: Exercise[] = [];
  let loading = true;

  onMount(async () => {
    const r = await api.list();
    items = r.exercises ?? [];
    loading = false;
  });

</script>

<div class="app-shell">
  <AppHeader title="" />
  <div class="row" style="margin-bottom: 1rem;">
    <h1 class="title-large grow" style="margin: 0;">Exercises</h1>
    <a class="btn ghost icon" href="/machines" aria-label="Machines"><Icon name="cog" size={20} /></a>
    <a class="btn primary pill" href="/exercises/new"><Icon name="plus" size={16} />New</a>
  </div>

  {#if loading}
    <div class="card muted">Loading…</div>
  {:else if items.length === 0}
    <div class="empty-state card">
      <span class="icon-circle"><Icon name="dumbbell" size={32} /></span>
      <h2 style="margin-top:0;">No exercises yet</h2>
      <p>Plank, push-ups, kata — whatever you do, define it once and reuse.</p>
      <a class="btn primary" href="/exercises/new" style="margin-top: 0.75rem;">
        <Icon name="plus" size={18} /> Create an exercise
      </a>
    </div>
  {:else}
    <div class="list">
      {#each items as e}
        <div class="card row" style="padding: 0.75rem 1rem;">
          <a href="/exercises/{e.id}" style="text-decoration: none; color: inherit;" class="row grow">
            <span class="icon-circle"><Icon name={kindIcon(e.kind)} size={20} /></span>
            <div class="grow">
              <strong>{e.name}</strong>
              <div class="muted tabular" style="font-size: 0.85rem;">{exerciseGoalSummary(e)}</div>
            </div>
          </a>
          <a class="btn primary pill" href="/exercises/{e.id}/do" style="padding: 0.55rem 1rem; min-height: 0;">
            <Icon name="play" size={14} />Do
          </a>
        </div>
      {/each}
    </div>
  {/if}
</div>
