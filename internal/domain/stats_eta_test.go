package domain_test

import (
	"testing"

	"taskflow/internal/domain"
)

func TestETAWithHistory(t *testing.T) {
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 25, State: domain.SessionStateCompleted},
		{WorkDuration: 25, State: domain.SessionStateCompleted},
	}
	eta, err := domain.CalculateETA(sessions, []int{2, 2}, 6)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 75.0 {
		t.Errorf("expected ETA=75.0, got %f", eta)
	}
}

func TestETANoHistory(t *testing.T) {
	_, err := domain.CalculateETA(nil, nil, 5)
	if err == nil {
		t.Error("expected error when no session history, got nil")
	}
}

func TestETAZeroRemaining(t *testing.T) {
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 25, State: domain.SessionStateCompleted},
	}
	eta, err := domain.CalculateETA(sessions, []int{2}, 0)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 0.0 {
		t.Errorf("expected ETA=0.0, got %f", eta)
	}
}

func TestETAFallbackWhenNoTasksCompleted(t *testing.T) {
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 30, State: domain.SessionStateCompleted},
	}
	eta, err := domain.CalculateETA(sessions, []int{0}, 3)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 90.0 {
		t.Errorf("expected ETA=90.0 with fallback, got %f", eta)
	}
}

func TestETAAsymmetricHistory(t *testing.T) {
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 20, State: domain.SessionStateCompleted},
		{WorkDuration: 40, State: domain.SessionStateCompleted},
	}
	eta, err := domain.CalculateETA(sessions, []int{1, 3}, 4)
	if err != nil {
		t.Fatalf("CalculateETA: %v", err)
	}
	if eta != 60.0 {
		t.Errorf("expected ETA=60.0, got %f", eta)
	}
}
