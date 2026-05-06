package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/geekgonecrazy/training-log/core/auth"
	"github.com/geekgonecrazy/training-log/core/progression"
	habitv1 "github.com/geekgonecrazy/training-log/models/habit/v1"
	"github.com/geekgonecrazy/training-log/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HabitController implements habitv1.HabitServiceServer.
type HabitController struct {
	habitv1.UnimplementedHabitServiceServer
	Store     store.Store
	Suggester *progression.Suggester
	Now       func() time.Time
}

func NewHabitController(s store.Store) *HabitController {
	return &HabitController{Store: s, Suggester: progression.New(s), Now: time.Now}
}

// --- Machines ---

func (c *HabitController) ListMachines(ctx context.Context, _ *habitv1.ListMachinesRequest) (*habitv1.ListMachinesResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	ms, err := c.Store.Machines().List(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list machines: %v", err)
	}
	out := make([]*habitv1.Machine, 0, len(ms))
	for _, m := range ms {
		out = append(out, machineToProto(m))
	}
	return &habitv1.ListMachinesResponse{Machines: out}, nil
}

func (c *HabitController) GetMachine(ctx context.Context, req *habitv1.GetMachineRequest) (*habitv1.GetMachineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	m, err := c.Store.Machines().Get(ctx, uid, req.GetId())
	if err != nil {
		return nil, mapStoreErr(err, "machine")
	}
	return &habitv1.GetMachineResponse{Machine: machineToProto(m)}, nil
}

func (c *HabitController) CreateMachine(ctx context.Context, req *habitv1.CreateMachineRequest) (*habitv1.CreateMachineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	m := &store.Machine{
		UserID:   uid,
		Name:     req.GetName(),
		Location: req.GetLocation(),
		Notes:    req.GetNotes(),
	}
	if _, err := c.Store.Machines().Create(ctx, m); err != nil {
		return nil, status.Errorf(codes.Internal, "create machine: %v", err)
	}
	return &habitv1.CreateMachineResponse{Machine: machineToProto(m)}, nil
}

func (c *HabitController) UpdateMachine(ctx context.Context, req *habitv1.UpdateMachineRequest) (*habitv1.UpdateMachineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	m := &store.Machine{
		ID:       req.GetId(),
		UserID:   uid,
		Name:     req.GetName(),
		Location: req.GetLocation(),
		Notes:    req.GetNotes(),
	}
	if err := c.Store.Machines().Update(ctx, m); err != nil {
		return nil, mapStoreErr(err, "machine")
	}
	updated, err := c.Store.Machines().Get(ctx, uid, m.ID)
	if err != nil {
		return nil, mapStoreErr(err, "machine")
	}
	return &habitv1.UpdateMachineResponse{Machine: machineToProto(updated)}, nil
}

func (c *HabitController) DeleteMachine(ctx context.Context, req *habitv1.DeleteMachineRequest) (*habitv1.DeleteMachineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.Store.Machines().Delete(ctx, uid, req.GetId()); err != nil {
		return nil, mapStoreErr(err, "machine")
	}
	return &habitv1.DeleteMachineResponse{}, nil
}

// --- Exercises ---

func (c *HabitController) ListExercises(ctx context.Context, req *habitv1.ListExercisesRequest) (*habitv1.ListExercisesResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	es, err := c.Store.Exercises().List(ctx, uid, req.GetIncludeArchived())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list exercises: %v", err)
	}
	out := make([]*habitv1.Exercise, 0, len(es))
	for _, e := range es {
		out = append(out, exerciseToProto(e))
	}
	return &habitv1.ListExercisesResponse{Exercises: out}, nil
}

func (c *HabitController) GetExercise(ctx context.Context, req *habitv1.GetExerciseRequest) (*habitv1.GetExerciseResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	e, err := c.Store.Exercises().Get(ctx, uid, req.GetId())
	if err != nil {
		return nil, mapStoreErr(err, "exercise")
	}
	return &habitv1.GetExerciseResponse{Exercise: exerciseToProto(e)}, nil
}

