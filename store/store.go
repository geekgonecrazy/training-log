// Package store defines the storage interface used by controllers.
//
// The shape is store.Users().Get(ctx, id), one sub-store per domain. The
// concrete SQLite implementation lives in store/sqlite. All methods take
// context and explicit user_id where ownership matters; the controllers
// pull user_id from the request context (set by auth middleware).
package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrForbidden = errors.New("forbidden")
)

// Store is the top-level handle. Sub-stores share the same DB connection.
type Store interface {
	Users() UserStore
	RefreshTokens() RefreshTokenStore
	Machines() MachineStore
	Exercises() ExerciseStore
	Routines() RoutineStore
	Sessions() SessionStore
	Reports() ReportStore
	Close() error
}

// --- Users ---

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
}

type UserStore interface {
	Create(ctx context.Context, email, passwordHash, name string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// --- Refresh tokens ---

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	FamilyID  string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
	UserAgent string
}

type RefreshTokenStore interface {
	Create(ctx context.Context, t *RefreshToken) (int64, error)
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id int64, at time.Time) error
	RevokeFamily(ctx context.Context, familyID string, at time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

// --- Machines ---

type Machine struct {
	ID        int64
	UserID    int64
	Name      string
	Location  string
	Notes     string
	CreatedAt time.Time
}

type MachineStore interface {
	List(ctx context.Context, userID int64) ([]*Machine, error)
	Get(ctx context.Context, userID, id int64) (*Machine, error)
	Create(ctx context.Context, m *Machine) (int64, error)
	Update(ctx context.Context, m *Machine) error
	Delete(ctx context.Context, userID, id int64) error
}

// --- Exercises ---

type Exercise struct {
	ID                  int64
	UserID              int64
	Name                string
	Kind                int32 // mirrors habitv1.ExerciseKind
	MachineID           *int64
	GoalCount           *int32
	GoalDurationSeconds *int32
	Instructions        string
	CreatedAt           time.Time
	ArchivedAt          *time.Time
	GoalSets            *int32
	GoalWeightLb        *float64
}

type ExerciseStore interface {
	List(ctx context.Context, userID int64, includeArchived bool) ([]*Exercise, error)
	Get(ctx context.Context, userID, id int64) (*Exercise, error)
	Create(ctx context.Context, e *Exercise) (int64, error)
	Update(ctx context.Context, e *Exercise) error
	Archive(ctx context.Context, userID, id int64, at time.Time) error
}

// --- Routines ---

type Routine struct {
	ID            int64
	UserID        int64
	Name          string
	CreatedAt     time.Time
	ArchivedAt    *time.Time
	Items         []*RoutineItem
	AlternateSets bool
}

type RoutineItem struct {
	ID         int64
	RoutineID  int64
	ExerciseID int64
	Position   int32
}

type RoutineRun struct {
	ID        int64
	UserID    int64
	RoutineID int64
	StartedAt time.Time
	EndedAt   *time.Time
}

type RoutineStore interface {
	List(ctx context.Context, userID int64, includeArchived bool) ([]*Routine, error)
	Get(ctx context.Context, userID, id int64) (*Routine, error)
	// Create inserts a routine and replaces its items with exerciseIDs in order.
	Create(ctx context.Context, userID int64, name string, exerciseIDs []int64, alternateSets bool) (*Routine, error)
	// Update changes name + alternateSets, and (if exerciseIDs non-nil) replaces the item set.
	Update(ctx context.Context, userID, id int64, name string, exerciseIDs []int64, alternateSets bool) (*Routine, error)
	Archive(ctx context.Context, userID, id int64, at time.Time) error

	StartRun(ctx context.Context, userID, routineID int64, startedAt time.Time) (*RoutineRun, error)
	EndRun(ctx context.Context, userID, runID int64, endedAt time.Time) (*RoutineRun, error)
	GetRun(ctx context.Context, userID, runID int64) (*RoutineRun, error)
}

// --- Sessions ---

type Session struct {
	ID              int64
	UserID          int64
	ExerciseID      int64
	RoutineRunID    *int64
	StartedAt       time.Time
	EndedAt         *time.Time
	Status          int32 // mirrors habitv1.SessionStatus
	CountCompleted  *int32
	CountGoal       *int32
	DurationSeconds *int32
	Difficulty      *int32 // mirrors habitv1.Difficulty
	Notes           string
	ClientID        string
	SetIndex        *int32
	SetTotal        *int32
	WeightLb        *float64
}

type SessionFilter struct {
	From         *time.Time
	To           *time.Time
	ExerciseID   *int64
	RoutineRunID *int64
	Limit        int32
}

type SessionStore interface {
	// Log writes a session, idempotent on (user_id, client_id). If a session
	// with the same client_id already exists for the user, returns the existing row.
	Log(ctx context.Context, s *Session) (*Session, error)
	List(ctx context.Context, userID int64, f SessionFilter) ([]*Session, error)
	// RecentForExercise returns the most recent N sessions for an exercise,
	// newest-first, filtered to completed status. Used by progression.
	RecentCompletedForExercise(ctx context.Context, userID, exerciseID int64, limit int) ([]*Session, error)
}

// --- Reports ---

type ExerciseRollup struct {
	ExerciseID           int64
	ExerciseName         string
	Kind                 int32
	SessionsTotal        int32
	SessionsCompleted    int32
	SessionsSkipped      int32
	SessionsFailed       int32
	TotalCount           int32
	TotalDurationSeconds int32
	AvgDifficulty        float64
}

type ReportStore interface {
	Rollup(ctx context.Context, userID int64, from, to time.Time) ([]*ExerciseRollup, error)
}
