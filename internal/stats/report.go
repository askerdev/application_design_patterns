package stats

import (
	"fmt"
	"strings"
)

// Report defines the two abstract steps each concrete report must implement.
type Report interface {
	fetchData(userID int64) error
	formatContent() string
}

// Generate is the template method — calls fetchData then formatContent in order.
// All reports share this skeleton; only the steps differ.
func Generate(r Report, userID int64) string {
	if err := r.fetchData(userID); err != nil {
		return "Error: " + err.Error()
	}
	return r.formatContent()
}

// TaskCountReport counts tasks by status.
type TaskCountReport struct {
	svc  Service
	data *Summary
}

// NewTaskCountReport returns a new TaskCountReport.
func NewTaskCountReport(svc Service) *TaskCountReport {
	return &TaskCountReport{svc: svc}
}

func (r *TaskCountReport) fetchData(userID int64) error {
	sum, err := r.svc.Summarize(userID)
	r.data = sum
	return err
}

func (r *TaskCountReport) formatContent() string {
	if r.data == nil {
		return "No data."
	}
	return fmt.Sprintf("Tasks: %d total  |  %d done  |  %d todo  |  %d in-progress",
		r.data.Total, r.data.Done, r.data.Todo, r.data.InProgress)
}

// ProjectETAReport shows per-project ETA.
type ProjectETAReport struct {
	svc  Service
	data *Summary
}

// NewProjectETAReport returns a new ProjectETAReport.
func NewProjectETAReport(svc Service) *ProjectETAReport {
	return &ProjectETAReport{svc: svc}
}

func (r *ProjectETAReport) fetchData(userID int64) error {
	sum, err := r.svc.Summarize(userID)
	r.data = sum
	return err
}

func (r *ProjectETAReport) formatContent() string {
	if r.data == nil || len(r.data.Projects) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Project ETA:\n")
	for _, p := range r.data.Projects {
		if !p.HasETA {
			fmt.Fprintf(&sb, "  %s: %d remaining (no history)\n", p.Name, p.Remaining)
		} else {
			hours := int(p.ETAMins) / 60
			mins := int(p.ETAMins) % 60
			fmt.Fprintf(&sb, "  %s: %d remaining — ETA ~%dh%dm\n", p.Name, p.Remaining, hours, mins)
		}
	}
	return sb.String()
}
