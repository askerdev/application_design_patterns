package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// GanttItem — элемент таймлайна для одной задачи.
type GanttItem struct {
	TaskID      int64     `json:"task_id"`
	Title       string    `json:"title"`
	Priority    string    `json:"priority"`
	StoryPoints int       `json:"story_points"`
	Status      string    `json:"status"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Note        string    `json:"note,omitempty"`
}

// GanttPlan — таймлайн проекта.
type GanttPlan struct {
	Project    *Project    `json:"project"`
	Items      []GanttItem `json:"items"`
	Generated  time.Time   `json:"generated"`
	UsedLLM    bool        `json:"used_llm"`
	WarningMsg string      `json:"warning,omitempty"`
}

// GanttPlanner — сервис, формирующий таймлайн проекта.
// Сначала пытается через LLM, если не получилось — детерминированный фолбэк.
type GanttPlanner interface {
	Plan(projectID int64) (*GanttPlan, error)
}

type ganttPlanner struct {
	tasks    TaskRepository
	projects ProjectRepository
	llm      LLMGenerator
}

func NewGanttPlanner(tasks TaskRepository, projects ProjectRepository, llm LLMGenerator) GanttPlanner {
	return &ganttPlanner{tasks: tasks, projects: projects, llm: llm}
}

func (g *ganttPlanner) Plan(projectID int64) (*GanttPlan, error) {
	project, err := g.projects.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	tasks, err := g.tasks.GetByProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	if len(tasks) == 0 {
		return &GanttPlan{Project: project, Items: nil, Generated: time.Now()}, nil
	}

	plan := &GanttPlan{Project: project, Generated: time.Now()}

	if g.llm != nil {
		items, llmErr := g.planViaLLM(project, tasks)
		if llmErr == nil && len(items) > 0 {
			plan.Items = items
			plan.UsedLLM = true
			return plan, nil
		}
		if llmErr != nil {
			plan.WarningMsg = "LLM-плэннер недоступен, использован фолбэк: " + llmErr.Error()
		}
	}

	plan.Items = fallbackPlan(project, tasks)
	return plan, nil
}

func (g *ganttPlanner) planViaLLM(project *Project, tasks []*Task) ([]GanttItem, error) {
	prompt := buildGanttPrompt(project, tasks)
	raw, err := g.llm.Generate(prompt)
	if err != nil {
		return nil, err
	}
	return parseGanttResponse(raw, tasks)
}

func buildGanttPrompt(project *Project, tasks []*Task) string {
	now := time.Now()
	var sb strings.Builder
	sb.WriteString("You are a project planning assistant. Build a Gantt timeline (start/end dates) for the tasks of a project.\n")
	sb.WriteString("Rules:\n")
	sb.WriteString("- Today is " + now.Format("2006-01-02") + ".\n")
	sb.WriteString("- Higher priority (HIGH > MEDIUM > LOW) and tasks with closer due dates go FIRST.\n")
	sb.WriteString("- Estimate duration from story_points: 1 SP ≈ 1 working day. Minimum 1 day per task.\n")
	sb.WriteString("- Up to 2 tasks may run in parallel; a single task is contiguous.\n")
	sb.WriteString("- Try not to schedule tasks past the project deadline if specified.\n")
	sb.WriteString("- Tasks with status DONE: start = end = today (they are already done).\n\n")

	if project.DueDate != nil {
		fmt.Fprintf(&sb, "Project: %s, deadline %s\n", project.Name, project.DueDate.Format("2006-01-02"))
	} else {
		fmt.Fprintf(&sb, "Project: %s, no fixed deadline\n", project.Name)
	}
	sb.WriteString("Tasks (JSON):\n")

	type taskIn struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Priority    string `json:"priority"`
		StoryPoints int    `json:"story_points"`
		Status      string `json:"status"`
		DueDate     string `json:"due_date,omitempty"`
	}
	in := make([]taskIn, len(tasks))
	for i, t := range tasks {
		ti := taskIn{
			ID: t.ID, Title: t.Content, Priority: string(t.Priority),
			StoryPoints: t.StoryPoints, Status: string(t.Status),
		}
		if t.DueDate != nil {
			ti.DueDate = t.DueDate.Format("2006-01-02")
		}
		in[i] = ti
	}
	b, _ := json.MarshalIndent(in, "", "  ")
	sb.Write(b)
	sb.WriteString("\n\n")
	sb.WriteString("Reply with ONLY a JSON array (no prose, no markdown fences) of objects:\n")
	sb.WriteString(`[{"task_id": <int>, "start": "YYYY-MM-DD", "end": "YYYY-MM-DD", "note": "<short reason>"}]` + "\n")
	sb.WriteString("Cover EVERY task id from the input. End date must be >= start date.\n")
	return sb.String()
}

type llmGanttItem struct {
	TaskID int64  `json:"task_id"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Note   string `json:"note"`
}

