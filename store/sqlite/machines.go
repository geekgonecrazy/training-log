package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/geekgonecrazy/training-log/store"
)

type machineStore struct{ db *sql.DB }

func (s *machineStore) List(ctx context.Context, userID int64) ([]*store.Machine, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, name, location, notes, created_at
		FROM machines WHERE user_id = ? ORDER BY name COLLATE NOCASE ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*store.Machine
	for rows.Next() {
		m, err := scanMachine(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *machineStore) Get(ctx context.Context, userID, id int64) (*store.Machine, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, location, notes, created_at
		FROM machines WHERE user_id = ? AND id = ?`, userID, id)
	var m store.Machine
	var created int64
	if err := row.Scan(&m.ID, &m.UserID, &m.Name, &m.Location, &m.Notes, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt = time.Unix(created, 0).UTC()
	return &m, nil
}

func (s *machineStore) Create(ctx context.Context, m *store.Machine) (int64, error) {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO machines(user_id, name, location, notes, created_at)
		VALUES(?, ?, ?, ?, ?)`,
		m.UserID, m.Name, m.Location, m.Notes, now)
	if err != nil {
		return 0, fmt.Errorf("insert machine: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	m.ID = id
	m.CreatedAt = time.Unix(now, 0).UTC()
	return id, nil
}

func (s *machineStore) Update(ctx context.Context, m *store.Machine) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE machines SET name = ?, location = ?, notes = ?
		WHERE id = ? AND user_id = ?`,
		m.Name, m.Location, m.Notes, m.ID, m.UserID)
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

func (s *machineStore) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM machines WHERE user_id = ? AND id = ?`, userID, id)
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

func scanMachine(rows *sql.Rows) (*store.Machine, error) {
	var m store.Machine
	var created int64
	if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Location, &m.Notes, &created); err != nil {
		return nil, err
	}
	m.CreatedAt = time.Unix(created, 0).UTC()
	return &m, nil
}
