<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { exercises as exApi, routines, machines as machinesApi } from '$lib/api/endpoints';
  import { logSessionResilient } from '$lib/api/outbox';
  import type { Exercise, Routine, Machine, Difficulty } from '$lib/api/types';
  import { nextItem, previousItem, type ItemRunState, type RunMode } from '$lib/runner/state';
  import { cues, unlockAudio } from '$lib/audio/cues';
  import { uuid } from '$lib/util/uuid';
  import DifficultyButtons from '$lib/components/DifficultyButtons.svelte';
  import RepsStepper from '$lib/components/RepsStepper.svelte';
  import Icon from '$lib/components/Icon.svelte';

  let runId: string;
  let routineId: string;
  let routine: Routine | null = null;
  let exercisesById: Map<string, Exercise> = new Map();
  let machinesById: Map<string, Machine> = new Map();

  // Per-item run state. Set progress is preserved across skips.
  let itemStates: ItemRunState[] = [];
  let cursor = -1;
  let mode: RunMode = 'sequential';

  let phase: 'ready' | 'running' | 'awaiting_difficulty' | 'done' = 'ready';

  // Set-level tracking (the set we're presenting now)
  let elapsedSec = 0;
  let timerHandle: number | null = null;
  let setStartedAt = 0;
  let actualCount = 0;
  let actualWeight: number | null = null;
  let note = '';
  let loading = true;

  $: current = cursor >= 0 ? routine?.items[cursor] : null;
  $: currentEx = current ? exercisesById.get(current.exerciseId) : null;
  $: currentMachine = currentEx?.machineId ? machinesById.get(currentEx.machineId) : null;
  $: currentItemState = cursor >= 0 ? itemStates[cursor] : null;
  // The set we're about to do = (sets already done) + 1.
  $: currentSetIndex = currentItemState ? currentItemState.setsDone + 1 : 1;
  $: totalSets = currentItemState?.totalSets ?? 1;

  onMount(async () => {
    runId = $page.params.runId;
    routineId = $page.url.searchParams.get('routineId') ?? '';
    if (!routineId) { await goto('/'); return; }

    const [r, ex, m] = await Promise.all([
      routines.get(routineId),
      exApi.list(),
      machinesApi.list()
    ]);
    routine = r.routine;
    exercisesById = new Map((ex.exercises ?? []).map((e) => [e.id, e]));
    machinesById = new Map((m.machines ?? []).map((x) => [x.id, x]));
    mode = routine.alternateSets ? 'alternating' : 'sequential';

    itemStates = routine.items.map((it) => {
      const eItem = exercisesById.get(it.exerciseId);
      return {
        setsDone: 0,
        totalSets: eItem?.goalSets && eItem.goalSets > 0 ? eItem.goalSets : 1,
        skipped: false
      };
    });

    advance(-1);
    loading = false;
  });

  onDestroy(() => { if (timerHandle !== null) clearInterval(timerHandle); });

  function presentCursor(i: number) {
    cursor = i;
    if (cursor >= 0 && itemStates[cursor].skipped) {
      // Visiting un-skips so the badge clears.
      itemStates[cursor].skipped = false;
      itemStates = [...itemStates];
    }
    actualWeight = currentEx?.goalWeightLb ?? null;
    actualCount = 0;
    elapsedSec = 0;
    note = '';
    phase = 'ready';
  }

  function advance(fromCursor: number) {
    const next = nextItem(itemStates, fromCursor, mode);
    if (next === -1) {
      phase = 'done';
    } else {
      presentCursor(next);
    }
  }

  function startSet() {
    unlockAudio();
    if (!currentEx) return;
    setStartedAt = Date.now();
    elapsedSec = 0;
    phase = 'running';
    cues.start();

    timerHandle = window.setInterval(() => {
      elapsedSec = Math.floor((Date.now() - setStartedAt) / 1000);
      if (currentEx?.kind === 'EXERCISE_KIND_TIMED' && currentEx.goalDurationSeconds &&
          elapsedSec >= currentEx.goalDurationSeconds) {
        finishSet();
      }
    }, 100) as unknown as number;
  }

  function finishSet() {
    if (timerHandle !== null) { clearInterval(timerHandle); timerHandle = null; }
    cues.end();
    // Start at 0 — the user enters what they actually did. "Completed all"
    // button on the difficulty screen is the one-tap full-credit shortcut.
    actualCount = 0; // user enters what they actually did; "Completed all" is the shortcut
    phase = 'awaiting_difficulty';
  }

  async function logCurrentSet(opts: { difficulty: Difficulty }) {
    if (!current || !currentEx || !currentItemState) return;
    const status: 'COMPLETED' | 'FAILED' = opts.difficulty === 'DIFFICULTY_FAILED' ? 'FAILED' : 'COMPLETED';
    const setIdx = currentSetIndex;
    const setTot = totalSets;

    await logSessionResilient({
      clientId: uuid(),
      exerciseId: current.exerciseId,
      routineRunId: runId,
      startedAt: new Date(setStartedAt || Date.now()).toISOString(),
      endedAt: new Date().toISOString(),
      status: `SESSION_STATUS_${status}` as any,
      countCompleted:
        (currentEx.kind === 'EXERCISE_KIND_COUNTED' || currentEx.kind === 'EXERCISE_KIND_WEIGHTED')
          ? actualCount : undefined,
      countGoal: (currentEx.kind === 'EXERCISE_KIND_COUNTED' || currentEx.kind === 'EXERCISE_KIND_WEIGHTED')
        ? currentEx.goalCount : undefined,
      durationSeconds: currentEx.kind === 'EXERCISE_KIND_TIMED' ? elapsedSec : undefined,
      weightLb: currentEx.kind === 'EXERCISE_KIND_WEIGHTED' && actualWeight !== null ? actualWeight : undefined,
      setIndex: setTot > 1 ? setIdx : undefined,
      setTotal: setTot > 1 ? setTot : undefined,
      difficulty: opts.difficulty,
      notes: note
    });

    // Bump set progress; this is durable as soon as the session is logged.
    itemStates[cursor].setsDone += 1;
    itemStates[cursor].skipped = false;
    itemStates = [...itemStates];
    advance(cursor);
  }

  function skipItem() {
    if (cursor < 0) return;
    itemStates[cursor].skipped = true;
    itemStates = [...itemStates];
    advance(cursor);
  }

  function previous() {
    const prev = previousItem(itemStates, cursor);
    if (prev !== -1) presentCursor(prev);
  }

  async function endRun() {
    try { await routines.endRun(runId); } catch {}
    goto('/');
  }

  function chooseDifficulty(d: Difficulty) {
    void logCurrentSet({ difficulty: d });
  }

  function logged() {
    void logCurrentSet({ difficulty: 'DIFFICULTY_MODERATE' });
  }

  function fmt(secs: number) {
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m}:${s.toString().padStart(2, '0')}`;
  }
</script>

<div class="run-shell">
  {#if loading}
    <p class="muted">Loading…</p>
  {:else if phase === 'done'}
    <div class="run-body">
      <span class="icon-circle big" style="color: var(--accent); background: rgba(0, 210, 168, 0.08); border-color: rgba(0, 210, 168, 0.25);">
        <Icon name="check" size={40} />
      </span>
      <h1 style="font-size: 2rem;">Routine complete</h1>
      <p class="muted">Nice work.</p>
    </div>
    <button class="primary big" on:click={endRun}>Finish</button>
  {:else if currentEx && current && currentItemState}
    <div class="run-header">
      <span class="muted tabular">{cursor + 1} / {routine?.items.length}</span>
      {#if mode === 'alternating'}
        <span class="badge accent" style="margin-left: 0.5rem;">Circuit</span>
      {/if}
      <div class="grow"></div>
      <button class="ghost icon" on:click={endRun} aria-label="End routine">
        <Icon name="logout" size={20} />
      </button>
    </div>

    <div style="margin-bottom: 1rem;">
      <h1 class="run-title">{currentEx.name}</h1>
      {#if currentMachine}
        <div class="row" style="margin-top: 0.4rem; gap: 0.5rem;">
          <span class="badge accent"><Icon name="cog" size={12} /> {currentMachine.name}</span>
          {#if currentMachine.location}
            <span class="muted" style="font-size: 0.85rem;">{currentMachine.location}</span>
          {/if}
        </div>
      {/if}
      {#if totalSets > 1}
        <div class="set-pills">
          {#each Array(totalSets) as _, i}
            <span class="set-pill"
                  class:done={i < currentItemState.setsDone}
                  class:active={i === currentItemState.setsDone}>{i + 1}</span>
          {/each}
          <span class="muted" style="margin-left: 0.5rem; font-size: 0.85rem;">
            Set {currentSetIndex} of {totalSets}
          </span>
        </div>
      {/if}
      {#if currentEx.instructions}
        <p class="muted" style="margin-top: 0.5rem;">{currentEx.instructions}</p>
      {/if}
    </div>

    {#if phase === 'ready'}
      <div class="run-body">
        {#if currentEx.kind === 'EXERCISE_KIND_TIMED' && currentEx.goalDurationSeconds}
          <div class="big-number tabular">{fmt(currentEx.goalDurationSeconds)}</div>
          <p class="muted">Goal duration</p>
        {:else if currentEx.kind === 'EXERCISE_KIND_COUNTED' && currentEx.goalCount}
          <div class="big-number">{currentEx.goalCount}</div>
          <p class="muted">Goal reps</p>
        {:else if currentEx.kind === 'EXERCISE_KIND_WEIGHTED'}
          <div class="big-number tabular">
            {currentEx.goalCount ?? '?'}<span class="muted" style="font-size: 0.45em;"> reps</span>
          </div>
          {#if currentEx.goalWeightLb}
            <p class="muted tabular" style="font-size: 1.1rem;">{currentEx.goalWeightLb} lb</p>
          {/if}
        {:else if currentEx.kind === 'EXERCISE_KIND_LOGGED'}
          <p>Mark this as done when complete.</p>
        {/if}
      </div>
      <div class="run-actions">
        {#if currentEx.kind === 'EXERCISE_KIND_LOGGED'}
          <button class="primary big" on:click={logged}>
            <Icon name="check" size={20} /> Done
          </button>
        {:else}
          <button class="primary big" on:click={startSet}>
            <Icon name="play" size={18} /> Start set {currentSetIndex}
          </button>
        {/if}
        <div class="row">
          <button class="ghost" on:click={previous} disabled={previousItem(itemStates, cursor) === -1}>
            <Icon name="chevron-left" size={18} /> Previous
          </button>
          <button class="ghost grow" on:click={skipItem}>Skip exercise</button>
        </div>
      </div>

    {:else if phase === 'running'}
      <div class="run-body">
        <div class="big-number tabular">{fmt(elapsedSec)}</div>
        {#if currentEx.kind === 'EXERCISE_KIND_TIMED' && currentEx.goalDurationSeconds}
          <p class="muted tabular">/ {fmt(currentEx.goalDurationSeconds)}</p>
        {:else if currentEx.kind === 'EXERCISE_KIND_COUNTED' && currentEx.goalCount}
          <p class="muted">{currentEx.goalCount} reps</p>
        {:else if currentEx.kind === 'EXERCISE_KIND_WEIGHTED'}
          <p class="muted">
            {#if currentEx.goalCount}{currentEx.goalCount} reps{/if}
            {#if currentEx.goalWeightLb} @ {currentEx.goalWeightLb} lb{/if}
          </p>
        {:else}
          <p class="muted">elapsed</p>
        {/if}
      </div>
      <div class="run-actions">
        <button class="primary big" on:click={finishSet}>Done</button>
      </div>

    {:else if phase === 'awaiting_difficulty'}
      <div class="run-body" style="align-items: stretch; text-align: left;">
        {#if currentEx.kind === 'EXERCISE_KIND_COUNTED' || currentEx.kind === 'EXERCISE_KIND_WEIGHTED'}
          <div class="field" style="text-align: center;">
            <label style="text-align: left;">Reps completed</label>
            <RepsStepper bind:value={actualCount} goal={currentEx.goalCount ?? null} />
          </div>
        {:else if currentEx.kind === 'EXERCISE_KIND_TIMED'}
          <div class="field">
            <label>Held for</label>
            <div class="big-number tabular" style="font-size: 2.5rem;">{fmt(elapsedSec)}</div>
          </div>
        {/if}
        {#if currentEx.kind === 'EXERCISE_KIND_WEIGHTED'}
          <div class="field">
            <label>Weight (lb)</label>
            <input type="number" min="0" step="0.5" bind:value={actualWeight} />
          </div>
        {/if}
        <div class="field">
          <label>Notes (optional)</label>
          <input bind:value={note} placeholder="Felt strong / form check / etc." />
        </div>
        <p style="margin: 1rem 0 0.5rem;">How hard was it?</p>
        <DifficultyButtons onChoose={chooseDifficulty} />
      </div>
    {/if}
  {/if}
</div>

<style>
  .run-header {
    display: flex;
    align-items: center;
    margin-bottom: 1rem;
  }
  .run-title {
    font-size: 2rem;
    font-weight: 800;
    letter-spacing: -0.025em;
    margin: 0;
    line-height: 1.1;
  }
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
  .set-pill.done {
    background: rgba(0, 210, 168, 0.12);
    border-color: rgba(0, 210, 168, 0.3);
    color: var(--accent);
  }
  .set-pill.active {
    background: var(--accent);
    border-color: var(--accent);
    color: #071a16;
  }
</style>
