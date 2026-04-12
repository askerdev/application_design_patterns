package tui

import (
	"fmt"
	"strings"

	"taskflow/internal/domain"
)

func (a *App) notesMenu() {
	for {
		fmt.Printf("\n%s-- Notes --%s\n", colorCyan, colorReset)
		notes, _ := a.noteRepo.GetAllByUser(a.currentUser.ID)
		if len(notes) == 0 {
			fmt.Println("  (no notes)")
		}
		for _, n := range notes {
			proj := ""
			if n.ProjectID != nil {
				proj = fmt.Sprintf(" [proj:%d]", *n.ProjectID)
			}
			fmt.Printf("  %d.%s %s\n", n.ID, proj, n.Title)
		}
		fmt.Println("\n  a.Add  d.Delete  b.Back")
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "a":
			title := strings.TrimSpace(a.readLine("Title: "))
			content := strings.TrimSpace(a.readLine("Content: "))
			projectID := a.pickProject()
			tagID := a.pickTag()
			n := &domain.Note{
				UserID:    a.currentUser.ID,
				ProjectID: projectID,
				TagID:     tagID,
				Title:     title,
				Content:   content,
			}
			if err := a.noteRepo.Create(n); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sNote created (id=%d)%s\n", colorGreen, n.ID, colorReset)
			}
		case "d":
			idStr := a.readLine("Note ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.noteRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}
