package pomodoro

import (
	"errors"
	"time"
)

type PomodoroState interface {
	Start(m *PomodoroMachine) error
	Pause(m *PomodoroMachine) error
	Resume(m *PomodoroMachine) error
	Complete(m *PomodoroMachine) error
	Cancel(m *PomodoroMachine) error
	Name() string
}

var errInvalidTransition = errors.New("invalid state transition")

type PomodoroMachine struct {
	state         PomodoroState
	WorkDuration  int
	StartTime     *time.Time
	FinishTime    *time.Time
	RemainingTime int
}

func NewPomodoroMachine(workDurationMin int) *PomodoroMachine {
	return &PomodoroMachine{
		state:        &IdleState{},
		WorkDuration: workDurationMin,
	}
}

func (m *PomodoroMachine) SetState(s PomodoroState) { m.state = s }
func (m *PomodoroMachine) StateName() string        { return m.state.Name() }

func (m *PomodoroMachine) Start() error    { return m.state.Start(m) }
func (m *PomodoroMachine) Pause() error    { return m.state.Pause(m) }
func (m *PomodoroMachine) Resume() error   { return m.state.Resume(m) }
func (m *PomodoroMachine) Complete() error { return m.state.Complete(m) }
func (m *PomodoroMachine) Cancel() error   { return m.state.Cancel(m) }

func (m *PomodoroMachine) Tick() bool {
	if m.state.Name() != "RUNNING" {
		return false
	}
	m.RemainingTime--
	if m.RemainingTime <= 0 {
		m.RemainingTime = 0
		m.Complete()
		return true
	}
	return false
}

type IdleState struct{}

func (s *IdleState) Name() string { return "IDLE" }
func (s *IdleState) Start(m *PomodoroMachine) error {
	now := time.Now()
	m.StartTime = &now
	m.RemainingTime = m.WorkDuration * 60
	m.SetState(&RunningState{})
	return nil
}
func (s *IdleState) Pause(m *PomodoroMachine) error    { return errInvalidTransition }
func (s *IdleState) Resume(m *PomodoroMachine) error   { return errInvalidTransition }
func (s *IdleState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *IdleState) Cancel(m *PomodoroMachine) error   { return errInvalidTransition }

type RunningState struct{}

func (s *RunningState) Name() string                   { return "RUNNING" }
func (s *RunningState) Start(m *PomodoroMachine) error { return errInvalidTransition }
func (s *RunningState) Pause(m *PomodoroMachine) error {
	m.SetState(&PausedState{})
	return nil
}
func (s *RunningState) Resume(m *PomodoroMachine) error { return errInvalidTransition }
func (s *RunningState) Complete(m *PomodoroMachine) error {
	now := time.Now()
	m.FinishTime = &now
	m.RemainingTime = 0
	m.SetState(&CompletedState{})
	return nil
}
func (s *RunningState) Cancel(m *PomodoroMachine) error {
	m.SetState(&CancelledState{})
	return nil
}

type PausedState struct{}

func (s *PausedState) Name() string                   { return "PAUSED" }
func (s *PausedState) Start(m *PomodoroMachine) error { return errInvalidTransition }
func (s *PausedState) Pause(m *PomodoroMachine) error { return errInvalidTransition }
func (s *PausedState) Resume(m *PomodoroMachine) error {
	m.SetState(&RunningState{})
	return nil
}
func (s *PausedState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *PausedState) Cancel(m *PomodoroMachine) error {
	m.SetState(&CancelledState{})
	return nil
}

type CompletedState struct{}

func (s *CompletedState) Name() string                      { return "COMPLETED" }
func (s *CompletedState) Start(m *PomodoroMachine) error    { return errInvalidTransition }
func (s *CompletedState) Pause(m *PomodoroMachine) error    { return errInvalidTransition }
func (s *CompletedState) Resume(m *PomodoroMachine) error   { return errInvalidTransition }
func (s *CompletedState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *CompletedState) Cancel(m *PomodoroMachine) error   { return errInvalidTransition }

type CancelledState struct{}

func (s *CancelledState) Name() string                      { return "CANCELLED" }
func (s *CancelledState) Start(m *PomodoroMachine) error    { return errInvalidTransition }
func (s *CancelledState) Pause(m *PomodoroMachine) error    { return errInvalidTransition }
func (s *CancelledState) Resume(m *PomodoroMachine) error   { return errInvalidTransition }
func (s *CancelledState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *CancelledState) Cancel(m *PomodoroMachine) error   { return errInvalidTransition }
