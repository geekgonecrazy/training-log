-- 0001_init.sql -- baseline schema

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    email           TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash   TEXT NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    created_at      INTEGER NOT NULL  -- unix seconds
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,         -- sha256 of opaque token
    family_id       TEXT NOT NULL,                -- shared across rotation chain; reuse → revoke whole family
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL,
    revoked_at      INTEGER,                      -- null = active
    user_agent      TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family ON refresh_tokens(family_id);

CREATE TABLE IF NOT EXISTS machines (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    location        TEXT NOT NULL DEFAULT '',
    notes           TEXT NOT NULL DEFAULT '',
    created_at      INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_machines_user ON machines(user_id);

CREATE TABLE IF NOT EXISTS exercises (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id                 INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                    TEXT NOT NULL,
    kind                    INTEGER NOT NULL,                -- ExerciseKind enum value
    machine_id              INTEGER REFERENCES machines(id) ON DELETE SET NULL,
    goal_count              INTEGER,
    goal_duration_seconds   INTEGER,
    instructions            TEXT NOT NULL DEFAULT '',
    created_at              INTEGER NOT NULL,
    archived_at             INTEGER
);
CREATE INDEX IF NOT EXISTS idx_exercises_user ON exercises(user_id);

CREATE TABLE IF NOT EXISTS routines (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    created_at      INTEGER NOT NULL,
    archived_at     INTEGER
);
CREATE INDEX IF NOT EXISTS idx_routines_user ON routines(user_id);

CREATE TABLE IF NOT EXISTS routine_items (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    routine_id      INTEGER NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    exercise_id     INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    position        INTEGER NOT NULL,
    UNIQUE(routine_id, position)
);
CREATE INDEX IF NOT EXISTS idx_routine_items_routine ON routine_items(routine_id, position);

CREATE TABLE IF NOT EXISTS routine_runs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    routine_id      INTEGER NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    started_at      INTEGER NOT NULL,
    ended_at        INTEGER
);
CREATE INDEX IF NOT EXISTS idx_routine_runs_user ON routine_runs(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS sessions (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id                 INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id             INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    routine_run_id          INTEGER REFERENCES routine_runs(id) ON DELETE SET NULL,
    started_at              INTEGER NOT NULL,
    ended_at                INTEGER,
    status                  INTEGER NOT NULL,    -- SessionStatus enum value
    count_completed         INTEGER,
    count_goal              INTEGER,
    duration_seconds        INTEGER,
    difficulty              INTEGER,             -- Difficulty enum value (1..5)
    notes                   TEXT NOT NULL DEFAULT '',
    client_id               TEXT NOT NULL,
    UNIQUE(user_id, client_id)
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_started ON sessions(user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_exercise_started ON sessions(exercise_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_routine_run ON sessions(routine_run_id);
