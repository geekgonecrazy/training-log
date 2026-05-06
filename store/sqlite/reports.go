package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/geekgonecrazy/training-log/store"
)

type reportStore struct{ db *sql.DB }

// Rollup returns per-exercise aggregates for sessions whose started_at falls in [from, to).
//
// Status enum mapping (mirrors habit.v1.SessionStatus):
//
//	1 = IN_PROGRESS, 2 = COMPLETED, 3 = SKIPPED, 4 = FAILED
//
// AvgDifficulty is computed only over rows that have a non-null difficulty AND status = COMPLETED;
// 0 if no qualifying rows.
func (s *reportStore) Rollup(ctx context.Context, userID int64, from, to time.Time) ([]*store.ExerciseRollup, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
		  e.id,
		  e.name,
		  e.kind,
		  COUNT(s.id) AS sessions_total,
		  COALESCE(SUM(CASE WHEN s.status = 2 THEN 1 ELSE 0 END), 0) AS sessions_completed,
		  COALESCE(SUM(CASE WHEN s.status = 3 THEN 1 ELSE 0 END), 0) AS sessions_skipped,
		  COALESCE(SUM(CASE WHEN s.status = 4 THEN 1 ELSE 0 END), 0) AS sessions_failed,
		  COALESCE(SUM(CASE WHEN s.status = 2 THEN s.count_completed ELSE 0 END), 0) AS total_count,
		  COALESCE(SUM(CASE WHEN s.status = 2 THEN s.duration_seconds ELSE 0 END), 0) AS total_duration_seconds,
		  COALESCE(AVG(CASE WHEN s.status = 2 AND s.difficulty IS NOT NULL THEN s.difficulty END), 0) AS avg_difficulty
		FROM exercises e
		LEFT JOIN sessions s
		  ON s.exercise_id = e.id
		 AND s.user_id = e.user_id
		 AND s.started_at >= ? AND s.started_at < ?
		WHERE e.user_id = ?
		GROUP BY e.id
		HAVING sessions_total > 0
		ORDER BY e.name COLLATE NOCASE ASC`,
		from.Unix(), to.Unix(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*store.ExerciseRollup
	for rows.Next() {
		var r store.ExerciseRollup
		var avgDiff sql.NullFloat64
		if err := rows.Scan(&r.ExerciseID, &r.ExerciseName, &r.Kind,
			&r.SessionsTotal, &r.SessionsCompleted, &r.SessionsSkipped, &r.SessionsFailed,
			&r.TotalCount, &r.TotalDurationSeconds, &avgDiff); err != nil {
			return nil, err
		}
		if avgDiff.Valid {
			r.AvgDifficulty = avgDiff.Float64
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}
