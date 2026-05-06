<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { exercises as exApi, machines as machinesApi } from '$lib/api/endpoints';
  import { logSessionResilient } from '$lib/api/outbox';
  import type { Exercise, Machine, Difficulty } from '$lib/api/types';
  import { cues, unlockAudio } from '$lib/audio/cues';
  import { uuid } from '$lib/util/uuid';
  import DifficultyButtons from '$lib/components/DifficultyButtons.svelte';
  import RepsStepper from '$lib/components/RepsStepper.svelte';
  import Icon from '$lib/components/Icon.svelte';

  let ex: Exercise | null = null;
  let machine: Machine | null = null;
  let phase: 'ready' | 'running' | 'awaiting_difficulty' | 'done' = 'ready';

  let currentSet = 1;
  let totalSets = 1;

  let elapsedSec = 0;
  let actualCount = 0;
  let actualWeight: number | null = null;
  let note = '';
  let setStartedAt = 0;
  let timerHandle: number | null = null;

  onMount(async () => {
    const id = $page.params.id;
    const r = await exApi.get(id);
    ex = r.exercise;
    totalSets = ex.goalSets && ex.goalSets > 0 ? ex.goalSets : 1;
    actualWeight = ex.goalWeightLb ?? null;
    if (ex.machineId) {
      const m = await machinesApi.list();
      machine = (m.machines ?? []).find((x) => x.id === ex!.machineId) ?? null;
    }
  });

  onDestroy(() => { if (timerHandle !== null) clearInterval(timerHandle); });

  function startSet() {
    if (!ex) return;
    unlockAudio();
    setStartedAt = Date.now();
    elapsedSec = 0;
    phase = 'running';
    cues.start();
    timerHandle = window.setInterval(() => {
      elapsedSec = Math.floor((Date.now() - setStartedAt) / 1000);
      if (ex?.kind === 'EXERCISE_KIND_TIMED' && ex.goalDurationSeconds &&
          elapsedSec >= ex.goalDurationSeconds) {
        finishSet();
      }
    }, 100) as unknown as number;
  }

  function finishSet() {
    if (timerHandle !== null) { clearInterval(timerHandle); timerHandle = null; }
    cues.end();
    actualCount = 0;
    phase = 'awaiting_difficulty';
  }

  async function chooseDifficulty(d: Difficulty) {
    if (!ex) return;
    await logSessionResilient({
      clientId: uuid(),
      exerciseId: ex.id,
      startedAt: new Date(setStartedAt || Date.now()).toISOString(),
      endedAt: new Date().toISOString(),
      status: d === 'DIFFICULTY_FAILED' ? 'SESSION_STATUS_FAILED' : 'SESSION_STATUS_COMPLETED',
      countCompleted: (ex.kind === 'EXERCISE_KIND_COUNTED' || ex.kind === 'EXERCISE_KIND_WEIGHTED')
        ? actualCount : undefined,
      countGoal: (ex.kind === 'EXERCISE_KIND_COUNTED' || ex.kind === 'EXERCISE_KIND_WEIGHTED') ? ex.goalCount : undefined,
      durationSeconds: ex.kind === 'EXERCISE_KIND_TIMED' ? elapsedSec : undefined,
      weightLb: ex.kind === 'EXERCISE_KIND_WEIGHTED' && actualWeight !== null ? actualWeight : undefined,
      setIndex: totalSets > 1 ? currentSet : undefined,
      setTotal: totalSets > 1 ? totalSets : undefined,
      difficulty: d,
      notes: note
    });
    note = '';
    elapsedSec = 0;
    actualCount = 0;
    if (currentSet < totalSets) {
      currentSet += 1;
      phase = 'ready';
    } else {
      phase = 'done';
    }
  }

  async function logged() {
    if (!ex) return;
    await logSessionResilient({
      clientId: uuid(),
      exerciseId: ex.id,
      startedAt: new Date().toISOString(),
      endedAt: new Date().toISOString(),
      status: 'SESSION_STATUS_COMPLETED',
      difficulty: 'DIFFICULTY_MODERATE',
      notes: note
    });
    phase = 'done';
  }

  function fmt(secs: number) {
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m}:${s.toString().padStart(2, '0')}`;
  }
</script>

<div class="run-shell">
  {#if !ex}
    <p class="muted">Loading…</p>
  {:else if phase === 'done'}
    <div class="run-body">
      <span class="icon-circle big" style="color: var(--accent); background: rgba(0, 210, 168, 0.08); border-color: rgba(0, 210, 168, 0.25);">
        <Icon name="check" size={40} />
      </span>
      <h1 style="font-size: 2rem;">Logged</h1>
      <p class="muted">{ex.name}{totalSets > 1 ? ` · ${totalSets} sets` : ''}</p>
    </div>
    <div class="run-actions">
      <button class="primary big" on:click={() => goto('/')}>Home</button>
      <button class="ghost" on:click={() => location.reload()}>Do again</button>
    </div>
  {:else}
    <div class="run-header">
      <a class="ghost btn icon" href="/" aria-label="Cancel"><Icon name="chevron-left" size={20} /></a>
      <div class="grow"></div>
    </div>

    <div style="margin-bottom: 1rem;">
      <h1 class="run-title">{ex.name}</h1>
      {#if machine}
        <div class="row" style="margin-top: 0.4rem;">
          <span class="badge accent"><Icon name="cog" size={12} /> {machine.name}</span>
        </div>
      {/if}
      {#if totalSets > 1}
        <div class="set-pills">
          {#each Array(totalSets) as _, i}
            <span class="set-pill" class:done={i + 1 < currentSet} class:active={i + 1 === currentSet}>{i + 1}</span>
          {/each}
          <span class="muted" style="margin-left: 0.5rem; font-size: 0.85rem;">Set {currentSet} of {totalSets}</span>
        </div>
      {/if}
      {#if ex.instructions}
        <p class="muted" style="margin-top: 0.5rem;">{ex.instructions}</p>
      {/if}
    </div>

    {#if phase === 'ready'}
      <div class="run-body">
        {#if ex.kind === 'EXERCISE_KIND_TIMED' && ex.goalDurationSeconds}
          <div class="big-number tabular">{fmt(ex.goalDurationSeconds)}</div>
        {:else if ex.kind === 'EXERCISE_KIND_COUNTED' && ex.goalCount}
          <div class="big-number">{ex.goalCount}</div>
          <p class="muted">Goal reps</p>
        {:else if ex.kind === 'EXERCISE_KIND_WEIGHTED'}
          <div class="big-number tabular">
            {ex.goalCount ?? '?'}<span class="muted" style="font-size: 0.45em;"> reps</span>
          </div>
          {#if ex.goalWeightLb}<p class="muted tabular">{ex.goalWeightLb} lb</p>{/if}
        {/if}
      </div>
      {#if ex.kind === 'EXERCISE_KIND_LOGGED'}
        <button class="primary big" on:click={logged}><Icon name="check" size={20} /> Done</button>
      {:else}
        <button class="primary big" on:click={startSet}>
          <Icon name="play" size={18} /> Start set {currentSet}
        </button>
      {/if}
    {:else if phase === 'running'}
      <div class="run-body">
        <div class="big-number tabular">{fmt(elapsedSec)}</div>
        {#if ex.kind === 'EXERCISE_KIND_TIMED' && ex.goalDurationSeconds}
          <p class="muted tabular">/ {fmt(ex.goalDurationSeconds)}</p>
        {:else if ex.kind === 'EXERCISE_KIND_COUNTED' && ex.goalCount}
          <p class="muted">{ex.goalCount} reps</p>
        {:else if ex.kind === 'EXERCISE_KIND_WEIGHTED'}
          <p class="muted">
            {#if ex.goalCount}{ex.goalCount} reps{/if}
            {#if ex.goalWeightLb} @ {ex.goalWeightLb} lb{/if}
          </p>
        {:else}
          <p class="muted">elapsed</p>
        {/if}
      </div>
      <button class="primary big" on:click={finishSet}>Done</button>
    {:else if phase === 'awaiting_difficulty'}
      <div class="run-body" style="align-items: stretch; text-align: left;">
        {#if ex.kind === 'EXERCISE_KIND_COUNTED' || ex.kind === 'EXERCISE_KIND_WEIGHTED'}
          <div class="field" style="text-align: center;">
            <label style="text-align: left;">Reps completed</label>
            <RepsStepper bind:value={actualCount} goal={ex.goalCount ?? null} />
          </div>
        {:else if ex.kind === 'EXERCISE_KIND_TIMED'}
          <div class="field">
            <label>Held for</label>
            <div class="big-number tabular" style="font-size: 2.5rem;">{fmt(elapsedSec)}</div>
          </div>
        {/if}
        {#if ex.kind === 'EXERCISE_KIND_WEIGHTED'}
          <div class="field">
            <label>Weight (lb)</label>
            <input type="number" min="0" step="0.5" bind:value={actualWeight} />
          </div>
        {/if}
        <div class="field">
          <label>Notes</label>
          <input bind:value={note} />
        </div>
        <p style="margin: 1rem 0 0.5rem;">How hard was it?</p>
        <DifficultyButtons onChoose={chooseDifficulty} />
      </div>
    {/if}
  {/if}
</div>

<style>
  .run-header { display: flex; align-items: center; margin-bottom: 1rem; }
  .run-title { font-size: 2rem; font-weight: 800; letter-spacing: -0.025em; margin: 0; line-height: 1.1; }
  .run-body {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 1rem 0;
  }
  .run-actions { display: flex; flex-direction: column; gap: 0.5rem; }
  .big-number {
    font-size: 5.5rem;
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.03em;
  }
  button.big { padding: 1.1rem; font-size: 1.1rem; font-weight: 700; }
  .icon-circle.big { width: 80px; height: 80px; border-radius: 28px; margin-bottom: 1rem; }
  .icon-circle.big :global(svg) { width: 40px; height: 40px; }

  .set-pills { display: flex; align-items: center; gap: 6px; margin-top: 0.75rem; flex-wrap: wrap; }
  .set-pill {
    width: 28px; height: 28px; border-radius: 8px;
    display: inline-flex; align-items: center; justify-content: center;
    font-size: 0.85rem; font-weight: 700;
    background: var(--surface-2); border: 1px solid var(--hairline);
    color: var(--muted);
  }
  .set-pill.done { background: rgba(0, 210, 168, 0.12); border-color: rgba(0, 210, 168, 0.3); color: var(--accent); }
  .set-pill.active { background: var(--accent); border-color: var(--accent); color: #071a16; }
</style>