func (c *HabitController) CreateExercise(ctx context.Context, req *habitv1.CreateExerciseRequest) (*habitv1.CreateExerciseResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetKind() == habitv1.ExerciseKind_EXERCISE_KIND_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "kind is required")
	}
	e := &store.Exercise{
		UserID:       uid,
		Name:         req.GetName(),
		Kind:         int32(req.GetKind()),
		Instructions: req.GetInstructions(),
	}
	if req.MachineId != nil {
		v := *req.MachineId
		e.MachineID = &v
	}
	if req.GoalCount != nil {
		v := *req.GoalCount
		e.GoalCount = &v
	}
	if req.GoalDurationSeconds != nil {
		v := *req.GoalDurationSeconds
		e.GoalDurationSeconds = &v
	}
	if req.GoalSets != nil {
		v := *req.GoalSets
		e.GoalSets = &v
	}
	if req.GoalWeightLb != nil {
		v := *req.GoalWeightLb
		e.GoalWeightLb = &v
	}
	if _, err := c.Store.Exercises().Create(ctx, e); err != nil {
		return nil, status.Errorf(codes.Internal, "create exercise: %v", err)
	}
	return &habitv1.CreateExerciseResponse{Exercise: exerciseToProto(e)}, nil
}

func (c *HabitController) UpdateExercise(ctx context.Context, req *habitv1.UpdateExerciseRequest) (*habitv1.UpdateExerciseResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	e := &store.Exercise{
		ID:           req.GetId(),
		UserID:       uid,
		Name:         req.GetName(),
		Kind:         int32(req.GetKind()),
		Instructions: req.GetInstructions(),
	}
	if req.MachineId != nil {
		v := *req.MachineId
		e.MachineID = &v
	}
	if req.GoalCount != nil {
		v := *req.GoalCount
		e.GoalCount = &v
	}
	if req.GoalDurationSeconds != nil {
		v := *req.GoalDurationSeconds
		e.GoalDurationSeconds = &v
	}
	if req.GoalSets != nil {
		v := *req.GoalSets
		e.GoalSets = &v
	}
	if req.GoalWeightLb != nil {
		v := *req.GoalWeightLb
		e.GoalWeightLb = &v
	}
	if err := c.Store.Exercises().Update(ctx, e); err != nil {
		return nil, mapStoreErr(err, "exercise")
	}
	updated, err := c.Store.Exercises().Get(ctx, uid, e.ID)
	if err != nil {
		return nil, mapStoreErr(err, "exercise")
	}
	return &habitv1.UpdateExerciseResponse{Exercise: exerciseToProto(updated)}, nil
}

func (c *HabitController) ArchiveExercise(ctx context.Context, req *habitv1.ArchiveExerciseRequest) (*habitv1.ArchiveExerciseResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.Store.Exercises().Archive(ctx, uid, req.GetId(), c.Now()); err != nil {
		return nil, mapStoreErr(err, "exercise")
	}
	return &habitv1.ArchiveExerciseResponse{}, nil
}

// --- Routines ---

func (c *HabitController) ListRoutines(ctx context.Context, req *habitv1.ListRoutinesRequest) (*habitv1.ListRoutinesResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	rs, err := c.Store.Routines().List(ctx, uid, req.GetIncludeArchived())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list routines: %v", err)
	}
	out := make([]*habitv1.Routine, 0, len(rs))
	for _, r := range rs {
		out = append(out, routineToProto(r))
	}
	return &habitv1.ListRoutinesResponse{Routines: out}, nil
}

func (c *HabitController) GetRoutine(ctx context.Context, req *habitv1.GetRoutineRequest) (*habitv1.GetRoutineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	r, err := c.Store.Routines().Get(ctx, uid, req.GetId())
	if err != nil {
		return nil, mapStoreErr(err, "routine")
	}
	return &habitv1.GetRoutineResponse{Routine: routineToProto(r)}, nil
}

func (c *HabitController) CreateRoutine(ctx context.Context, req *habitv1.CreateRoutineRequest) (*habitv1.CreateRoutineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	r, err := c.Store.Routines().Create(ctx, uid, req.GetName(), req.GetExerciseIds(), req.GetAlternateSets())
	if err != nil {
		return nil, mapStoreErr(err, "routine")
	}
	return &habitv1.CreateRoutineResponse{Routine: routineToProto(r)}, nil
}

func (c *HabitController) UpdateRoutine(ctx context.Context, req *habitv1.UpdateRoutineRequest) (*habitv1.UpdateRoutineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	var ids []int64
	if len(req.GetExerciseIds()) > 0 {
		ids = req.GetExerciseIds()
	}
	r, err := c.Store.Routines().Update(ctx, uid, req.GetId(), req.GetName(), ids, req.GetAlternateSets())
	if err != nil {
		return nil, mapStoreErr(err, "routine")
	}
	return &habitv1.UpdateRoutineResponse{Routine: routineToProto(r)}, nil
}

func (c *HabitController) ArchiveRoutine(ctx context.Context, req *habitv1.ArchiveRoutineRequest) (*habitv1.ArchiveRoutineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.Store.Routines().Archive(ctx, uid, req.GetId(), c.Now()); err != nil {
		return nil, mapStoreErr(err, "routine")
	}
	return &habitv1.ArchiveRoutineResponse{}, nil
}

