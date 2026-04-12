package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"taskflow/internal/domain"
	"taskflow/internal/patterns/state"
)

func (a *App) pomodoroMenu() {
	for {
		fmt.Printf("\n%s-- Pomodoro --%s\n", colorCyan, colorReset)
		sessions, _ := a.pomodoroRepo.GetAllByUser(a.currentUser.ID)
		if len(sessions) == 0 {
			fmt.Println("  (no sessions)")
		}
		for _, s := range sessions {
			proj := ""
			if s.ProjectID != nil {
				proj = fmt.Sprintf(" [proj:%d]", *s.ProjectID)
			}
			start := ""
			if s.StartTime != nil {
				start = " @ " + s.StartTime.Format("2006-01-02 15:04")
			}
			fmt.Printf("  %d. [%s]%s %dmin%s\n", s.ID, s.State, proj, s.WorkDuration, start)
		}
		fmt.Println("\n  s.Start new session  b.Back")
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "s":
			a.startPomodoroSession()
		case "b":
			return
		}
	}
}

func (a *App) startPomodoroSession() {
	durStr := strings.TrimSpace(a.readLine("Work duration in minutes [25]: "))
	dur := 25
	if durStr != "" {
		fmt.Sscanf(durStr, "%d", &dur)
	}
	if dur <= 0 {
		dur = 25
	}
	projectID := a.pickProject()

	session := &domain.PomodoroSession{
		UserID:        a.currentUser.ID,
		ProjectID:     projectID,
		WorkDuration:  dur,
		RemainingTime: dur * 60,
		State:         domain.SessionStateIdle,
	}
	if err := a.pomodoroRepo.Create(session); err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	machine := state.NewPomodoroMachine(dur)
	a.runTimer(machine, session)
}

func (a *App) runTimer(machine *state.PomodoroMachine, session *domain.PomodoroSession) {
	// Start the state machine
	machine.Start()
	session.State = domain.SessionState(machine.StateName())
	a.pomodoroRepo.Update(session)

	// Try to set raw terminal mode for single-keypress input
	oldState, rawOK := setRaw()

	fmt.Println("\n" + colorBold + "Pomodoro started! Controls: p=pause  r=resume  c=complete  x=cancel" + colorReset)

	inputCh := make(chan byte, 4)
	go func() {
		buf := make([]byte, 1)
		for {
			n, _ := os.Stdin.Read(buf)
			if n > 0 {
				inputCh <- buf[0]
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		rem := machine.RemainingTime
		mins := rem / 60
		secs := rem % 60
		stateColor := colorGreen
		if machine.StateName() == "PAUSED" {
			stateColor = colorYellow
		}
		fmt.Printf("\r\033[K  %s[%s]%s  %02d:%02d  — p pause  r resume  c done  x cancel  ",
			stateColor, machine.StateName(), colorReset, mins, secs)

		select {
		case <-ticker.C:
			if machine.StateName() == "RUNNING" {
				machine.RemainingTime--
				session.RemainingTime = machine.RemainingTime
				if machine.RemainingTime <= 0 {
					machine.Complete()
					a.syncSession(machine, session)
					fmt.Printf("\n%sSession completed!%s\n", colorGreen, colorReset)
					if rawOK {
						restoreRaw(oldState)
					}
					return
				}
			}
		case key := <-inputCh:
			switch key {
			case 'p', 'P':
				machine.Pause()
			case 'r', 'R':
				machine.Resume()
			case 'c', 'C':
				machine.Complete()
				a.syncSession(machine, session)
				fmt.Printf("\n%sSession completed!%s\n", colorGreen, colorReset)
				if rawOK {
					restoreRaw(oldState)
				}
				return
			case 'x', 'X', 'q', 'Q', 3: // x, q, Ctrl+C
				machine.Cancel()
				a.syncSession(machine, session)
				fmt.Printf("\n%sSession cancelled.%s\n", colorRed, colorReset)
				if rawOK {
					restoreRaw(oldState)
				}
				return
			}
			a.syncSession(machine, session)
		}
	}
}

func (a *App) syncSession(machine *state.PomodoroMachine, session *domain.PomodoroSession) {
	session.State = domain.SessionState(machine.StateName())
	session.RemainingTime = machine.RemainingTime
	if machine.StartTime != nil {
		session.StartTime = machine.StartTime
	}
	if machine.FinishTime != nil {
		session.FinishTime = machine.FinishTime
	}
	a.pomodoroRepo.Update(session)
}

func setRaw() (*term.State, bool) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, false
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, false
	}
	return oldState, true
}

func restoreRaw(oldState *term.State) {
	if oldState != nil {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}
}
