<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import LineChart from '$lib/components/LineChart.svelte';
  import type { Series } from '$lib/components/LineChart';
  import { sessions as sessionsApi, exercises as exApi } from '$lib/api/endpoints';
  import type { Session, Exercise, Difficulty, SessionStatus } from '$lib/api/types';
  import { formatDuration } from '$lib/util/format';

  let items: Session[] = [];
  let exById: Map<string, Exercise> = new Map();
  let loading = true;

  // Edit state.
  let editingId: string | null = null;
  let draft: {
    countCompleted: number | null;
    durationSeconds: number | null;
    weightLb: number | null;
    difficulty: Difficulty | '';
    status: SessionStatus;
    notes: string;
  } | null = null;
  let saving = false;
  let editError = '';

  // Trends range selector.
  type RangeKey = '1w' | '2w' | '1m';
  let range: RangeKey = '1w';
  const RANGE_DAYS: Record<RangeKey, number> = { '1w': 7, '2w': 14, '1m': 30 };
  const RANGES: RangeKey[] = ['1w', '2w', '1m'];
  const RANGE_LABELS: Record<RangeKey, string> = {
    '1w': '1 Week',
    '2w': '2 Weeks',
    '1m': '1 Month'
  };

  // Single-series isolation: clicking a legend entry shows just that one;
  // clicking it again (or "Show all") clears.
  let isolatedSeries: string | null = null;

  function onLegendClick(e: CustomEvent<{ name: string }>) {
    const { name } = e.detail;
    isolatedSeries = isolatedSeries === name ? null : name;
  }

  // Distinct, OLED-friendly palette. Cycles if there are more exercises than colors.
  const PALETTE = [
    '#3b82f6', // blue
    '#a855f7', // purple
    '#22c55e', // green
    '#f97316', // orange
    '#ec4899', // pink
    '#14b8a6', // teal
    '#eab308', // yellow
    '#ef4444', // red
    '#06b6d4', // cyan
    '#84cc16'  // lime
  ];

  onMount(async () => {
    await reload();
  });

  async function reload() {
    loading = true;
    const [s, e] = await Promise.all([sessionsApi.list({ limit: 1000 }), exApi.list(true)]);
    items = s.sessions ?? [];
    exById = new Map((e.exercises ?? []).map((x) => [x.id, x]));
    loading = false;
  }

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

  function startEdit(s: Session) {
    editingId = s.id;
    editError = '';
    draft = {
      countCompleted: s.countCompleted ?? null,
      durationSeconds: s.durationSeconds ?? null,
      weightLb: s.weightLb ?? null,
      difficulty: s.difficulty ?? '',
      status: s.status,
      notes: s.notes ?? ''
    };
  }

  function cancelEdit() {
    editingId = null;
    draft = null;
    editError = '';
  }

  async function saveEdit(s: Session) {
    if (!draft) return;
    saving = true;
    editError = '';
    try {
      const patch: Parameters<typeof sessionsApi.update>[1] = {
        notes: draft.notes,
        status: draft.status
      };
      if (draft.countCompleted !== null && !Number.isNaN(draft.countCompleted)) {
        patch.countCompleted = draft.countCompleted;
      }
      if (draft.durationSeconds !== null && !Number.isNaN(draft.durationSeconds)) {
        patch.durationSeconds = draft.durationSeconds;
      }
      if (draft.weightLb !== null && !Number.isNaN(draft.weightLb)) {
        patch.weightLb = draft.weightLb;
      }
      if (draft.difficulty) patch.difficulty = draft.difficulty;
      const { session: updated } = await sessionsApi.update(s.id, patch);
      items = items.map((x) => (x.id === s.id ? updated : x));
      cancelEdit();
    } catch (err) {
      editError = err instanceof Error ? err.message : String(err);
    } finally {
      saving = false;
    }
  }

  async function deleteSession(s: Session) {
    if (!confirm('Delete this session? This cannot be undone.')) return;
    saving = true;
    editError = '';
    try {
      await sessionsApi.remove(s.id);
      items = items.filter((x) => x.id !== s.id);
      cancelEdit();
    } catch (err) {
      editError = err instanceof Error ? err.message : String(err);
    } finally {
      saving = false;
    }
  }

  // --- Trends ---

  function localDayKey(iso: string): string {
    const d = new Date(iso);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  }

  // Time-to-complete a set in seconds. Failed/incomplete count (longer time =
  // harder); skipped sessions are excluded since the user never attempted.
  function sessionDuration(s: Session): number | null {
    if (s.status === 'SESSION_STATUS_SKIPPED') return null;
    if (s.durationSeconds !== undefined && s.durationSeconds > 0) return s.durationSeconds;
    if (s.endedAt) {
      const ms = new Date(s.endedAt).getTime() - new Date(s.startedAt).getTime();
      const secs = Math.round(ms / 1000);
      return secs > 0 ? secs : null;
    }
    return null;
  }

  $: rangeMs = RANGE_DAYS[range] * 24 * 60 * 60 * 1000;
  $: now = Date.now();
  $: from = now - rangeMs;

  // One series per exercise, with daily-average time-to-complete points.
  $: trendSeries = (() => {
    type Bucket = { sum: number; count: number };
    const byExercise = new Map<string, Map<string, Bucket>>();

    for (const s of items) {
      const t = new Date(s.startedAt).getTime();
      if (t < from || t > now) continue;
      const d = sessionDuration(s);
      if (d === null) continue;
      const dayKey = localDayKey(s.startedAt);
      let m = byExercise.get(s.exerciseId);
      if (!m) {
        m = new Map();
        byExercise.set(s.exerciseId, m);
      }
      const cur = m.get(dayKey) ?? { sum: 0, count: 0 };
      cur.sum += d;
      cur.count += 1;
      m.set(dayKey, cur);
    }

    const out: Series[] = [];
    let i = 0;
    // Order by exercise name for stable color assignment.
    const ids = [...byExercise.keys()].sort((a, b) => {
      const an = exById.get(a)?.name ?? '';
      const bn = exById.get(b)?.name ?? '';
      return an.localeCompare(bn);
    });
    for (const exId of ids) {
      const ex = exById.get(exId);
      if (!ex) continue;
      const m = byExercise.get(exId)!;
      const points = [...m.entries()]
        .map(([key, b]) => {
          const [y, mo, d] = key.split('-').map(Number);
          const x = new Date(y, mo - 1, d, 12, 0, 0).getTime();
          const avg = b.sum / b.count;
          return { x, y: avg, label: `${formatDuration(avg)} (n=${b.count})` };
        })
        .sort((a, b) => a.x - b.x);
      out.push({
        name: ex.name,
        color: PALETTE[i % PALETTE.length],
        points
      });
      i++;
    }
    return out;
  })();

