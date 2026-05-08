package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/geekgonecrazy/training-log/store"
)

type refreshTokenStore struct{ db *sql.DB }

func (s *refreshTokenStore) Create(ctx context.Context, t *store.RefreshToken) (int64, error) {
	defer trackDBOp("insert", "refresh_tokens", "Create")()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens(user_id, token_hash, family_id, expires_at, created_at, user_agent)
		VALUES(?, ?, ?, ?, ?, ?)`,
		t.UserID, t.TokenHash, t.FamilyID, t.ExpiresAt.Unix(), t.CreatedAt.Unix(), t.UserAgent)
	if err != nil {
		return 0, fmt.Errorf("insert refresh_token: %w", err)
	}
	return res.LastInsertId()
}

func (s *refreshTokenStore) GetByHash(ctx context.Context, tokenHash string) (*store.RefreshToken, error) {
	defer trackDBOp("select", "refresh_tokens", "GetByHash")()
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, family_id, expires_at, created_at, revoked_at, user_agent
		FROM refresh_tokens WHERE token_hash = ?`, tokenHash)
	var (
		t              store.RefreshToken
		expires, made  int64
		revoked        sql.NullInt64
	)
	if err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.FamilyID, &expires, &made, &revoked, &t.UserAgent); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	t.ExpiresAt = time.Unix(expires, 0).UTC()
	t.CreatedAt = time.Unix(made, 0).UTC()
	if revoked.Valid {
		ts := time.Unix(revoked.Int64, 0).UTC()
		t.RevokedAt = &ts
	}
	return &t, nil
}

func (s *refreshTokenStore) Revoke(ctx context.Context, id int64, at time.Time) error {
	defer trackDBOp("update", "refresh_tokens", "Revoke")()
	_, err := s.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
		at.Unix(), id)
	return err
}

func (s *refreshTokenStore) RevokeFamily(ctx context.Context, familyID string, at time.Time) error {
	defer trackDBOp("update", "refresh_tokens", "RevokeFamily")()
	_, err := s.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = ? WHERE family_id = ? AND revoked_at IS NULL`,
		at.Unix(), familyID)
	return err
}

func (s *refreshTokenStore) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	defer trackDBOp("delete", "refresh_tokens", "DeleteExpired")()
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < ?`, before.Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
