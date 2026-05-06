<script lang="ts">
  // Big-number rep entry: tap +/− to step, type to override, or "Completed all"
  // to fill in the goal. Default value is 0 so you only build up what you actually did.

  export let value: number;
  export let goal: number | null = null;

  function step(delta: number) {
    const next = (value ?? 0) + delta;
    value = Math.max(0, next);
  }
  function completedAll() {
    if (goal != null) value = goal;
  }
  function onFocus(e: FocusEvent) {
    const t = e.currentTarget as HTMLInputElement;
    t.select();
  }
</script>

<div class="reps-stepper">
  <button type="button" class="step" on:click={() => step(-1)} aria-label="Decrement">−</button>
  <input
    type="number"
    inputmode="numeric"
    min="0"
    class="reps-input tabular"
    bind:value
    on:focus={onFocus}
  />
  <button type="button" class="step" on:click={() => step(1)} aria-label="Increment">+</button>
</div>

{#if goal != null && goal > 0}
  <button type="button" class="completed-all" on:click={completedAll}>
    Completed all {goal}
  </button>
{/if}

<style>
  .reps-stepper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    margin: 0.5rem 0;
  }
  .step {
    width: 64px;
    height: 64px;
    flex-shrink: 0;
    font-size: 2rem;
    font-weight: 700;
    border-radius: 50%;
    background: var(--surface-2);
    border: 1px solid var(--hairline);
    color: var(--text);
    line-height: 1;
    padding: 0;
  }
  .step:active { transform: scale(0.92); }

  .reps-input {
    width: 0;
    flex: 1;
    max-width: 220px;
    background: transparent;
    border: none;
    text-align: center;
    font-size: 5.5rem;
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.03em;
    padding: 0;
    color: var(--text);
    -moz-appearance: textfield;
  }
  .reps-input::-webkit-outer-spin-button,
  .reps-input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
  .reps-input:focus {
    outline: none;
    background: transparent;
  }

  .completed-all {
    width: 100%;
    background: rgba(0, 210, 168, 0.08);
    border: 1px solid rgba(0, 210, 168, 0.3);
    color: var(--accent);
    font-weight: 600;
    margin-top: 0.5rem;
  }
  .completed-all:hover { background: rgba(0, 210, 168, 0.14); }
</style>
