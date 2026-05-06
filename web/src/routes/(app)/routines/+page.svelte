<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import { routines as api } from '$lib/api/endpoints';
  import type { Routine } from '$lib/api/types';

  let items: Routine[] = [];
  let loading = true;

  onMount(async () => {
    const r = await api.list();
    items = r.routines ?? [];
    loading = false;
  });
</script>

<div class="app-shell">
  <AppHeader title="" />
  <div class="row" style="margin-bottom: 1rem;">
    <h1 class="title-large grow" style="margin: 0;">Routines</h1>
    <a class="btn primary pill" href="/routines/new"><Icon name="plus" size={16} />New</a>
  </div>

  {#if loading}
    <div class="card muted">Loading…</div>
  {:else if items.length === 0}
    <div class="empty-state card">
      <span class="icon-circle"><Icon name="list" size={32} /></span>
      <h2 style="margin-top:0;">No routines yet</h2>
      <p>Stack a few exercises into a routine and run them as one workout.</p>
      <a class="btn primary" href="/routines/new" style="margin-top: 0.75rem;">
        <Icon name="plus" size={18} /> Create a routine
      </a>
    </div>
  {:else}
    <div class="list">
      {#each items as r}
        <div class="card row" style="padding: 0.75rem 1rem;">
          <a href="/routines/{r.id}" style="text-decoration: none; color: inherit;" class="row grow">
            <span class="icon-circle"><Icon name="list" size={20} /></span>
            <div class="grow">
              <strong>{r.name}</strong>
              <div class="muted" style="font-size: 0.85rem;">
                {r.items.length} exercise{r.items.length === 1 ? '' : 's'}{r.alternateSets ? ' · circuit' : ''}
              </div>
            </div>
          </a>
          <a class="btn primary pill" href="/routines/{r.id}/start" style="padding: 0.55rem 1rem; min-height: 0;">
            <Icon name="play" size={14} />Start
          </a>
        </div>
      {/each}
    </div>
  {/if}
</div>