func (c *HabitController) StartRoutine(ctx context.Context, req *habitv1.StartRoutineRequest) (*habitv1.StartRoutineResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	run, err := c.Store.Routines().StartRun(ctx, uid, req.GetRoutineId(), c.Now())
	if err != nil {
		return nil, mapStoreErr(err, "routine")
	}
	return &habitv1.StartRoutineResponse{Run: routineRunToProto(run)}, nil
}

func (c *HabitController) EndRoutineRun(ctx context.Context, req *habitv1.EndRoutineRunRequest) (*habitv1.EndRoutineRunResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	run, err := c.Store.Routines().EndRun(ctx, uid, req.GetRunId(), c.Now())
	if err != nil {
		return nil, mapStoreErr(err, "routine run")
	}
	return &habitv1.EndRoutineRunResponse{Run: routineRunToProto(run)}, nil
}

// --- Sessions ---

func (c *HabitController) LogSession(ctx context.Context, req *habitv1.LogSessionRequest) (*habitv1.LogSessionResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetClientId() == "" {
		return nil, status.Error(codes.InvalidArgument, "client_id is required")
	}
	if req.GetExerciseId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "exercise_id is required")
	}

	// Verify exercise belongs to this user before logging against it.
	if _, err := c.Store.Exercises().Get(ctx, uid, req.GetExerciseId()); err != nil {
		return nil, mapStoreErr(err, "exercise")
	}
	// Same for the routine run, if present.
	if req.RoutineRunId != nil {
		if _, err := c.Store.Routines().GetRun(ctx, uid, *req.RoutineRunId); err != nil {
			return nil, mapStoreErr(err, "routine run")
		}
	}

	sn := &store.Session{
		UserID:     uid,
		ExerciseID: req.GetExerciseId(),
		StartedAt:  req.GetStartedAt().AsTime(),
		Status:     int32(req.GetStatus()),
		Notes:      req.GetNotes(),
		ClientID:   req.GetClientId(),
	}
	if req.RoutineRunId != nil {
		v := *req.RoutineRunId
		sn.RoutineRunID = &v
	}
	if req.EndedAt != nil {
		t := req.EndedAt.AsTime()
		sn.EndedAt = &t
	}
	if req.CountCompleted != nil {
		v := *req.CountCompleted
		sn.CountCompleted = &v
	}
	if req.CountGoal != nil {
		v := *req.CountGoal
		sn.CountGoal = &v
	}
	if req.DurationSeconds != nil {
		v := *req.DurationSeconds
		sn.DurationSeconds = &v
	}
	if req.Difficulty != nil {
		d := int32(*req.Difficulty)
		sn.Difficulty = &d
	}
	if req.SetIndex != nil {
		v := *req.SetIndex
		sn.SetIndex = &v
	}
	if req.SetTotal != nil {
		v := *req.SetTotal
		sn.SetTotal = &v
	}
	if req.WeightLb != nil {
		v := *req.WeightLb
		sn.WeightLb = &v
	}

	logged, err := c.Store.Sessions().Log(ctx, sn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "log session: %v", err)
	}
	return &habitv1.LogSessionResponse{Session: sessionToProto(logged)}, nil
}

func (c *HabitController) ListSessions(ctx context.Context, req *habitv1.ListSessionsRequest) (*habitv1.ListSessionsResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	f := store.SessionFilter{Limit: req.GetLimit()}
	if req.From != nil {
		t := req.From.AsTime()
		f.From = &t
	}
	if req.To != nil {
		t := req.To.AsTime()
		f.To = &t
	}
	if req.ExerciseId != nil {
		v := *req.ExerciseId
		f.ExerciseID = &v
	}
	if req.RoutineRunId != nil {
		v := *req.RoutineRunId
		f.RoutineRunID = &v
	}

	ss, err := c.Store.Sessions().List(ctx, uid, f)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list sessions: %v", err)
	}
	out := make([]*habitv1.Session, 0, len(ss))
	for _, s := range ss {
		out = append(out, sessionToProto(s))
	}
	return &habitv1.ListSessionsResponse{Sessions: out}, nil
}

// --- Reports ---

