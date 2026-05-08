package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/geekgonecrazy/training-log/store"
)

type sessionStore struct{ db *sql.DB }

func (s *sessionStore) Log(ctx context.Context, sn *store.Session) (*store.Session, error) {
	defer trackDBOp("insert", "sessions", "Log")()
	// Idempotent on (user_id, client_id).
	existing, err := s.getByClientID(ctx, sn.UserID, sn.ClientID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions(user_id, exercise_id, routine_run_id, started_at, ended_at, status,
		                    count_completed, count_goal, duration_seconds, difficulty, notes, client_id,
		                    set_index, set_total, weight_lb)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sn.UserID, sn.ExerciseID, nullableInt64(sn.RoutineRunID),
		sn.StartedAt.Unix(), nullableUnix(sn.EndedAt),
		sn.Status, nullableInt32(sn.CountCompleted), nullableInt32(sn.CountGoal),
		nullableInt32(sn.DurationSeconds), nullableInt32(sn.Difficulty),
		sn.Notes, sn.ClientID,
		nullableInt32(sn.SetIndex), nullableInt32(sn.SetTotal), nullableFloat64(sn.WeightLb))
	if err != nil {
		// Concurrent insert may race; treat unique violation as idempotent return.
		if isUniqueViolation(err) {
			return s.getByClientID(ctx, sn.UserID, sn.ClientID)
		}
		return nil, fmt.Errorf("insert session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	sn.ID = id
	return sn, nil
}

func (s *sessionStore) List(ctx context.Context, userID int64, f store.SessionFilter) ([]*store.Session, error) {
	defer trackDBOp("select", "sessions", "List")()
	var (
		clauses []string
		args    []any
	)
	clauses = append(clauses, "user_id = ?")
	args = append(args, userID)
	if f.From != nil {
		clauses = append(clauses, "started_at >= ?")
		args = append(args, f.From.Unix())
	}
	if f.To != nil {
		clauses = append(clauses, "started_at < ?")
		args = append(args, f.To.Unix())
	}
	if f.ExerciseID != nil {
		clauses = append(clauses, "exercise_id = ?")
		args = append(args, *f.ExerciseID)
	}
	if f.RoutineRunID != nil {
		clauses = append(clauses, "routine_run_id = ?")
		args = append(args, *f.RoutineRunID)
	}

	q := `SELECT ` + sessionCols + ` FROM sessions WHERE ` + strings.Join(clauses, " AND ") +
		` ORDER BY started_at DESC`
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", f.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*store.Session
	for rows.Next() {
		sn, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sn)
	}
	return out, rows.Err()
}

func (s *sessionStore) RecentCompletedForExercise(ctx context.Context, userID, exerciseID int64, limit int) ([]*store.Session, error) {
	defer trackDBOp("select", "sessions", "RecentCompletedForExercise")()
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+sessionCols+` FROM sessions
		 WHERE user_id = ? AND exercise_id = ? AND status = 2
		 ORDER BY started_at DESC LIMIT ?`,
		userID, exerciseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*store.Session
	for rows.Next() {
		sn, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sn)
	}
	return out, rows.Err()
}

func (s *sessionStore) getByClientID(ctx context.Context, userID int64, clientID string) (*store.Session, error) {
	defer trackDBOp("select", "sessions", "getByClientID")()
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+sessionCols+` FROM sessions WHERE user_id = ? AND client_id = ?`,
		userID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, store.ErrNotFound
	}
	return scanSession(rows)
}

const sessionCols = `id, user_id, exercise_id, routine_run_id, started_at, ended_at, status,
	count_completed, count_goal, duration_seconds, difficulty, notes, client_id,
	set_index, set_total, weight_lb`

func scanSession(rows *sql.Rows) (*store.Session, error) {
	var sn store.Session
	var (
		runID                                               sql.NullInt64
		started                                             int64
		ended                                               sql.NullInt64
		countCompleted, countGoal, durationSecs, difficulty sql.NullInt64
		setIndex, setTotal                                  sql.NullInt64
		weightLb                                            sql.NullFloat64
	)
	if err := rows.Scan(&sn.ID, &sn.UserID, &sn.ExerciseID, &runID, &started, &ended, &sn.Status,
		&countCompleted, &countGoal, &durationSecs, &difficulty, &sn.Notes, &sn.ClientID,
		&setIndex, &setTotal, &weightLb); err != nil {
		return nil, err
	}
	if setIndex.Valid {
		v := int32(setIndex.Int64)
		sn.SetIndex = &v
	}
	if setTotal.Valid {
		v := int32(setTotal.Int64)
		sn.SetTotal = &v
	}
	if weightLb.Valid {
		v := weightLb.Float64
		sn.WeightLb = &v
	}
	if runID.Valid {
		v := runID.Int64
		sn.RoutineRunID = &v
	}
	sn.StartedAt = time.Unix(started, 0).UTC()
	if ended.Valid {
		ts := time.Unix(ended.Int64, 0).UTC()
		sn.EndedAt = &ts
	}
	if countCompleted.Valid {
		v := int32(countCompleted.Int64)
		sn.CountCompleted = &v
	}
	if countGoal.Valid {
		v := int32(countGoal.Int64)
		sn.CountGoal = &v
	}
	if durationSecs.Valid {
		v := int32(durationSecs.Int64)
		sn.DurationSeconds = &v
	}
	if difficulty.Valid {
		v := int32(difficulty.Int64)
		sn.Difficulty = &v
	}
	return &sn, nil
}

func nullableUnix(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Unix()
}
