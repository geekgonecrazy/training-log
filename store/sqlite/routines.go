package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/geekgonecrazy/training-log/store"
)

type routineStore struct{ db *sql.DB }

func (s *routineStore) List(ctx context.Context, userID int64, includeArchived bool) ([]*store.Routine, error) {
	q := `SELECT id, user_id, name, created_at, archived_at, alternate_sets FROM routines WHERE user_id = ?`
	if !includeArchived {
		q += ` AND archived_at IS NULL`
	}
	q += ` ORDER BY name COLLATE NOCASE ASC`
	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*store.Routine
	for rows.Next() {
		r, err := scanRoutine(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, r := range out {
		items, err := s.loadItems(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		r.Items = items
	}
	return out, nil
}

func (s *routineStore) Get(ctx context.Context, userID, id int64) (*store.Routine, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, created_at, archived_at, alternate_sets FROM routines WHERE user_id = ? AND id = ?`,
		userID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, store.ErrNotFound
	}
	r, err := scanRoutine(rows)
	if err != nil {
		return nil, err
	}
	items, err := s.loadItems(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	r.Items = items
	return r, nil
}

func (s *routineStore) Create(ctx context.Context, userID int64, name string, exerciseIDs []int64, alternateSets bool) (*store.Routine, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now().Unix()
	res, err := tx.ExecContext(ctx,
		`INSERT INTO routines(user_id, name, created_at, alternate_sets) VALUES(?, ?, ?, ?)`,
		userID, name, now, boolToInt(alternateSets))
	if err != nil {
		return nil, fmt.Errorf("insert routine: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := replaceItemsTx(ctx, tx, userID, id, exerciseIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

func (s *routineStore) Update(ctx context.Context, userID, id int64, name string, exerciseIDs []int64, alternateSets bool) (*store.Routine, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.ExecContext(ctx,
		`UPDATE routines SET name = ?, alternate_sets = ? WHERE user_id = ? AND id = ?`,
		name, boolToInt(alternateSets), userID, id)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, store.ErrNotFound
	}
	if exerciseIDs != nil {
		if err := replaceItemsTx(ctx, tx, userID, id, exerciseIDs); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *routineStore) Archive(ctx context.Context, userID, id int64, at time.Time) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE routines SET archived_at = ? WHERE user_id = ? AND id = ? AND archived_at IS NULL`,
		at.Unix(), userID, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *routineStore) StartRun(ctx context.Context, userID, routineID int64, startedAt time.Time) (*store.RoutineRun, error) {
	// Verify ownership of the routine.
	var ownedID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM routines WHERE id = ? AND user_id = ?`, routineID, userID).Scan(&ownedID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO routine_runs(user_id, routine_id, started_at) VALUES(?, ?, ?)`,
		userID, routineID, startedAt.Unix())
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &store.RoutineRun{
		ID:        id,
		UserID:    userID,
		RoutineID: routineID,
		StartedAt: startedAt.UTC(),
	}, nil
}

func (s *routineStore) EndRun(ctx context.Context, userID, runID int64, endedAt time.Time) (*store.RoutineRun, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE routine_runs SET ended_at = ? WHERE id = ? AND user_id = ? AND ended_at IS NULL`,
		endedAt.Unix(), runID, userID)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		// Either not found or already ended; treat as not-found for simplicity.
		return nil, store.ErrNotFound
	}
	return s.GetRun(ctx, userID, runID)
}

func (s *routineStore) GetRun(ctx context.Context, userID, runID int64) (*store.RoutineRun, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, routine_id, started_at, ended_at FROM routine_runs WHERE id = ? AND user_id = ?`,
		runID, userID)
	var (
		r       store.RoutineRun
		started int64
		ended   sql.NullInt64
	)
	if err := row.Scan(&r.ID, &r.UserID, &r.RoutineID, &started, &ended); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	r.StartedAt = time.Unix(started, 0).UTC()
	if ended.Valid {
		ts := time.Unix(ended.Int64, 0).UTC()
		r.EndedAt = &ts
	}
	return &r, nil
}

func (s *routineStore) loadItems(ctx context.Context, routineID int64) ([]*store.RoutineItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, routine_id, exercise_id, position FROM routine_items WHERE routine_id = ? ORDER BY position`,
		routineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*store.RoutineItem
	for rows.Next() {
		var i store.RoutineItem
		if err := rows.Scan(&i.ID, &i.RoutineID, &i.ExerciseID, &i.Position); err != nil {
			return nil, err
		}
		out = append(out, &i)
	}
	return out, rows.Err()
}

func replaceItemsTx(ctx context.Context, tx *sql.Tx, userID, routineID int64, exerciseIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM routine_items WHERE routine_id = ?`, routineID); err != nil {
		return fmt.Errorf("clear routine_items: %w", err)
	}
	for pos, exID := range exerciseIDs {
		// Verify the exercise belongs to the user.
		var ok int64
		err := tx.QueryRowContext(ctx,
			`SELECT id FROM exercises WHERE id = ? AND user_id = ?`, exID, userID).Scan(&ok)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("exercise %d: %w", exID, store.ErrNotFound)
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO routine_items(routine_id, exercise_id, position) VALUES(?, ?, ?)`,
			routineID, exID, pos); err != nil {
			return fmt.Errorf("insert routine_item: %w", err)
		}
	}
	return nil
}

func scanRoutine(rows *sql.Rows) (*store.Routine, error) {
	var (
		r         store.Routine
		created   int64
		archived  sql.NullInt64
		alternate int64
	)
	if err := rows.Scan(&r.ID, &r.UserID, &r.Name, &created, &archived, &alternate); err != nil {
		return nil, err
	}
	r.CreatedAt = time.Unix(created, 0).UTC()
	r.AlternateSets = alternate != 0
	if archived.Valid {
		ts := time.Unix(archived.Int64, 0).UTC()
		r.ArchivedAt = &ts
	}
	return &r, nil
}
