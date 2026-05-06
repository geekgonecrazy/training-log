<script lang="ts">
  import type { Routine, Exercise } from '$lib/api/types';

  export let initial: Partial<Routine> = {};
  export let allExercises: Exercise[] = [];
  export let onSubmit: (input: { name: string; exerciseIds: string[]; alternateSets: boolean }) => Promise<void>;
  export let submitLabel = 'Save';

  let name = initial.name ?? '';
  let selectedIds: string[] = (initial.items ?? []).map((it) => it.exerciseId);
  let alternateSets: boolean = initial.alternateSets ?? false;
  let busy = false;
  let error = '';

  $: exById = new Map(allExercises.map((e) => [e.id, e]));

  function addExercise(id: string) {
    if (!id) return;
    selectedIds = [...selectedIds, id];
  }

  function remove(idx: number) {
    selectedIds = selectedIds.filter((_, i) => i !== idx);
  }

  function move(idx: number, dir: -1 | 1) {
    const next = idx + dir;
    if (next < 0 || next >= selectedIds.length) return;
    const arr = [...selectedIds];
    [arr[idx], arr[next]] = [arr[next], arr[idx]];
    selectedIds = arr;
  }

  async function submit() {
    busy = true;
    error = '';
    try {
      await onSubmit({ name, exerciseIds: selectedIds, alternateSets });
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  }

  let pickerValue = '';
  $: if (pickerValue) {
    addExercise(pickerValue);
    pickerValue = '';
  }
</script>

<form on:submit|preventDefault={submit} class="stack">
  <div class="field">
    <label for="name">Name</label>
    <input id="name" bind:value={name} required />
  </div>

  <div class="field">
    <label>Exercises (in order)</label>
    {#if selectedIds.length === 0}
      <p class="muted">No exercises yet. Pick one below.</p>
    {:else}
      <div class="stack">
        {#each selectedIds as id, i (i + ':' + id)}
          {@const ex = exById.get(id)}
          <div class="card row">
            <div class="grow">
              <strong>{i + 1}. {ex?.name ?? '?'}</strong>
            </div>
            <button type="button" class="ghost" on:click={() => move(i, -1)} disabled={i === 0}>↑</button>
            <button
              type="button"
              class="ghost"
              on:click={() => move(i, 1)}
              disabled={i === selectedIds.length - 1}
            >
              ↓
            </button>
            <button type="button" class="danger" on:click={() => remove(i)}>×</button>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <div class="field">
    <label>Add exercise</label>
    <select bind:value={pickerValue}>
      <option value="">— Select —</option>
      {#each allExercises as e}
        <option value={e.id}>{e.name}</option>
      {/each}
    </select>
  </div>

  <label class="toggle-row card">
    <div class="grow">
      <strong>Alternate sets</strong>
      <div class="muted" style="font-size: 0.85rem; margin-top: 0.2rem;">
        Round-robin: do one set of each exercise, then loop. Off = finish all sets of one before moving on.
      </div>
    </div>
    <input type="checkbox" bind:checked={alternateSets} class="switch" />
  </label>

  <button class="primary" type="submit" disabled={busy}>{busy ? '…' : submitLabel}</button>
  {#if error}<p class="error">{error}</p>{/if}
</form>

<style>
  .toggle-row {
    display: flex;
    align-items: center;
    gap: 1rem;
    cursor: pointer;
    user-select: none;
  }
  .switch {
    appearance: none;
    width: 48px;
    height: 28px;
    background: var(--surface-2);
    border: 1px solid var(--hairline);
    border-radius: 999px;
    position: relative;
    cursor: pointer;
    transition: background 0.15s ease, border-color 0.15s ease;
    flex-shrink: 0;
  }
  .switch::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 22px;
    height: 22px;
    background: var(--text);
    border-radius: 50%;
    transition: transform 0.15s ease, background 0.15s ease;
  }
  .switch:checked {
    background: var(--accent);
    border-color: var(--accent);
  }
  .switch:checked::after {
    transform: translateX(20px);
    background: #071a16;
  }
</style>
