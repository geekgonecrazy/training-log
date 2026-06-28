<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import { reports } from '$lib/api/endpoints';
  import type { PeriodReport } from '$lib/api/types';
  import { formatDuration } from '$lib/util/format';

  let mode: 'weekly' | 'monthly' = 'weekly';
  let report: PeriodReport | null = null;
  let loading = true;

  async function load() {
    loading = true;
    const r = mode === 'weekly' ? await reports.weekly() : await reports.monthly();
    report = r.report;
    loading = false;
  }

  onMount(load);

  $: if (mode) load();

  function difficultyLabel(n: number) {
    if (n === 0) return '—';
    const labels = ['', 'Easy', 'Moderate', 'Hard', 'Very hard', 'Failed'];
    const rounded = Math.round(n);
    return `${labels[rounded] ?? '?'} (${n.toFixed(1)})`;
  }
</script>

<div class="app-shell">
  <AppHeader title="" />
  <h1 class="title-large">Reports</h1>

  <div class="row" style="margin: 0 0 1rem; padding: 0.25rem; background: var(--surface); border: 1px solid var(--hairline); border-radius: 999px;">
    <button
      class="pill"
      style="flex:1; min-height:0; padding:0.5rem; border: none; background: {mode === 'weekly' ? 'var(--surface-3)' : 'transparent'}; color: {mode === 'weekly' ? 'var(--text)' : 'var(--muted)'};"
      on:click={() => (mode = 'weekly')}
    >Weekly</button>
    <button
      class="pill"
      style="flex:1; min-height:0; padding:0.5rem; border: none; background: {mode === 'monthly' ? 'var(--surface-3)' : 'transparent'}; color: {mode === 'monthly' ? 'var(--text)' : 'var(--muted)'};"
      on:click={() => (mode = 'monthly')}
    >Monthly</button>
  </div>

  {#if loading || !report}
    <div class="card muted">Loading…</div>
  {:else}
    <p class="subtitle tabular">{report.period}</p>
    {#if report.rollups.length === 0}
      <div class="empty-state card">
        <span class="icon-circle"><Icon name="chart" size={32} /></span>
        <h2 style="margin-top:0;">No sessions in this period</h2>
        <p>Log a workout and it'll show up here.</p>
      </div>
    {:else}
      <div class="list">
        {#each report.rollups as r}
          <div class="card">
            <div class="row" style="margin-bottom: 0.5rem;">
              <strong class="grow">{r.exerciseName}</strong>
              <span class="badge accent">{r.sessionsCompleted} done</span>
            </div>
            <div class="muted tabular" style="font-size: 0.85rem;">
              {#if r.sessionsSkipped > 0}{r.sessionsSkipped} skipped · {/if}
              {#if r.sessionsFailed > 0}{r.sessionsFailed} failed · {/if}
              difficulty {difficultyLabel(r.avgDifficulty)}
            </div>
            <div class="row tabular" style="margin-top: 0.5rem; font-size: 0.95rem;">
              {#if r.totalCount > 0}<span><strong>{r.totalCount}</strong> <span class="muted">reps</span></span>{/if}
              {#if r.totalDurationSeconds > 0}<span><strong>{formatDuration(r.totalDurationSeconds)}</strong></span>{/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
