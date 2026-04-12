package math_test

import (
	"testing"

	taskmath "taskflow/internal/math"
	"taskflow/internal/domain"
)

func TestETAWithHistory(t *testing.T) {
	// 2 completed sessions: 25 min each → avg = 25 min
	// 4 tasks completed total across 2 sessions → avg 2 tasks/session
	// 6 remaining tasks → ETA = 6/2 * 25 = 75 min
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 25, State: domain.SessionStateCompleted},
		{WorkDuration: 25, State: domain.SessionStateCompleted},
	}
	completedPerSession := []int{2, 2}
	remaining := 6

	eta, err := taskmath.CalculateETA(sessions, completedPerSession, remaining)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 75.0 {
		t.Errorf("expected ETA=75.0, got %f", eta)
	}
}

func TestETANoHistory(t *testing.T) {
	_, err := taskmath.CalculateETA(nil, nil, 5)
	if err == nil {
		t.Error("expected error when no session history, got nil")
	}
}

func TestETAZeroRemaining(t *testing.T) {
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 25, State: domain.SessionStateCompleted},
	}
	eta, err := taskmath.CalculateETA(sessions, []int{2}, 0)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 0.0 {
		t.Errorf("expected ETA=0.0, got %f", eta)
	}
}

func TestETAFallbackWhenNoTasksCompleted(t *testing.T) {
	// sessions exist but 0 tasks completed → fallback 1 task/session
	// 1 session of 30 min, 3 remaining → ETA = 3/1 * 30 = 90 min
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 30, State: domain.SessionStateCompleted},
	}
	eta, err := taskmath.CalculateETA(sessions, []int{0}, 3)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 90.0 {
		t.Errorf("expected ETA=90.0 with fallback, got %f", eta)
	}
}

func TestETAAsymmetricHistory(t *testing.T) {
	// 2 sessions: 20 and 40 min → avg 30 min
	// 1 and 3 tasks → avg 2 tasks/session
	// 4 remaining → ETA = 4/2 * 30 = 60 min
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 20, State: domain.SessionStateCompleted},
		{WorkDuration: 40, State: domain.SessionStateCompleted},
	}
	eta, err := taskmath.CalculateETA(sessions, []int{1, 3}, 4)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 60.0 {
		t.Errorf("expected ETA=60.0, got %f", eta)
	}
}
