import type { Exercise } from '$lib/api/types';

/** Concise summary like "3 × 20 reps", "3 × 30s", "3 × 8 @ 135 lb", or "Log only". */
export function exerciseGoalSummary(e: Exercise): string {
  const sets = e.goalSets && e.goalSets > 0 ? e.goalSets : 1;
  const setsPrefix = sets > 1 ? `${sets} × ` : '';

  switch (e.kind) {
    case 'EXERCISE_KIND_TIMED':
      return e.goalDurationSeconds ? `${setsPrefix}${e.goalDurationSeconds}s` : 'Timed';
    case 'EXERCISE_KIND_COUNTED':
      return e.goalCount ? `${setsPrefix}${e.goalCount} reps` : 'Reps';
    case 'EXERCISE_KIND_WEIGHTED': {
      const reps = e.goalCount ? `${e.goalCount}` : '?';
      const weight = e.goalWeightLb ? ` @ ${e.goalWeightLb} lb` : '';
      return `${setsPrefix}${reps}${weight}`;
    }
    case 'EXERCISE_KIND_LOGGED':
      return 'Log only';
    default:
      return '';
  }
}

export function kindIcon(k: string) {
  switch (k) {
    case 'EXERCISE_KIND_TIMED': return 'timer';
    case 'EXERCISE_KIND_COUNTED': return 'dumbbell';
    case 'EXERCISE_KIND_LOGGED': return 'check';
    case 'EXERCISE_KIND_WEIGHTED': return 'dumbbell-fill';
    default: return 'dumbbell';
  }
}