func (c *HabitController) GetWeeklyReport(ctx context.Context, req *habitv1.GetWeeklyReportRequest) (*habitv1.GetWeeklyReportResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	week := req.GetWeek()
	if week == "" {
		y, w := c.Now().UTC().ISOWeek()
		week = fmt.Sprintf("%04d-W%02d", y, w)
	}
	from, to, err := isoWeekRange(week)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	rollups, err := c.Store.Reports().Rollup(ctx, uid, from, to)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rollup: %v", err)
	}
	return &habitv1.GetWeeklyReportResponse{Report: &habitv1.PeriodReport{
		Period:  week,
		Rollups: rollupsToProto(rollups),
	}}, nil
}

func (c *HabitController) GetMonthlyReport(ctx context.Context, req *habitv1.GetMonthlyReportRequest) (*habitv1.GetMonthlyReportResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	month := req.GetMonth()
	if month == "" {
		t := c.Now().UTC()
		month = fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))
	}
	from, to, err := monthRange(month)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	rollups, err := c.Store.Reports().Rollup(ctx, uid, from, to)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rollup: %v", err)
	}
	return &habitv1.GetMonthlyReportResponse{Report: &habitv1.PeriodReport{
		Period:  month,
		Rollups: rollupsToProto(rollups),
	}}, nil
}

func (c *HabitController) ListProgressionSuggestions(ctx context.Context, _ *habitv1.ListProgressionSuggestionsRequest) (*habitv1.ListProgressionSuggestionsResponse, error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	sugs, err := c.Suggester.ListSuggestions(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "suggestions: %v", err)
	}
	return &habitv1.ListProgressionSuggestionsResponse{Suggestions: sugs}, nil
}

// --- helpers ---

func userID(ctx context.Context) (int64, error) {
	uid, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "not signed in")
	}
	return uid, nil
}

func mapStoreErr(err error, kind string) error {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return status.Errorf(codes.NotFound, "%s not found", kind)
	case errors.Is(err, store.ErrConflict):
		return status.Errorf(codes.AlreadyExists, "%s conflict", kind)
	case errors.Is(err, store.ErrForbidden):
		return status.Errorf(codes.PermissionDenied, "%s forbidden", kind)
	default:
		return status.Errorf(codes.Internal, "%s: %v", kind, err)
	}
}

func rollupsToProto(rs []*store.ExerciseRollup) []*habitv1.ExerciseRollup {
	out := make([]*habitv1.ExerciseRollup, 0, len(rs))
	for _, r := range rs {
		out = append(out, &habitv1.ExerciseRollup{
			ExerciseId:           r.ExerciseID,
			ExerciseName:         r.ExerciseName,
			Kind:                 habitv1.ExerciseKind(r.Kind),
			SessionsTotal:        r.SessionsTotal,
			SessionsCompleted:    r.SessionsCompleted,
			SessionsSkipped:      r.SessionsSkipped,
			SessionsFailed:       r.SessionsFailed,
			TotalCount:           r.TotalCount,
			TotalDurationSeconds: r.TotalDurationSeconds,
			AvgDifficulty:        r.AvgDifficulty,
		})
	}
	return out
}

// isoWeekRange parses a "YYYY-Www" string and returns the [from, to) UTC range
// covering the corresponding ISO week (Monday 00:00 UTC, exclusive of the next Monday).
func isoWeekRange(s string) (time.Time, time.Time, error) {
	if len(s) != 8 || s[4:6] != "-W" {
		return time.Time{}, time.Time{}, fmt.Errorf("week %q must be YYYY-Www", s)
	}
	year, err := strconv.Atoi(s[0:4])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("week year: %w", err)
	}
	week, err := strconv.Atoi(s[6:8])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("week number: %w", err)
	}
	// Find ISO week 1: contains Jan 4. The Monday on/before Jan 4 of the given
	// year is the start of week 1. Week N starts (N-1)*7 days later.
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7 // Make Sunday=7 so Monday=1.
	}
	week1Mon := jan4.AddDate(0, 0, -(weekday - 1))
	from := week1Mon.AddDate(0, 0, (week-1)*7)
	to := from.AddDate(0, 0, 7)
	return from, to, nil
}

// monthRange parses "YYYY-MM" and returns [from, to) UTC range covering that calendar month.
func monthRange(s string) (time.Time, time.Time, error) {
	if len(s) != 7 || s[4] != '-' {
		return time.Time{}, time.Time{}, fmt.Errorf("month %q must be YYYY-MM", s)
	}
	year, err := strconv.Atoi(s[0:4])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("month year: %w", err)
	}
	mm, err := strconv.Atoi(s[5:7])
	if err != nil || mm < 1 || mm > 12 {
		return time.Time{}, time.Time{}, fmt.Errorf("month month: %v", err)
	}
	from := time.Date(year, time.Month(mm), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	return from, to, nil
}
