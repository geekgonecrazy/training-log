package controllers

import (
	"time"

	habitv1 "github.com/geekgonecrazy/training-log/models/habit/v1"
	"github.com/geekgonecrazy/training-log/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helpers to convert between store domain types and proto messages.

func tsFrom(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func tsFromPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func userToProto(u *store.User) *habitv1.User {
	return &habitv1.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: tsFrom(u.CreatedAt),
	}
}

func machineToProto(m *store.Machine) *habitv1.Machine {
	return &habitv1.Machine{
		Id:        m.ID,
		Name:      m.Name,
		Location:  m.Location,
		Notes:     m.Notes,
		CreatedAt: tsFrom(m.CreatedAt),
	}
}

func exerciseToProto(e *store.Exercise) *habitv1.Exercise {
	out := &habitv1.Exercise{
		Id:           e.ID,
		Name:         e.Name,
		Kind:         habitv1.ExerciseKind(e.Kind),
		Instructions: e.Instructions,
		CreatedAt:    tsFrom(e.CreatedAt),
		ArchivedAt:   tsFromPtr(e.ArchivedAt),
	}
	if e.MachineID != nil {
		v := *e.MachineID
		out.MachineId = &v
	}
	if e.GoalCount != nil {
		v := *e.GoalCount
		out.GoalCount = &v
	}
	if e.GoalDurationSeconds != nil {
		v := *e.GoalDurationSeconds
		out.GoalDurationSeconds = &v
	}
	if e.GoalSets != nil {
		v := *e.GoalSets
		out.GoalSets = &v
	}
	if e.GoalWeightLb != nil {
		v := *e.GoalWeightLb
		out.GoalWeightLb = &v
	}
	return out
}

func routineToProto(r *store.Routine) *habitv1.Routine {
	items := make([]*habitv1.RoutineItem, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, &habitv1.RoutineItem{
			Id:         it.ID,
			RoutineId:  it.RoutineID,
			ExerciseId: it.ExerciseID,
			Position:   it.Position,
		})
	}
	return &habitv1.Routine{
		Id:            r.ID,
		Name:          r.Name,
		CreatedAt:     tsFrom(r.CreatedAt),
		ArchivedAt:    tsFromPtr(r.ArchivedAt),
		Items:         items,
		AlternateSets: r.AlternateSets,
	}
}

func routineRunToProto(r *store.RoutineRun) *habitv1.RoutineRun {
	return &habitv1.RoutineRun{
		Id:        r.ID,
		RoutineId: r.RoutineID,
		StartedAt: tsFrom(r.StartedAt),
		EndedAt:   tsFromPtr(r.EndedAt),
	}
}

func sessionToProto(s *store.Session) *habitv1.Session {
	out := &habitv1.Session{
		Id:         s.ID,
		ExerciseId: s.ExerciseID,
		StartedAt:  tsFrom(s.StartedAt),
		EndedAt:    tsFromPtr(s.EndedAt),
		Status:     habitv1.SessionStatus(s.Status),
		Notes:      s.Notes,
		ClientId:   s.ClientID,
	}
	if s.RoutineRunID != nil {
		v := *s.RoutineRunID
		out.RoutineRunId = &v
	}
	if s.CountCompleted != nil {
		v := *s.CountCompleted
		out.CountCompleted = &v
	}
	if s.CountGoal != nil {
		v := *s.CountGoal
		out.CountGoal = &v
	}
	if s.DurationSeconds != nil {
		v := *s.DurationSeconds
		out.DurationSeconds = &v
	}
	if s.Difficulty != nil {
		d := habitv1.Difficulty(*s.Difficulty)
		out.Difficulty = &d
	}
	if s.SetIndex != nil {
		v := *s.SetIndex
		out.SetIndex = &v
	}
	if s.SetTotal != nil {
		v := *s.SetTotal
		out.SetTotal = &v
	}
	if s.WeightLb != nil {
		v := *s.WeightLb
		out.WeightLb = &v
	}
	return out
}