func parseGanttResponse(raw string, tasks []*Task) ([]GanttItem, error) {
	jsonText, err := extractJSONArray(raw)
	if err != nil {
		return nil, err
	}
	var llmItems []llmGanttItem
	if err := json.Unmarshal([]byte(jsonText), &llmItems); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	if len(llmItems) == 0 {
		return nil, fmt.Errorf("empty plan from LLM")
	}

	taskByID := make(map[int64]*Task, len(tasks))
	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	out := make([]GanttItem, 0, len(llmItems))
	for _, li := range llmItems {
		t, ok := taskByID[li.TaskID]
		if !ok {
			continue
		}
		start, err := time.Parse("2006-01-02", li.Start)
		if err != nil {
			continue
		}
		end, err := time.Parse("2006-01-02", li.End)
		if err != nil {
			continue
		}
		if end.Before(start) {
			end = start
		}
		out = append(out, GanttItem{
			TaskID:      t.ID,
			Title:       t.Content,
			Priority:    string(t.Priority),
			StoryPoints: t.StoryPoints,
			Status:      string(t.Status),
			Start:       start,
			End:         end.Add(24*time.Hour - time.Second),
			Note:        strings.TrimSpace(li.Note),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("none of the LLM items matched known tasks")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out, nil
}

var jsonArrayRe = regexp.MustCompile(`(?s)\[.*\]`)

func extractJSONArray(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	if m := jsonArrayRe.FindString(s); m != "" {
		return m, nil
	}
	return "", fmt.Errorf("no JSON array found in LLM response")
}

// fallbackPlan детерминированно расставляет задачи (если LLM не сработал).
func fallbackPlan(project *Project, tasks []*Task) []GanttItem {
	priOrder := func(p Priority) int {
		switch p {
		case PriorityHigh:
			return 0
		case PriorityMedium:
			return 1
		default:
			return 2
		}
	}
	sorted := make([]*Task, len(tasks))
	copy(sorted, tasks)
	sort.SliceStable(sorted, func(i, j int) bool {
		ti, tj := sorted[i], sorted[j]
		if ti.IsOverdue() != tj.IsOverdue() {
			return ti.IsOverdue()
		}
		if priOrder(ti.Priority) != priOrder(tj.Priority) {
			return priOrder(ti.Priority) < priOrder(tj.Priority)
		}
		switch {
		case ti.DueDate != nil && tj.DueDate != nil:
			return ti.DueDate.Before(*tj.DueDate)
		case ti.DueDate != nil:
			return true
		case tj.DueDate != nil:
			return false
		}
		return ti.ID < tj.ID
	})

	const tracks = 2
	now := time.Now().Truncate(24 * time.Hour)
	trackEnd := make([]time.Time, tracks)
	for i := range trackEnd {
		trackEnd[i] = now
	}

	items := make([]GanttItem, 0, len(sorted))
	for _, t := range sorted {
		if t.Status == TaskStatusDone {
			items = append(items, GanttItem{
				TaskID: t.ID, Title: t.Content, Priority: string(t.Priority),
				StoryPoints: t.StoryPoints, Status: string(t.Status),
				Start: now, End: now.Add(24*time.Hour - time.Second),
				Note: "DONE",
			})
			continue
		}
		dur := t.StoryPoints
		if dur < 1 {
			dur = 1
		}
		idx := 0
		for i := 1; i < tracks; i++ {
			if trackEnd[i].Before(trackEnd[idx]) {
				idx = i
			}
		}
		start := trackEnd[idx]
		end := start.AddDate(0, 0, dur)
		trackEnd[idx] = end

		note := ""
		if t.IsOverdue() {
			note = "OVERDUE"
		} else if t.DueDate != nil && end.After(*t.DueDate) {
			note = "не укладывается в дедлайн задачи"
		}
		if project.DueDate != nil && end.After(*project.DueDate) {
			if note != "" {
				note += "; "
			}
			note += "выходит за дедлайн проекта"
		}

		items = append(items, GanttItem{
			TaskID:      t.ID,
			Title:       t.Content,
			Priority:    string(t.Priority),
			StoryPoints: t.StoryPoints,
			Status:      string(t.Status),
			Start:       start,
			End:         end.Add(-time.Second),
			Note:        note,
		})
	}
	return items
}
