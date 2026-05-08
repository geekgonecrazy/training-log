<script lang="ts">
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import LineChart from '$lib/components/LineChart.svelte';
  import { sessions as sessionsApi, exercises as exApi } from '$lib/api/endpoints';
  import type { Session, Exercise, Difficulty, SessionStatus } from '$lib/api/types';

  type Tab = 'list' | 'trends';
  let tab: Tab = 'list';

  let items: Session[] = [];
  let exById: Map<string, Exercise> = new Map();
  let loading = true;

  // Edit state: id of the session being edited + a draft.
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

  // Trends tab.
  let selectedExerciseId: string = '';

  onMount(async () => {
    await reload();
  });

  async function reload() {
    loading = true;
    const [s, e] = await Promise.all([sessionsApi.list({ limit: 500 }), exApi.list(true)]);
    items = s.sessions ?? [];
    exById = new Map((e.exercises ?? []).map((x) => [x.id, x]));
    // Default the trends selector to the most-recent exercise the user logged.
    if (!selectedExerciseId && items.length > 0) {
      selectedExerciseId = items[0].exerciseId;
    }
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
      // Replace in place.
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

  // Time-to-complete a set in seconds. Failed/incomplete still count (see issue
  // discussion: longer time = harder). Skipped sessions are excluded since the
  // user never attempted the set.
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

  $: exercisesWithSessions = (() => {
    const seen = new Set<string>();
    const out: Exercise[] = [];
    for (const s of items) {
      if (seen.has(s.exerciseId)) continue;
      seen.add(s.exerciseId);
      const e = exById.get(s.exerciseId);
      if (e) out.push(e);
    }
    return out.sort((a, b) => a.name.localeCompare(b.name));
  })();

  // For the chart: average time-to-complete per local day, for the selected exercise.
  $: trendPoints = (() => {
    if (!selectedExerciseId) return [] as { x: number; y: number; label: string }[];
    const buckets = new Map<string, { sum: number; count: number; sample: string }>();
    for (const s of items) {
      if (s.exerciseId !== selectedExerciseId) continue;
      const d = sessionDuration(s);
      if (d === null) continue;
      const key = localDayKey(s.startedAt);
      const cur = buckets.get(key) ?? { sum: 0, count: 0, sample: s.startedAt };
      cur.sum += d;
      cur.count += 1;
      buckets.set(key, cur);
    }
    return [...buckets.entries()]
      .map(([key, b]) => {
        // Use noon local time on that day so the x-axis label matches the bucket.
        const [y, m, d] = key.split('-').map(Number);
        const x = new Date(y, m - 1, d, 12, 0, 0).getTime();
        const avg = b.sum / b.count;
        return {
          x,
          y: avg,
          label: `${fmtChartTooltipDate(x)} — ${fmtSecs(avg)} (n=${b.count})`
        };
      })
      .sort((a, b) => a.x - b.x);
  })();

  function fmtSecs(secs: number): string {
    if (secs < 60) return `${Math.round(secs)}s`;
    const m = Math.floor(secs / 60);
    const s = Math.round(secs % 60);
    return s === 0 ? `${m}m` : `${m}m ${s}s`;
  }

  function fmtChartTooltipDate(ms: number): string {
    return new Date(ms).toLocaleDateString([], { month: 'short', day: 'numeric' });
  }
</script>

<div class="app-shell">
  <AppHeader title="" />
  <h1 class="title-large">History</h1>

  <div class="row tabs" style="margin: 0 0 1rem; padding: 0.25rem; background: var(--surface); border: 1px solid var(--hairline); border-radius: 999px;">
    <button
      class="pill"
      style="flex:1; min-height:0; padding:0.5rem; border: none; background: {tab === 'list' ? 'var(--surface-3)' : 'transparent'}; color: {tab === 'list' ? 'var(--text)' : 'var(--muted)'};"
      on:click={() => (tab = 'list')}
    >List</button>
    <button
      class="pill"
      style="flex:1; min-height:0; padding:0.5rem; border: none; background: {tab === 'trends' ? 'var(--surface-3)' : 'transparent'}; color: {tab === 'trends' ? 'var(--text)' : 'var(--muted)'};"
      on:click={() => (tab = 'trends')}
    >Trends</button>
  </div>

  {#if loading}
    <div class="card muted">Loading…</div>
  {:else if tab === 'list'}
    {#if items.length === 0}
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
                  <input
                    type="number"
                    min="0"
                    bind:value={draft.countCompleted}
                    placeholder="—"
                  />
                </div>
              {/if}

              {#if ex?.kind === 'EXERCISE_KIND_TIMED'}
                <div class="field">
                  <label>Duration (seconds){ex?.goalDurationSeconds ? ` (goal ${ex.goalDurationSeconds}s)` : ''}</label>
                  <input
                    type="number"
                    min="0"
                    bind:value={draft.durationSeconds}
                    placeholder="—"
                  />
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
                {#if s.durationSeconds !== undefined}{s.countCompleted !== undefined ? ' · ' : ''}{s.durationSeconds}s{/if}
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
  {:else}
    <!-- Trends tab -->
    {#if exercisesWithSessions.length === 0}
      <div class="empty-state card">
        <span class="icon-circle"><Icon name="chart" size={32} /></span>
        <h2 style="margin-top:0;">Nothing to graph yet</h2>
        <p>Log some sessions and the trend will appear here.</p>
      </div>
    {:else}
      <div class="field">
        <label for="trend-ex">Exercise</label>
        <select id="trend-ex" bind:value={selectedExerciseId}>
          {#each exercisesWithSessions as e}
            <option value={e.id}>{e.name}</option>
          {/each}
        </select>
      </div>

      <div class="card" style="padding: 0.75rem;">
        <p class="muted" style="margin: 0 0 0.5rem; font-size: 0.85rem;">
          Average time per set, by day. Higher = harder. Skipped sets excluded.
        </p>
        <LineChart points={trendPoints} yLabel="seconds" yFormat={fmtSecs} />
        {#if trendPoints.length === 1}
          <p class="muted" style="margin: 0.5rem 0 0; font-size: 0.85rem;">
            Only one day of data — log on another day to see a trend.
          </p>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<style>
  select { width: 100%; }
</style>
