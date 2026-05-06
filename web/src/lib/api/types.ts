// TypeScript shapes matching the proto JSON serialization (camelCase, enums as strings).
// Hand-maintained; see models/habit/v1/*.proto for the source of truth.

export type ExerciseKind =
  | 'EXERCISE_KIND_UNSPECIFIED'
  | 'EXERCISE_KIND_TIMED'
  | 'EXERCISE_KIND_COUNTED'
  | 'EXERCISE_KIND_LOGGED'
  | 'EXERCISE_KIND_WEIGHTED';

export type SessionStatus =
  | 'SESSION_STATUS_UNSPECIFIED'
  | 'SESSION_STATUS_IN_PROGRESS'
  | 'SESSION_STATUS_COMPLETED'
  | 'SESSION_STATUS_SKIPPED'
  | 'SESSION_STATUS_FAILED';

export type Difficulty =
  | 'DIFFICULTY_UNSPECIFIED'
  | 'DIFFICULTY_EASY'
  | 'DIFFICULTY_MODERATE'
  | 'DIFFICULTY_HARD'
  | 'DIFFICULTY_VERY_HARD'
  | 'DIFFICULTY_FAILED';

export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
}

export interface Machine {
  id: string;
  name: string;
  location: string;
  notes: string;
  createdAt: string;
}

export interface Exercise {
  id: string;
  name: string;
  kind: ExerciseKind;
  machineId?: string;
  goalCount?: number;            // reps per set (COUNTED, WEIGHTED)
  goalDurationSeconds?: number;  // seconds per set (TIMED)
  goalSets?: number;             // total sets per session (default 1)
  goalWeightLb?: number;         // weight in pounds (WEIGHTED only)
  instructions: string;
  createdAt: string;
  archivedAt?: string;
}

export interface RoutineItem {
  id: string;
  routineId: string;
  exerciseId: string;
  position: number;
}

export interface Routine {
  id: string;
  name: string;
  createdAt: string;
  archivedAt?: string;
  items: RoutineItem[];
  alternateSets: boolean;
}

export interface RoutineRun {
  id: string;
  routineId: string;
  startedAt: string;
  endedAt?: string;
}

export interface Session {
  id: string;
  exerciseId: string;
  routineRunId?: string;
  startedAt: string;
  endedAt?: string;
  status: SessionStatus;
  countCompleted?: number;
  countGoal?: number;
  durationSeconds?: number;
  difficulty?: Difficulty;
  notes: string;
  clientId: string;
  setIndex?: number;
  setTotal?: number;
  weightLb?: number;
}

export interface ExerciseRollup {
  exerciseId: string;
  exerciseName: string;
  kind: ExerciseKind;
  sessionsTotal: number;
  sessionsCompleted: number;
  sessionsSkipped: number;
  sessionsFailed: number;
  totalCount: number;
  totalDurationSeconds: number;
  avgDifficulty: number;
}

export interface PeriodReport {
  period: string;
  rollups: ExerciseRollup[];
}

export interface ProgressionSuggestion {
  exerciseId: string;
  exerciseName: string;
  kind: ExerciseKind;
  currentGoalCount?: number;
  currentGoalDurationSeconds?: number;
  supportingSessions: number;
  recommendation: string;
}