</script>

<div class="app-shell">
  <AppHeader title="" />
  <h1 class="title-large">History</h1>

  {#if !loading && items.length > 0}
    <div class="card chart-card">
      <div class="chart-head">
        <div class="row" style="gap: 0.4rem; align-items: center;">
          <Icon name="chart" size={16} />
          <strong>Time per set</strong>
        </div>
        <div class="row range-tabs" role="tablist">
          {#each RANGES as r}
            <button
              type="button"
              class="range-tab"
              class:active={range === r}
              on:click={() => (range = r)}
              role="tab"
              aria-selected={range === r}
            >{RANGE_LABELS[r]}</button>
          {/each}
        </div>
      </div>
      <LineChart
        series={trendSeries}
        xRange={[from, now]}
        yLabel="seconds"
        yFormat={formatDuration}
        isolatedName={isolatedSeries}
        on:legendClick={onLegendClick}
      />
    </div>
  {/if}

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
      {#each items as s (s.id)}
        {@const b = statusBadge(s.status)}
        {@const ex = exById.get(s.exerciseId)}
        <div class="card">
          {#if editingId === s.id && draft}
            <div class="row" style="margin-bottom: 0.4rem;">
              <strong class="grow">{ex?.name ?? '?'}</strong>
              <span class="muted tabular" style="font-size: 0.85rem;">{fmtDate(s.startedAt)}</span>
            </div>

            {#if ex?.kind === 'EXERCISE_KIND_COUNTED' || ex?.kind === 'EXERCISE_KIND_WEIGHTED'}
              <div class="field">
                <label>Reps completed{ex?.goalCount ? ` (goal ${ex.goalCount})` : ''}</label>
                <input type="number" min="0" bind:value={draft.countCompleted} placeholder="—" />
              </div>
            {/if}

            {#if ex?.kind === 'EXERCISE_KIND_TIMED'}
              <div class="field">
                <label>Duration (seconds){ex?.goalDurationSeconds ? ` (goal ${formatDuration(ex.goalDurationSeconds)})` : ''}</label>
                <input type="number" min="0" bind:value={draft.durationSeconds} placeholder="—" />
              </div>
            {/if}

            {#if ex?.kind === 'EXERCISE_KIND_WEIGHTED'}
              <div class="field">
                <label>Weight (lb)</label>
                <input type="number" min="0" step="0.5" bind:value={draft.weightLb} placeholder="—" />
              </div>
            {/if}

            <div class="field">
              <label>Status</label>
              <select bind:value={draft.status}>
                <option value="SESSION_STATUS_COMPLETED">Completed</option>
                <option value="SESSION_STATUS_FAILED">Failed</option>
                <option value="SESSION_STATUS_SKIPPED">Skipped</option>
              </select>
            </div>

            <div class="field">
              <label>Difficulty</label>
              <select bind:value={draft.difficulty}>
                <option value="">—</option>
                <option value="DIFFICULTY_EASY">Easy</option>
                <option value="DIFFICULTY_MODERATE">Moderate</option>
                <option value="DIFFICULTY_HARD">Hard</option>
                <option value="DIFFICULTY_VERY_HARD">Very hard</option>
                <option value="DIFFICULTY_FAILED">Failed</option>
              </select>
            </div>

            <div class="field">
              <label>Notes</label>
              <input bind:value={draft.notes} />
            </div>

            {#if editError}<p class="error" style="color: var(--bad); font-size: 0.85rem;">{editError}</p>{/if}

            <div class="row" style="gap: 0.5rem; margin-top: 0.75rem;">
              <button class="primary grow" on:click={() => saveEdit(s)} disabled={saving}>
                {saving ? '…' : 'Save'}
              </button>
              <button class="ghost" on:click={cancelEdit} disabled={saving}>Cancel</button>
              <button
                class="ghost icon"
                style="color: var(--bad);"
                on:click={() => deleteSession(s)}
                disabled={saving}
                aria-label="Delete"
              ><Icon name="trash" size={18} /></button>
            </div>
          {:else}
            <div class="row" style="margin-bottom: 0.4rem;">
              <strong class="grow">{ex?.name ?? '?'}</strong>
              <span class="badge {b.cls}">{b.label}</span>
              <button
                class="ghost icon"
                style="margin-left: 0.4rem; min-height: 0; padding: 0.3rem;"
                on:click={() => startEdit(s)}
                aria-label="Edit"
              ><Icon name="pencil" size={16} /></button>
            </div>
            <div class="muted tabular" style="font-size: 0.85rem;">
              {fmtDate(s.startedAt)}
            </div>
            <div class="tabular" style="font-size: 0.9rem; margin-top: 0.4rem;">
              {#if s.setIndex && s.setTotal}Set {s.setIndex}/{s.setTotal} · {/if}
              {#if s.countCompleted !== undefined}{s.countCompleted}{#if s.countGoal} / {s.countGoal}{/if} reps{/if}
              {#if s.durationSeconds !== undefined}{s.countCompleted !== undefined ? ' · ' : ''}{formatDuration(s.durationSeconds)}{/if}
              {#if s.weightLb !== undefined} @ {s.weightLb} lb{/if}
              {#if s.difficulty}· {s.difficulty.replace('DIFFICULTY_', '').toLowerCase()}{/if}
            </div>
            {#if s.notes}
              <p style="margin: 0.4rem 0 0; font-size: 0.92rem; color: var(--text-2);">{s.notes}</p>
            {/if}
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  select { width: 100%; }

  .chart-card { padding: 0.85rem 0.85rem 0.5rem; margin-bottom: 1rem; }

  .chart-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-bottom: 0.6rem;
  }

  .range-tabs {
    padding: 0.2rem;
    background: var(--surface);
    border: 1px solid var(--hairline);
    border-radius: 999px;
    gap: 0;
  }
  .range-tab {
    border: none;
    background: transparent;
    color: var(--muted);
    font-size: 0.8rem;
    font-weight: 600;
    padding: 0.35rem 0.7rem;
    border-radius: 999px;
    min-height: 0;
  }
  .range-tab.active {
    background: var(--surface-3);
    color: var(--text);
  }
</style>
