package pomodoro_test

import (
	"testing"

	"taskflow/internal/pomodoro"
)

func TestPomodoroStartFromIdle(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	if err := s.Start(); err != nil {
		t.Fatalf("Start from IDLE: %v", err)
	}
	if s.StateName() != "RUNNING" {
		t.Errorf("expected RUNNING, got %s", s.StateName())
	}
}

func TestPomodoroPauseResume(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	s.Start()

	if err := s.Pause(); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if s.StateName() != "PAUSED" {
		t.Errorf("expected PAUSED, got %s", s.StateName())
	}

	if err := s.Resume(); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if s.StateName() != "RUNNING" {
		t.Errorf("expected RUNNING after resume, got %s", s.StateName())
	}
}

func TestPomodoroComplete(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	s.Start()
	if err := s.Complete(); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if s.StateName() != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", s.StateName())
	}
}

func TestPomodoroCancel(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	s.Start()
	if err := s.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if s.StateName() != "CANCELLED" {
		t.Errorf("expected CANCELLED, got %s", s.StateName())
	}
}

func TestPomodoroInvalidTransitionFromIdle(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	if err := s.Pause(); err == nil {
		t.Error("expected error when Pause from IDLE, got nil")
	}
}

func TestPomodoroCannotStartTwice(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	s.Start()
	if err := s.Start(); err == nil {
		t.Error("expected error when Start from RUNNING, got nil")
	}
}

func TestPomodoroStartSetsRemainingTime(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	s.Start()
	if s.RemainingTime != 25*60 {
		t.Errorf("expected RemainingTime=1500, got %d", s.RemainingTime)
	}
}

func TestPomodoroCancelFromPaused(t *testing.T) {
	s := pomodoro.NewPomodoroMachine(25)
	s.Start()
	s.Pause()
	if err := s.Cancel(); err != nil {
		t.Fatalf("Cancel from PAUSED: %v", err)
	}
	if s.StateName() != "CANCELLED" {
		t.Errorf("expected CANCELLED, got %s", s.StateName())
	}
}
