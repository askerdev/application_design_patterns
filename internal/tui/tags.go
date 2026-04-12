package tui

import (
	"fmt"
	"strings"

	"taskflow/internal/domain"
)

func (a *App) tagsMenu() {
	for {
		fmt.Printf("\n%s-- Tags --%s\n", colorCyan, colorReset)
		tags, _ := a.tagRepo.GetAllByUser(a.currentUser.ID)
		if len(tags) == 0 {
			fmt.Println("  (no tags)")
		}
		for _, t := range tags {
			fmt.Printf("  %d. %s  %s■%s\n", t.ID, t.Name, t.Color, colorReset)
		}
		fmt.Println("\n  a.Add  d.Delete  b.Back")
		switch strings.TrimSpace(a.readLine("Choice: ")) {
		case "a":
			name := strings.TrimSpace(a.readLine("Tag name: "))
			color := strings.TrimSpace(a.readLine("Color (e.g. #ff0000): "))
			if color == "" {
				color = "#ffffff"
			}
			tag := &domain.Tag{
				UserID: a.currentUser.ID,
				Name:   name,
				Color:  color,
			}
			if err := a.tagRepo.Create(tag); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sTag created (id=%d)%s\n", colorGreen, tag.ID, colorReset)
			}
		case "d":
			idStr := a.readLine("Tag ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.tagRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}
