// Package progression decides whether a user is "ready to bump" the goal
// of a goal-based exercise (timed or counted) based on recent session history.
//
// Heuristic: if the most recent N completed sessions all hit the goal AND the
// difficulty was Easy or Moderate, surface a suggestion. The suggestion is a
// recommendation only — the user manually applies it.
package progression

import (
	"context"
	"fmt"

	habitv1 "github.com/geekgonecrazy/training-log/models/habit/v1"
	"github.com/geekgonecrazy/training-log/store"
)

// DefaultWindow is the number of recent completed sessions considered.
const DefaultWindow = 3

type Suggester struct {
	Store  store.Store
	Window int
}

func New(s store.Store) *Suggester {
	return &Suggester{Store: s, Window: DefaultWindow}
}

func (s *Suggester) ListSuggestions(ctx context.Context, userID int64) ([]*habitv1.ProgressionSuggestion, error) {
	exs, err := s.Store.Exercises().List(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	var out []*habitv1.ProgressionSuggestion
	for _, e := range exs {
		switch e.Kind {
		case int32(habitv1.ExerciseKind_EXERCISE_KIND_COUNTED),
			int32(habitv1.ExerciseKind_EXERCISE_KIND_TIMED):
			// only goal-based kinds get suggestions
		default:
			continue
		}

		recent, err := s.Store.Sessions().RecentCompletedForExercise(ctx, userID, e.ID, s.Window)
		if err != nil {
			return nil, err
		}
		if len(recent) < s.Window {
			continue
		}

		ok := true
		for _, sn := range recent {
			if sn.Difficulty == nil {
				ok = false
				break
			}
			d := *sn.Difficulty
			if d != int32(habitv1.Difficulty_DIFFICULTY_EASY) &&
				d != int32(habitv1.Difficulty_DIFFICULTY_MODERATE) {
				ok = false
				break
			}
			if !metGoal(e, sn) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		sug := &habitv1.ProgressionSuggestion{
			ExerciseId:         e.ID,
			ExerciseName:       e.Name,
			Kind:               habitv1.ExerciseKind(e.Kind),
			SupportingSessions: int32(len(recent)),
		}
		switch habitv1.ExerciseKind(e.Kind) {
		case habitv1.ExerciseKind_EXERCISE_KIND_COUNTED:
			if e.GoalCount != nil {
				cur := *e.GoalCount
				sug.CurrentGoalCount = &cur
				bump := bumpCount(cur)
				sug.Recommendation = fmt.Sprintf("+%d reps (try %d)", bump, cur+bump)
			} else {
				sug.Recommendation = "Set a goal count and try bumping it next session"
			}
		case habitv1.ExerciseKind_EXERCISE_KIND_TIMED:
			if e.GoalDurationSeconds != nil {
				cur := *e.GoalDurationSeconds
				sug.CurrentGoalDurationSeconds = &cur
				bump := bumpDuration(cur)
				sug.Recommendation = fmt.Sprintf("+%ds (try %ds)", bump, cur+bump)
			} else {
				sug.Recommendation = "Set a goal duration and try bumping it next session"
			}
		}
		out = append(out, sug)
	}
	return out, nil
}

func metGoal(e *store.Exercise, sn *store.Session) bool {
	switch habitv1.ExerciseKind(e.Kind) {
	case habitv1.ExerciseKind_EXERCISE_KIND_COUNTED:
		if e.GoalCount == nil || sn.CountCompleted == nil {
			return false
		}
		return *sn.CountCompleted >= *e.GoalCount
	case habitv1.ExerciseKind_EXERCISE_KIND_TIMED:
		if e.GoalDurationSeconds == nil || sn.DurationSeconds == nil {
			return false
		}
		return *sn.DurationSeconds >= *e.GoalDurationSeconds
	}
	return false
}

func bumpCount(cur int32) int32 {
	switch {
	case cur < 10:
		return 1
	case cur < 30:
		return 2
	case cur < 60:
		return 5
	default:
		return 10
	}
}

func bumpDuration(cur int32) int32 {
	switch {
	case cur < 30:
		return 5
	case cur < 120:
		return 10
	case cur < 300:
		return 30
	default:
		return 60
	}
}
