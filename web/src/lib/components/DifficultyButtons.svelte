<script lang="ts">
  import type { Difficulty } from '$lib/api/types';

  export let onChoose: (d: Difficulty) => void;

  const options: { value: Difficulty; label: string; color: string }[] = [
    { value: 'DIFFICULTY_EASY', label: 'Easy', color: '#22c55e' },
    { value: 'DIFFICULTY_MODERATE', label: 'Moderate', color: '#84cc16' },
    { value: 'DIFFICULTY_HARD', label: 'Hard', color: '#f59e0b' },
    { value: 'DIFFICULTY_VERY_HARD', label: 'Very hard', color: '#f97316' },
    { value: 'DIFFICULTY_FAILED', label: 'Failed', color: '#ef4444' }
  ];
</script>

<div class="difficulty-grid">
  {#each options as o}
    <button on:click={() => onChoose(o.value)} style="border-color: {o.color}; color: {o.color};">
      {o.label}
    </button>
  {/each}
</div>

<style>
  .difficulty-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
  }
  .difficulty-grid button {
    background: transparent;
    font-weight: 600;
    padding: 1.25rem 0.5rem;
  }
  .difficulty-grid button:nth-child(5) {
    grid-column: 1 / -1;
  }
</style>
