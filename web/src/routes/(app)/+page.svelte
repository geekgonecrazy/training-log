<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import { exercises as exercisesApi, routines as routinesApi, reports as reportsApi } from '$lib/api/endpoints';
  import type { Exercise, Routine, ProgressionSuggestion } from '$lib/api/types';
  import { setupOutboxAutoFlush, flushOutbox, outboxCount } from '$lib/api/outbox';
  import { authStore } from '$lib/stores/auth';
  import { exerciseGoalSummary, kindIcon } from '$lib/util/format';

  let exs: Exercise[] = [];
  let rts: Routine[] = [];
  let suggestions: ProgressionSuggestion[] = [];
  let pending = 0;
  let loading = true;

  onMount(async () => {
    setupOutboxAutoFlush();
    const [a, b, c] = await Promise.all([
      exercisesApi.list(),
      routinesApi.list(),
      reportsApi.progression()
    ]);
    exs = a.exercises ?? [];
    rts = b.routines ?? [];
    suggestions = c.suggestions ?? [];
    pending = await outboxCount();
    if (navigator.onLine && pending > 0) {
      const r = await flushOutbox();
      pending = r.remaining;
    }
    loading = false;
  });

  function greeting(): string {
    const h = new Date().getHours();
    if (h < 5) return 'Late night';
    if (h < 12) return 'Good morning';
    if (h < 17) return 'Good afternoon';
    if (h < 21) return 'Good evening';
    return 'Good night';
  }

  $: firstName = $authStore.user?.name?.split(' ')[0] ?? '';
</script>

<div class="app-shell">
  <AppHeader title="" showLogout />

  <h1 class="title-large">{greeting()}{firstName ? `, ${firstName}` : ''}</h1>
  <p class="subtitle">Let's get a workout in.</p>

  {#if pending > 0}
    <div class="card row" style="margin-bottom: 1rem; gap: 0.75rem;">
      <span class="icon-circle" style="color: var(--warn); background: rgba(245,158,11,0.08); border-color: rgba(245,158,11,0.25);">
        <Icon name="cloud-off" size={20} />
      </span>
      <div class="grow">
        <strong>{pending}</strong> session{pending === 1 ? '' : 's'} waiting to sync
        <div class="muted" style="font-size: 0.85rem;">Will upload when you're back online.</div>
      </div>
    </div>
  {/if}

  {#if suggestions.length > 0}
    <div class="card" style="margin-bottom: 1rem;">
      <div class="row" style="margin-bottom: 0.75rem;">
        <span class="icon-circle"><Icon name="sparkle" size={20} /></span>
        <h3 class="grow">Ready to bump it up?</h3>
      </div>
      <div class="list">
        {#each suggestions as s}
          <div class="row">
            <div class="grow">
              <strong>{s.exerciseName}</strong>
              <div class="muted" style="font-size: 0.85rem;">{s.recommendation}</div>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <h2>Routines</h2>
  {#if loading}
    <div class="card muted">Loading…</div>
  {:else if rts.length === 0}
    <div class="empty-state card">
      <span class="icon-circle"><Icon name="list" size={32} /></span>
      <h2 style="margin-top:0;">No routines yet</h2>
      <p>Group exercises into a routine and run them back-to-back.</p>
      <a class="btn primary" href="/routines/new" style="margin-top: 0.75rem;">
        <Icon name="plus" size={18} />
        Create a routine
      </a>
    </div>
  {:else}
    <div class="list">
      {#each rts as r}
        <div class="card row" style="padding: 0.75rem 1rem;">
          <span class="icon-circle"><Icon name="list" size={20} /></span>
          <div class="grow">
            <strong>{r.name}</strong>
            <div class="muted" style="font-size: 0.85rem;">
              {r.items.length} exercise{r.items.length === 1 ? '' : 's'}
            </div>
          </div>
          <a class="btn primary pill" href="/routines/{r.id}/start" style="padding: 0.55rem 1rem; min-height: 0;">
            <Icon name="play" size={14} />
            Start
          </a>
        </div>
      {/each}
    </div>
  {/if}

  <h2>Quick log</h2>
  {#if exs.length === 0}
    <div class="empty-state card">
      <span class="icon-circle"><Icon name="dumbbell" size={32} /></span>
      <h2 style="margin-top:0;">No exercises yet</h2>
      <p>Add a few exercises to start logging.</p>
      <a class="btn primary" href="/exercises/new" style="margin-top: 0.75rem;">
        <Icon name="plus" size={18} />
        Create an exercise
      </a>
    </div>
  {:else}
    <div class="list">
      {#each exs as e}
        <a class="card row tappable" href="/exercises/{e.id}/do" style="text-decoration: none; color: inherit; padding: 0.75rem 1rem;">
          <span class="icon-circle"><Icon name={kindIcon(e.kind)} size={20} /></span>
          <div class="grow">
            <strong>{e.name}</strong>
            <div class="muted tabular" style="font-size: 0.85rem;">{exerciseGoalSummary(e)}</div>
          </div>
          <span class="muted"><Icon name="chevron-right" size={20} /></span>
        </a>
      {/each}
    </div>
  {/if}
</div>
