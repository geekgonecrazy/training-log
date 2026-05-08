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

type userStore struct{ db *sql.DB }

func (s *userStore) Create(ctx context.Context, email, passwordHash, name string) (*store.User, error) {
	defer trackDBOp("insert", "users", "Create")()
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users(email, password_hash, name, created_at) VALUES(?, ?, ?, ?)`,
		email, passwordHash, name, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, store.ErrConflict
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &store.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		CreatedAt:    time.Unix(now, 0).UTC(),
	}, nil
}

func (s *userStore) GetByID(ctx context.Context, id int64) (*store.User, error) {
	defer trackDBOp("select", "users", "GetByID")()
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*store.User, error) {
	defer trackDBOp("select", "users", "GetByEmail")()
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = ? COLLATE NOCASE`, email)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*store.User, error) {
	var u store.User
	var createdAt int64
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	u.CreatedAt = time.Unix(createdAt, 0).UTC()
	return &u, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite surfaces UNIQUE constraint failures in the message.
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
