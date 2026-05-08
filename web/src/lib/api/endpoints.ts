import { api } from './client';
import type {
  Exercise,
  ExerciseKind,
  Machine,
  PeriodReport,
  ProgressionSuggestion,
  Routine,
  RoutineRun,
  Session,
  SessionStatus,
  Difficulty,
  User
} from './types';

// --- Auth ---

export const auth = {
  register: (email: string, password: string, name: string) =>
    api<{ user: User }>('/v1/auth/register', { method: 'POST', body: { email, password, name } }),
  login: (email: string, password: string, rememberMe: boolean) =>
    api<{ user: User }>('/v1/auth/login', {
      method: 'POST',
      body: { email, password, rememberMe },
      skipAuthRefresh: true
    }),
  logout: () => api<{}>('/v1/auth/logout', { method: 'POST', body: {} }),
  me: () => api<{ user: User }>('/v1/auth/me', { skipAuthRefresh: false })
};

// --- Machines ---

export const machines = {
  list: () => api<{ machines?: Machine[] }>('/v1/machines'),
  get: (id: string) => api<{ machine: Machine }>(`/v1/machines/${id}`),
  create: (m: Pick<Machine, 'name' | 'location' | 'notes'>) =>
    api<{ machine: Machine }>('/v1/machines', { method: 'POST', body: m }),
  update: (id: string, m: Pick<Machine, 'name' | 'location' | 'notes'>) =>
    api<{ machine: Machine }>(`/v1/machines/${id}`, { method: 'PUT', body: { id, ...m } }),
  delete: (id: string) => api<{}>(`/v1/machines/${id}`, { method: 'DELETE' })
};

// --- Exercises ---

export interface ExerciseInput {
  name: string;
  kind: ExerciseKind;
  machineId?: string;
  goalCount?: number;
  goalDurationSeconds?: number;
  goalSets?: number;
  goalWeightLb?: number;
  instructions?: string;
}

export const exercises = {
  list: (includeArchived = false) =>
    api<{ exercises?: Exercise[] }>(`/v1/exercises${includeArchived ? '?includeArchived=true' : ''}`),
  get: (id: string) => api<{ exercise: Exercise }>(`/v1/exercises/${id}`),
  create: (e: ExerciseInput) =>
    api<{ exercise: Exercise }>('/v1/exercises', {
      method: 'POST',
      body: { ...e, instructions: e.instructions ?? '' }
    }),
  update: (id: string, e: ExerciseInput) =>
    api<{ exercise: Exercise }>(`/v1/exercises/${id}`, {
      method: 'PUT',
      body: { id, ...e, instructions: e.instructions ?? '' }
    }),
  archive: (id: string) => api<{}>(`/v1/exercises/${id}`, { method: 'DELETE' })
};

// --- Routines ---

export const routines = {
  list: (includeArchived = false) =>
    api<{ routines?: Routine[] }>(`/v1/routines${includeArchived ? '?includeArchived=true' : ''}`),
  get: (id: string) => api<{ routine: Routine }>(`/v1/routines/${id}`),
  create: (name: string, exerciseIds: string[], alternateSets: boolean) =>
    api<{ routine: Routine }>('/v1/routines', {
      method: 'POST',
      body: { name, exerciseIds, alternateSets }
    }),
  update: (id: string, name: string, exerciseIds: string[], alternateSets: boolean) =>
    api<{ routine: Routine }>(`/v1/routines/${id}`, {
      method: 'PUT',
      body: { id, name, exerciseIds, alternateSets }
    }),
  archive: (id: string) => api<{}>(`/v1/routines/${id}`, { method: 'DELETE' }),
  start: (routineId: string) =>
    api<{ run: RoutineRun }>(`/v1/routines/${routineId}:start`, { method: 'POST', body: { routineId } }),
  endRun: (runId: string) =>
    api<{ run: RoutineRun }>(`/v1/routine-runs/${runId}:end`, { method: 'POST', body: { runId } })
};

// --- Sessions ---

export interface SessionInput {
  clientId: string;
  exerciseId: string;
  routineRunId?: string;
  startedAt: string; // ISO
  endedAt?: string; // ISO
  status: SessionStatus;
  countCompleted?: number;
  countGoal?: number;
  durationSeconds?: number;
  difficulty?: Difficulty;
  notes?: string;
  setIndex?: number;
  setTotal?: number;
  weightLb?: number;
}

export interface SessionPatch {
  status?: SessionStatus;
  countCompleted?: number;
  durationSeconds?: number;
  difficulty?: Difficulty;
  notes?: string;
  weightLb?: number;
}

export const sessions = {
  log: (s: SessionInput) =>
    api<{ session: Session }>('/v1/sessions', {
      method: 'POST',
      body: { ...s, notes: s.notes ?? '' }
    }),
  list: (params: { from?: string; to?: string; exerciseId?: string; routineRunId?: string; limit?: number } = {}) => {
    const q = new URLSearchParams();
    if (params.from) q.set('from', params.from);
    if (params.to) q.set('to', params.to);
    if (params.exerciseId) q.set('exerciseId', params.exerciseId);
    if (params.routineRunId) q.set('routineRunId', params.routineRunId);
    if (params.limit !== undefined) q.set('limit', String(params.limit));
    const qs = q.toString();
    return api<{ sessions?: Session[] }>(`/v1/sessions${qs ? `?${qs}` : ''}`);
  },
  update: (id: string, patch: SessionPatch) =>
    api<{ session: Session }>(`/v1/sessions/${id}`, {
      method: 'PUT',
      body: { id, ...patch }
    }),
  remove: (id: string) => api<{}>(`/v1/sessions/${id}`, { method: 'DELETE' })
};

// --- Reports + progression ---

export const reports = {
  weekly: (week?: string) =>
    api<{ report: PeriodReport }>(`/v1/reports/weekly${week ? `?week=${encodeURIComponent(week)}` : ''}`),
  monthly: (month?: string) =>
    api<{ report: PeriodReport }>(`/v1/reports/monthly${month ? `?month=${encodeURIComponent(month)}` : ''}`),
  progression: () =>
    api<{ suggestions?: ProgressionSuggestion[] }>('/v1/progression-suggestions')
};
