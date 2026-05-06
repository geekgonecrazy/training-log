<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import { sessions as sessionsApi, exercises as exApi } from '$lib/api/endpoints';
  import type { Session, Exercise } from '$lib/api/types';

  let items: Session[] = [];
  let exById: Map<string, Exercise> = new Map();
  let loading = true;

  onMount(async () => {
    const [s, e] = await Promise.all([sessionsApi.list({ limit: 200 }), exApi.list(true)]);
    items = s.sessions ?? [];
    exById = new Map((e.exercises ?? []).map((x) => [x.id, x]));
    loading = false;
  });

  function fmtDate(iso: string) {
    const d = new Date(iso);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);
    const sameDay = (a: Date, b: Date) =>
      a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
    const time = d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
    if (sameDay(d, today)) return `Today, ${time}`;
    if (sameDay(d, yesterday)) return `Yesterday, ${time}`;
    return d.toLocaleDateString([], { month: 'short', day: 'numeric' }) + `, ${time}`;
  }

  function statusBadge(s: string): { label: string; cls: string } {
    switch (s) {
      case 'SESSION_STATUS_COMPLETED': return { label: 'Done', cls: 'good' };
      case 'SESSION_STATUS_SKIPPED': return { label: 'Skipped', cls: '' };
      case 'SESSION_STATUS_FAILED': return { label: 'Failed', cls: 'bad' };
      default: return { label: s.replace('SESSION_STATUS_', '').toLowerCase(), cls: '' };
    }
  }
</script>

<div class="app-shell">
  <AppHeader title="" />
  <h1 class="title-large">History</h1>

  {#if loading}
    <div class="card muted">Loading…</div>
  {:else if items.length === 0}
    <div class="empty-state card">
      <span class="icon-circle"><Icon name="clock" size={32} /></span>
      <h2 style="margin-top:0;">No sessions yet</h2>
      <p>Logged workouts show up here.</p>
    </div>
  {:else}
    <div class="list">
      {#each items as s}
        {@const b = statusBadge(s.status)}
        <div class="card">
          <div class="row" style="margin-bottom: 0.4rem;">
            <strong class="grow">{exById.get(s.exerciseId)?.name ?? '?'}</strong>
            <span class="badge {b.cls}">{b.label}</span>
          </div>
          <div class="muted tabular" style="font-size: 0.85rem;">
            {fmtDate(s.startedAt)}
          </div>
          <div class="tabular" style="font-size: 0.9rem; margin-top: 0.4rem;">
            {#if s.setIndex && s.setTotal}Set {s.setIndex}/{s.setTotal} · {/if}
            {#if s.countCompleted !== undefined}{s.countCompleted}{#if s.countGoal} / {s.countGoal}{/if} reps{/if}
            {#if s.durationSeconds !== undefined}{s.countCompleted !== undefined ? ' · ' : ''}{s.durationSeconds}s{/if}
            {#if s.weightLb !== undefined} @ {s.weightLb} lb{/if}
            {#if s.difficulty}· {s.difficulty.replace('DIFFICULTY_', '').toLowerCase()}{/if}
          </div>
          {#if s.notes}
            <p style="margin: 0.4rem 0 0; font-size: 0.92rem; color: var(--text-2);">{s.notes}</p>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
