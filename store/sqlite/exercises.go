package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/geekgonecrazy/training-log/store"
)

type exerciseStore struct{ db *sql.DB }

func (s *exerciseStore) List(ctx context.Context, userID int64, includeArchived bool) ([]*store.Exercise, error) {
	defer trackDBOp("select", "exercises", "List")()
	q := `
		SELECT id, user_id, name, kind, machine_id, goal_count, goal_duration_seconds,
		       instructions, created_at, archived_at, goal_sets, goal_weight_lb
		FROM exercises WHERE user_id = ?`
	if !includeArchived {
		q += ` AND archived_at IS NULL`
	}
	q += ` ORDER BY name COLLATE NOCASE ASC`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*store.Exercise
	for rows.Next() {
		e, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *exerciseStore) Get(ctx context.Context, userID, id int64) (*store.Exercise, error) {
	defer trackDBOp("select", "exercises", "Get")()
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, name, kind, machine_id, goal_count, goal_duration_seconds,
		       instructions, created_at, archived_at, goal_sets, goal_weight_lb
		FROM exercises WHERE user_id = ? AND id = ?`, userID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, store.ErrNotFound
	}
	return scanExercise(rows)
}

func (s *exerciseStore) Create(ctx context.Context, e *store.Exercise) (int64, error) {
	defer trackDBOp("insert", "exercises", "Create")()
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO exercises(user_id, name, kind, machine_id, goal_count, goal_duration_seconds,
		                     instructions, created_at, goal_sets, goal_weight_lb)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.UserID, e.Name, e.Kind, nullableInt64(e.MachineID), nullableInt32(e.GoalCount),
		nullableInt32(e.GoalDurationSeconds), e.Instructions, now,
		nullableInt32(e.GoalSets), nullableFloat64(e.GoalWeightLb))
	if err != nil {
		return 0, fmt.Errorf("insert exercise: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	e.ID = id
	e.CreatedAt = time.Unix(now, 0).UTC()
	return id, nil
}

func (s *exerciseStore) Update(ctx context.Context, e *store.Exercise) error {
	defer trackDBOp("update", "exercises", "Update")()
	res, err := s.db.ExecContext(ctx, `
		UPDATE exercises
		SET name = ?, kind = ?, machine_id = ?, goal_count = ?, goal_duration_seconds = ?,
		    instructions = ?, goal_sets = ?, goal_weight_lb = ?
		WHERE id = ? AND user_id = ?`,
		e.Name, e.Kind, nullableInt64(e.MachineID), nullableInt32(e.GoalCount),
		nullableInt32(e.GoalDurationSeconds), e.Instructions,
		nullableInt32(e.GoalSets), nullableFloat64(e.GoalWeightLb),
		e.ID, e.UserID)
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

func (s *exerciseStore) Archive(ctx context.Context, userID, id int64, at time.Time) error {
	defer trackDBOp("update", "exercises", "Archive")()
	res, err := s.db.ExecContext(ctx,
		`UPDATE exercises SET archived_at = ? WHERE user_id = ? AND id = ? AND archived_at IS NULL`,
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

func scanExercise(rows *sql.Rows) (*store.Exercise, error) {
	var e store.Exercise
	var (
		machineID, goalCount, goalDur sql.NullInt64
		created                       int64
		archived                      sql.NullInt64
		goalSets                      sql.NullInt64
		goalWeight                    sql.NullFloat64
	)
	if err := rows.Scan(&e.ID, &e.UserID, &e.Name, &e.Kind, &machineID, &goalCount, &goalDur,
		&e.Instructions, &created, &archived, &goalSets, &goalWeight); err != nil {
		return nil, err
	}
	if machineID.Valid {
		v := machineID.Int64
		e.MachineID = &v
	}
	if goalCount.Valid {
		v := int32(goalCount.Int64)
		e.GoalCount = &v
	}
	if goalDur.Valid {
		v := int32(goalDur.Int64)
		e.GoalDurationSeconds = &v
	}
	if goalSets.Valid {
		v := int32(goalSets.Int64)
		e.GoalSets = &v
	}
	if goalWeight.Valid {
		v := goalWeight.Float64
		e.GoalWeightLb = &v
	}
	e.CreatedAt = time.Unix(created, 0).UTC()
	if archived.Valid {
		ts := time.Unix(archived.Int64, 0).UTC()
		e.ArchivedAt = &ts
	}
	return &e, nil
}

func nullableInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}
func nullableInt32(p *int32) any {
	if p == nil {
		return nil
	}
	return *p
}
func nullableFloat64(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

