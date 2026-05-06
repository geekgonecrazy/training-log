<script lang="ts">
  import type { Exercise, ExerciseKind, Machine } from '$lib/api/types';
  import type { ExerciseInput } from '$lib/api/endpoints';

  export let initial: Partial<Exercise> = {};
  export let machines: Machine[] = [];
  export let onSubmit: (input: ExerciseInput) => Promise<void>;
  export let submitLabel = 'Save';

  let name = initial.name ?? '';
  let kind: ExerciseKind = (initial.kind as ExerciseKind) ?? 'EXERCISE_KIND_COUNTED';
  let machineId: string = initial.machineId ?? '';
  let goalCount: number | null = initial.goalCount ?? null;
  let goalDurationSeconds: number | null = initial.goalDurationSeconds ?? null;
  let goalSets: number | null = initial.goalSets ?? 1;
  let goalWeightLb: number | null = initial.goalWeightLb ?? null;
  let instructions = initial.instructions ?? '';
  let busy = false;
  let error = '';

  $: hasReps = kind === 'EXERCISE_KIND_COUNTED' || kind === 'EXERCISE_KIND_WEIGHTED';
  $: hasDuration = kind === 'EXERCISE_KIND_TIMED';
  $: hasWeight = kind === 'EXERCISE_KIND_WEIGHTED';
  // Logged exercises (kata, etc.) are intrinsically one-shot.
  $: hasSets = kind !== 'EXERCISE_KIND_LOGGED';

  async function submit() {
    busy = true;
    error = '';
    try {
      const payload: ExerciseInput = { name, kind, instructions };
      if (machineId) payload.machineId = machineId;
      if (hasReps && goalCount !== null) payload.goalCount = goalCount;
      if (hasDuration && goalDurationSeconds !== null) payload.goalDurationSeconds = goalDurationSeconds;
      if (hasSets && goalSets !== null && goalSets > 0) payload.goalSets = goalSets;
      if (hasWeight && goalWeightLb !== null) payload.goalWeightLb = goalWeightLb;
      await onSubmit(payload);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  }
</script>

<form on:submit|preventDefault={submit} class="stack">
  <div class="field">
    <label for="name">Name</label>
    <input id="name" bind:value={name} required placeholder="Push-ups" />
  </div>

  <div class="field">
    <label for="kind">Type</label>
    <select id="kind" bind:value={kind}>
      <option value="EXERCISE_KIND_TIMED">Timed (e.g. plank)</option>
      <option value="EXERCISE_KIND_COUNTED">Counted (e.g. push-ups)</option>
      <option value="EXERCISE_KIND_WEIGHTED">Weighted (gym)</option>
      <option value="EXERCISE_KIND_LOGGED">Logged (yes/no)</option>
    </select>
  </div>

  {#if hasSets}
    <div class="field">
      <label for="goalSets">Sets</label>
      <input id="goalSets" type="number" min="1" max="20" bind:value={goalSets} placeholder="3" />
    </div>
  {/if}

  {#if hasReps}
    <div class="field">
      <label for="goalCount">Reps per set</label>
      <input id="goalCount" type="number" min="1" bind:value={goalCount} placeholder="20" />
    </div>
  {/if}

  {#if hasDuration}
    <div class="field">
      <label for="goalDuration">Seconds per set</label>
      <input id="goalDuration" type="number" min="1" bind:value={goalDurationSeconds} placeholder="30" />
    </div>
  {/if}

  {#if hasWeight}
    <div class="field">
      <label for="goalWeight">Weight (lb)</label>
      <input id="goalWeight" type="number" min="0" step="0.5" bind:value={goalWeightLb} placeholder="135" />
    </div>
  {/if}

  <div class="field">
    <label for="machine">Machine (optional)</label>
    {#if machines.length === 0}
      <p class="muted" style="font-size: 0.85rem; margin: 0.25rem 0 0;">
        No machines yet. <a href="/machines">Add one →</a>
      </p>
    {:else}
      <select id="machine" bind:value={machineId}>
        <option value="">— None —</option>
        {#each machines as m}
          <option value={m.id}>{m.name}</option>
        {/each}
      </select>
    {/if}
  </div>

  <div class="field">
    <label for="instructions">Instructions / notes</label>
    <textarea id="instructions" rows="3" bind:value={instructions} placeholder="Form cues, range of motion, etc."></textarea>
  </div>

  <button class="primary" type="submit" disabled={busy}>{busy ? '…' : submitLabel}</button>
  {#if error}<p class="error">{error}</p>{/if}
</form>
