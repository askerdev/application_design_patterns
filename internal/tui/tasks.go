package tui

import (
	"fmt"
	"strings"

	"taskflow/internal/domain"
)

func (a *App) tasksMenu() {
	for {
		fmt.Printf("\n%s-- Tasks --%s\n", colorCyan, colorReset)
		tasks, _ := a.taskRepo.GetAllByUser(a.currentUser.ID)
		if len(tasks) == 0 {
			fmt.Println("  (no tasks)")
		}
		for _, t := range tasks {
			overdue := ""
			if t.IsOverdue() {
				overdue = colorRed + " OVERDUE" + colorReset
			}
			proj := ""
			if t.ProjectID != nil {
				proj = fmt.Sprintf(" [proj:%d]", *t.ProjectID)
			}
			fmt.Printf("  %d. [%s][%s]%s %s%s\n", t.ID, t.Priority, t.Status, proj, t.Content, overdue)
		}
		fmt.Println("\n  a.Add  c.Complete  d.Delete  b.Back")
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "a":
			a.addTask()
		case "c":
			idStr := a.readLine("Task ID to complete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			a.completeTask(id)
		case "d":
			idStr := a.readLine("Task ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.taskRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) addTask() {
	content := strings.TrimSpace(a.readLine("Content: "))
	if content == "" {
		return
	}
	prioStr := strings.ToUpper(strings.TrimSpace(a.readLine("Priority H/M/L [M]: ")))
	priority := domain.PriorityMedium
	switch prioStr {
	case "H":
		priority = domain.PriorityHigh
	case "L":
		priority = domain.PriorityLow
	}
	projectID := a.pickProject()
	tagID := a.pickTag()

	task := &domain.Task{
		UserID:    a.currentUser.ID,
		ProjectID: projectID,
		TagID:     tagID,
		Content:   content,
		Status:    domain.TaskStatusTodo,
		Priority:  priority,
	}
	if err := a.taskRepo.Create(task); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%sTask created (id=%d)%s\n", colorGreen, task.ID, colorReset)
	}
}

func (a *App) completeTask(id int64) {
	task, err := a.taskRepo.GetByID(id)
	if err != nil {
		fmt.Println("Task not found")
		return
	}
	task.Complete()
	if err := a.taskRepo.Update(task); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(colorGreen + "Marked as done." + colorReset)
	}
}
