package tui

import (
	"fmt"
	"strings"
	"time"

	"taskflow/internal/domain"
)

func (a *App) remindersMenu() {
	for {
		fmt.Printf("\n%s-- Reminders --%s\n", colorCyan, colorReset)
		reminders, _ := a.reminderRepo.GetAllByUser(a.currentUser.ID)
		if len(reminders) == 0 {
			fmt.Println("  (no reminders)")
		}
		for _, r := range reminders {
			ready := ""
			if r.IsReady() {
				ready = colorYellow + " READY" + colorReset
			}
			proj := ""
			if r.ProjectID != nil {
				proj = fmt.Sprintf(" [proj:%d]", *r.ProjectID)
			}
			fmt.Printf("  %d. [%s]%s%s %s @ %s\n",
				r.ID, r.Status, proj, ready, r.Content, r.ReminderTime.Format("2006-01-02 15:04"))
		}
		fmt.Println("\n  a.Add  s.Send pending  d.Delete  b.Back")
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "a":
			a.addReminder()
		case "s":
			if err := a.reminderService.CheckAndNotify(); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Pending reminders processed." + colorReset)
			}
		case "d":
			idStr := a.readLine("Reminder ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.reminderRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) addReminder() {
	content := strings.TrimSpace(a.readLine("Content: "))
	if content == "" {
		return
	}
	timeStr := strings.TrimSpace(a.readLine("Time (YYYY-MM-DD HH:MM): "))
	t, err := time.ParseInLocation("2006-01-02 15:04", timeStr, time.Local)
	if err != nil {
		fmt.Println("Invalid time format. Use YYYY-MM-DD HH:MM")
		return
	}
	projectID := a.pickProject()
	tagID := a.pickTag()
	r := &domain.Reminder{
		UserID:       a.currentUser.ID,
		ProjectID:    projectID,
		TagID:        tagID,
		Content:      content,
		ReminderTime: t,
		Status:       domain.ReminderStatusPending,
	}
	if err := a.reminderRepo.Create(r); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%sReminder created (id=%d)%s\n", colorGreen, r.ID, colorReset)
	}
}
