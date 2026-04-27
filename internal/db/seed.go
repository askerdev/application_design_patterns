package db

import (
	"database/sql"
	"fmt"
)

func RunSeed(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRow(`SELECT id FROM users LIMIT 1`).Scan(&userID)
	if err == sql.ErrNoRows {
		res, err := tx.Exec(`INSERT INTO users(username) VALUES('demo')`)
		if err != nil {
			return fmt.Errorf("seed user: %w", err)
		}
		userID, _ = res.LastInsertId()
	} else if err != nil {
		return err
	}

	projectIDs := make([]int64, 0, 5)
	projects := []struct {
		name, desc, status string
		dueDate            string
	}{
		{"TaskFlow Backend", "Go CLI и TUI разработка", "ACTIVE", "2026-05-10"},
		{"Дизайн паттерны", "Курсовой проект по ООП", "ACTIVE", "2026-05-01"},
		{"Личные дела", "Бытовые задачи и напоминания", "ACTIVE", ""},
		{"Архив Q1", "Завершённые задачи первого квартала", "DONE", ""},
		{"Мобильное приложение", "React Native клиент для TaskFlow", "ACTIVE", "2026-05-20"},
	}
	for _, p := range projects {
		var res sql.Result
		if p.dueDate != "" {
			res, err = tx.Exec(
				`INSERT OR IGNORE INTO projects(user_id,name,description,status,due_date) VALUES(?,?,?,?,?)`,
				userID, p.name, p.desc, p.status, p.dueDate,
			)
		} else {
			res, err = tx.Exec(
				`INSERT OR IGNORE INTO projects(user_id,name,description,status) VALUES(?,?,?,?)`,
				userID, p.name, p.desc, p.status,
			)
		}
		if err != nil {
			return fmt.Errorf("seed project %q: %w", p.name, err)
		}
		id, _ := res.LastInsertId()
		if id == 0 {
			tx.QueryRow(`SELECT id FROM projects WHERE name=? AND user_id=?`, p.name, userID).Scan(&id)
		}
		projectIDs = append(projectIDs, id)
	}

	tagIDs := make([]int64, 0, 4)
	tags := []struct{ name, color string }{
		{"backend", "#4A90D9"},
		{"urgent", "#E74C3C"},
		{"docs", "#27AE60"},
		{"review", "#F39C12"},
	}
	for _, t := range tags {
		res, err := tx.Exec(
			`INSERT OR IGNORE INTO tags(user_id,name,color) VALUES(?,?,?)`,
			userID, t.name, t.color,
		)
		if err != nil {
			return fmt.Errorf("seed tag %q: %w", t.name, err)
		}
		id, _ := res.LastInsertId()
		if id == 0 {
			tx.QueryRow(`SELECT id FROM tags WHERE name=? AND user_id=?`, t.name, userID).Scan(&id)
		}
		tagIDs = append(tagIDs, id)
	}

	p := func(id int64) *int64 { return &id }
	// Задачи с SP для корректной оценки LLM.
	// Проект "Дизайн паттерны" (projectIDs[1]) — дедлайн 2026-05-01, много работы.
	// Проект "TaskFlow Backend" (projectIDs[0]) — дедлайн 2026-05-10, тоже нагружен.
	tasks := []struct {
		content     string
		status      string
		priority    string
		sp          int
		projectID   *int64
		tagID       *int64
		dueDate     string
	}{
		// Дизайн паттерны — просроченные и близкие дедлайны
		{"Реализовать паттерн Стратегия", "DONE", "HIGH", 3, p(projectIDs[1]), p(tagIDs[0]), ""},
		{"Реализовать паттерн Наблюдатель", "DONE", "HIGH", 3, p(projectIDs[1]), p(tagIDs[0]), ""},
		{"Реализовать паттерн Декоратор", "DONE", "MEDIUM", 2, p(projectIDs[1]), p(tagIDs[2]), ""},
		{"Реализовать паттерн Фабрика", "DONE", "MEDIUM", 2, p(projectIDs[1]), p(tagIDs[0]), ""},
		{"Написать BPMN диаграмму", "IN_PROGRESS", "HIGH", 5, p(projectIDs[1]), p(tagIDs[2]), "2026-04-28"},
		{"Написать ERD диаграмму", "IN_PROGRESS", "HIGH", 3, p(projectIDs[1]), p(tagIDs[2]), "2026-04-28"},
		{"Подготовить презентацию", "TODO", "HIGH", 8, p(projectIDs[1]), p(tagIDs[3]), "2026-04-30"},
		{"Написать отчёт по паттернам", "TODO", "HIGH", 5, p(projectIDs[1]), p(tagIDs[2]), "2026-04-29"},
		{"Покрыть код тестами", "TODO", "MEDIUM", 5, p(projectIDs[1]), p(tagIDs[0]), "2026-05-01"},
		{"Сделать код-ревью паттернов", "TODO", "MEDIUM", 3, p(projectIDs[1]), p(tagIDs[3]), "2026-05-01"},
		// TaskFlow Backend
		{"Добавить seed-данные в БД", "DONE", "MEDIUM", 2, p(projectIDs[0]), p(tagIDs[0]), ""},
		{"Настроить Telegram уведомления", "IN_PROGRESS", "HIGH", 5, p(projectIDs[0]), p(tagIDs[1]), "2026-05-05"},
		{"Написать README", "TODO", "LOW", 2, p(projectIDs[0]), p(tagIDs[2]), "2026-05-08"},
		{"Реализовать AI-советник", "TODO", "HIGH", 8, p(projectIDs[0]), p(tagIDs[0]), "2026-05-07"},
		{"Добавить экспорт задач в CSV", "TODO", "MEDIUM", 3, p(projectIDs[0]), p(tagIDs[0]), "2026-05-09"},
		{"Оптимизировать SQLite запросы", "TODO", "LOW", 3, p(projectIDs[0]), p(tagIDs[0]), "2026-05-10"},
		// Личные дела
		{"Записаться к врачу", "TODO", "HIGH", 1, p(projectIDs[2]), p(tagIDs[1]), "2026-04-30"},
		{"Оплатить счета", "TODO", "MEDIUM", 1, p(projectIDs[2]), nil, "2026-04-27"},
		{"Купить продукты", "DONE", "LOW", 1, p(projectIDs[2]), nil, ""},
		// Мобильное приложение
		{"Настроить React Native проект", "DONE", "HIGH", 3, p(projectIDs[4]), p(tagIDs[0]), ""},
		{"Реализовать экран списка задач", "IN_PROGRESS", "HIGH", 8, p(projectIDs[4]), p(tagIDs[0]), "2026-05-10"},
		{"Реализовать экран pomodoro", "TODO", "HIGH", 8, p(projectIDs[4]), p(tagIDs[0]), "2026-05-12"},
		{"Интеграция с API", "TODO", "HIGH", 13, p(projectIDs[4]), p(tagIDs[1]), "2026-05-15"},
		{"Тестирование на iOS/Android", "TODO", "MEDIUM", 5, p(projectIDs[4]), p(tagIDs[3]), "2026-05-18"},
	}
	for _, t := range tasks {
		pID := sql.NullInt64{}
		if t.projectID != nil {
			pID = sql.NullInt64{Int64: *t.projectID, Valid: true}
		}
		tID := sql.NullInt64{}
		if t.tagID != nil {
			tID = sql.NullInt64{Int64: *t.tagID, Valid: true}
		}
		dueDate := sql.NullString{}
		if t.dueDate != "" {
			dueDate = sql.NullString{String: t.dueDate, Valid: true}
		}
		_, err := tx.Exec(
			`INSERT INTO tasks(user_id,project_id,tag_id,content,status,priority,story_points,due_date) VALUES(?,?,?,?,?,?,?,?)`,
			userID, pID, tID, t.content, t.status, t.priority, t.sp, dueDate,
		)
		if err != nil {
			return fmt.Errorf("seed task %q: %w", t.content, err)
		}
	}

	notes := []struct {
		title, content string
		projectID      *int64
	}{
		{"Архитектура системы", "Проект использует Clean Architecture: domain, repository, service, TUI слои.", p(projectIDs[0])},
		{"Паттерны GoF", "Реализованы все 11 паттернов: Стратегия, Наблюдатель, Декоратор, Фабрика, Одиночка, Команда, Адаптер, Фасад, Шаблонный метод, Итератор, Компоновщик, Состояние, Заместитель.", p(projectIDs[1])},
		{"Формула ETA", "ETA = (remaining / avg_tasks_per_session) * avg_pomodoro_min", p(projectIDs[1])},
		{"Telegram Bot", "Токен передаётся через TELEGRAM_BOT_TOKEN env var. Chat ID через TELEGRAM_CHAT_ID.", p(projectIDs[0])},
	}
	for _, n := range notes {
		pID := sql.NullInt64{}
		if n.projectID != nil {
			pID = sql.NullInt64{Int64: *n.projectID, Valid: true}
		}
		_, err := tx.Exec(
			`INSERT INTO notes(user_id,project_id,title,content) VALUES(?,?,?,?)`,
			userID, pID, n.title, n.content,
		)
		if err != nil {
			return fmt.Errorf("seed note %q: %w", n.title, err)
		}
	}

	reminders := []struct {
		content, reminderTime string
		projectID             *int64
	}{
		{"Сдать курсовой проект", "2026-05-01 10:00:00", p(projectIDs[1])},
		{"Встреча с куратором", "2026-04-28 14:00:00", p(projectIDs[1])},
		{"Проверить Telegram уведомления", "2026-04-29 09:00:00", p(projectIDs[0])},
		{"Презентация проекта", "2026-05-01 12:00:00", p(projectIDs[1])},
	}
	for _, r := range reminders {
		pID := sql.NullInt64{}
		if r.projectID != nil {
			pID = sql.NullInt64{Int64: *r.projectID, Valid: true}
		}
		_, err := tx.Exec(
			`INSERT INTO reminders(user_id,project_id,content,reminder_time,status) VALUES(?,?,?,?,?)`,
			userID, pID, r.content, r.reminderTime, "PENDING",
		)
		if err != nil {
			return fmt.Errorf("seed reminder %q: %w", r.content, err)
		}
	}

	sessions := []struct {
		projectID  *int64
		duration   int
		state      string
		startTime  string
		finishTime string
		remaining  int
	}{
		// Дизайн паттерны — мало сессий, много работы => высокий риск просрочки
		{p(projectIDs[1]), 25, "COMPLETED", "2026-04-14 10:00:00", "2026-04-14 10:25:00", 0},
		{p(projectIDs[1]), 25, "COMPLETED", "2026-04-14 11:00:00", "2026-04-14 11:25:00", 0},
		{p(projectIDs[1]), 25, "CANCELLED", "2026-04-16 14:00:00", "", 1200},
		// TaskFlow Backend — хорошая история сессий
		{p(projectIDs[0]), 25, "COMPLETED", "2026-04-15 09:00:00", "2026-04-15 09:25:00", 0},
		{p(projectIDs[0]), 25, "COMPLETED", "2026-04-15 10:00:00", "2026-04-15 10:25:00", 0},
		{p(projectIDs[0]), 25, "COMPLETED", "2026-04-17 10:00:00", "2026-04-17 10:25:00", 0},
		{p(projectIDs[0]), 50, "COMPLETED", "2026-04-18 09:00:00", "2026-04-18 09:50:00", 0},
		{p(projectIDs[0]), 25, "CANCELLED", "2026-04-20 14:00:00", "", 1800},
		// Мобильное приложение — только началось
		{p(projectIDs[4]), 25, "COMPLETED", "2026-04-22 11:00:00", "2026-04-22 11:25:00", 0},
	}
	for _, s := range sessions {
		pID := sql.NullInt64{}
		if s.projectID != nil {
			pID = sql.NullInt64{Int64: *s.projectID, Valid: true}
		}
		finish := sql.NullString{}
		if s.finishTime != "" {
			finish = sql.NullString{String: s.finishTime, Valid: true}
		}
		_, err := tx.Exec(
			`INSERT INTO pomodoro_sessions(user_id,project_id,start_time,work_duration,finish_time,remaining_time,state) VALUES(?,?,?,?,?,?,?)`,
			userID, pID, s.startTime, s.duration, finish, s.remaining, s.state,
		)
		if err != nil {
			return fmt.Errorf("seed session: %w", err)
		}
	}

	return tx.Commit()
}
