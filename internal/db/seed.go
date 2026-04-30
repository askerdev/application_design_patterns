package db

import (
	"database/sql"
	"fmt"
	"time"
)

// RunSeed заполняет БД демонстрационными данными.
// Все даты задач/проектов/напоминаний рассчитываются ОТНОСИТЕЛЬНО текущего момента,
// чтобы диаграмма Ганта и LLM-планировщик всегда показывали свежую и наглядную картину
// (HIGH-задачи в начале, MEDIUM/LOW дальше, у некоторых проектов — реалистичный риск просрочки).
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

	now := time.Now()
	// daysFromNow возвращает дату в формате YYYY-MM-DD, сдвинутую на n дней от сегодня.
	daysFromNow := func(n int) string {
		return now.AddDate(0, 0, n).Format("2006-01-02")
	}
	// dateTimeFromNow — то же, но с временем (для reminders / pomodoro_sessions).
	dateTimeFromNow := func(daysOffset int, hour, minute int) string {
		t := now.AddDate(0, 0, daysOffset)
		t = time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
		return t.Format("2006-01-02 15:04:05")
	}

	projectIDs := make([]int64, 0, 7)
	projects := []struct {
		name, desc, status string
		dueDate            string // относительно now
	}{
		// 🌟 ВИТРИННЫЙ ПРОЕКТ: спроектирован под красивые графики
		// (разный SP, плотная загрузка по дням, ненулевая просрочка, статусы перемешаны)
		{"🚀 Запуск SaaS MVP", "Витринный проект для демо: запуск продукта за 30 дней", "ACTIVE", daysFromNow(30)},
		// Активный плотный проект — близкий дедлайн, много работы (риск просрочки)
		{"Дизайн паттерны", "Курсовой проект по ООП — критичный дедлайн", "ACTIVE", daysFromNow(7)},
		// Активный средний проект — комфортный дедлайн
		{"TaskFlow Backend", "Go CLI и TUI разработка", "ACTIVE", daysFromNow(21)},
		// Активный длинный проект — далёкий горизонт
		{"Мобильное приложение", "React Native клиент для TaskFlow", "ACTIVE", daysFromNow(45)},
		// Личные дела — без дедлайна, мелкие задачи
		{"Личные дела", "Бытовые задачи и напоминания", "ACTIVE", ""},
		// Архив — для контраста, не должен фигурировать в диаграмме (статус DONE)
		{"Архив Q1", "Завершённые задачи прошлого квартала", "DONE", ""},
		// Стартап-исследование — длинные epics, большой story_points
		{"R&D: AI Assistant", "Исследовательский трек: интеграция LLM в продукт", "ACTIVE", daysFromNow(60)},
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
	type taskSeed struct {
		content   string
		status    string
		priority  string
		sp        int
		projectID *int64
		tagID     *int64
		dueDate   string
	}

	// Алиасы для читаемости
	pSaaS := projectIDs[0]     // 🌟 витринный проект (дедлайн +30)
	pPatterns := projectIDs[1] // дедлайн +7 дней
	pBackend := projectIDs[2]  // дедлайн +21
	pMobile := projectIDs[3]   // дедлайн +45
	pPersonal := projectIDs[4] // без дедлайна
	pArchive := projectIDs[5]  // DONE
	pRD := projectIDs[6]       // дедлайн +60

	tasks := []taskSeed{
		// === 🚀 ЗАПУСК SAAS MVP — ВИТРИННЫЙ ПРОЕКТ для демо ========================
		// 18 задач, разнообразие SP (1, 2, 3, 5, 8, 13), смесь статусов,
		// плотная загрузка в первые две недели, несколько просрочек.
		// Discovery / planning — уже сделано
		{"Customer interviews (5 шт)", "DONE", "HIGH", 5, p(pSaaS), p(tagIDs[2]), ""},
		{"Конкурентный анализ", "DONE", "MEDIUM", 3, p(pSaaS), p(tagIDs[2]), ""},
		{"Сформулировать value proposition", "DONE", "HIGH", 2, p(pSaaS), p(tagIDs[2]), ""},
		// Design — частично сделано, частично в работе
		{"UX-исследование пути пользователя", "DONE", "HIGH", 3, p(pSaaS), p(tagIDs[3]), ""},
		{"Дизайн-макеты в Figma (8 экранов)", "IN_PROGRESS", "HIGH", 8, p(pSaaS), p(tagIDs[3]), daysFromNow(3)},
		{"Design system: токены и компоненты", "IN_PROGRESS", "MEDIUM", 5, p(pSaaS), p(tagIDs[3]), daysFromNow(5)},
		// Backend — большая часть впереди
		{"Схема БД и миграции", "DONE", "HIGH", 3, p(pSaaS), p(tagIDs[0]), ""},
		{"Auth: регистрация / login / JWT", "IN_PROGRESS", "HIGH", 8, p(pSaaS), p(tagIDs[0]), daysFromNow(-2)}, // OVERDUE
		{"REST API: основные ресурсы", "TODO", "HIGH", 13, p(pSaaS), p(tagIDs[0]), daysFromNow(10)},
		{"Биллинг: интеграция Stripe", "TODO", "HIGH", 8, p(pSaaS), p(tagIDs[0]), daysFromNow(18)},
		{"Email-уведомления (transactional)", "TODO", "MEDIUM", 3, p(pSaaS), p(tagIDs[0]), daysFromNow(20)},
		// Frontend
		{"Setup Next.js + TypeScript", "DONE", "MEDIUM", 2, p(pSaaS), p(tagIDs[0]), ""},
		{"Главная страница (landing)", "TODO", "MEDIUM", 5, p(pSaaS), p(tagIDs[0]), daysFromNow(12)},
		{"Дашборд пользователя", "TODO", "HIGH", 8, p(pSaaS), p(tagIDs[0]), daysFromNow(16)},
		{"Settings & profile", "TODO", "LOW", 3, p(pSaaS), p(tagIDs[0]), daysFromNow(22)},
		// QA / Launch
		{"E2E тесты основных флоу", "TODO", "MEDIUM", 5, p(pSaaS), p(tagIDs[3]), daysFromNow(25)},
		{"Performance audit + Lighthouse", "TODO", "MEDIUM", 2, p(pSaaS), p(tagIDs[3]), daysFromNow(27)},
		{"Деплой на prod (k8s + CDN)", "TODO", "HIGH", 5, p(pSaaS), p(tagIDs[1]), daysFromNow(28)},
		{"Маркетинг-сайт + ProductHunt запуск", "TODO", "HIGH", 5, p(pSaaS), p(tagIDs[2]), daysFromNow(30)},

		// === Дизайн паттерны (горящий проект, дедлайн +7) ==========================
		// История: 4 паттерна уже сделаны
		{"Реализовать паттерн Стратегия", "DONE", "HIGH", 3, p(pPatterns), p(tagIDs[0]), ""},
		{"Реализовать паттерн Наблюдатель", "DONE", "HIGH", 3, p(pPatterns), p(tagIDs[0]), ""},
		{"Реализовать паттерн Декоратор", "DONE", "MEDIUM", 2, p(pPatterns), p(tagIDs[2]), ""},
		{"Реализовать паттерн Фабрика", "DONE", "MEDIUM", 2, p(pPatterns), p(tagIDs[0]), ""},
		// Просроченные / в работе — диаграмма должна это показать
		{"Написать BPMN диаграмму", "IN_PROGRESS", "HIGH", 5, p(pPatterns), p(tagIDs[2]), daysFromNow(-1)},
		{"Написать ERD диаграмму", "IN_PROGRESS", "HIGH", 3, p(pPatterns), p(tagIDs[2]), daysFromNow(0)},
		// Срочное — высокий приоритет, разные SP
		{"Подготовить презентацию", "TODO", "HIGH", 8, p(pPatterns), p(tagIDs[3]), daysFromNow(5)},
		{"Написать отчёт по паттернам", "TODO", "HIGH", 5, p(pPatterns), p(tagIDs[2]), daysFromNow(4)},
		{"Покрыть код тестами", "TODO", "MEDIUM", 5, p(pPatterns), p(tagIDs[0]), daysFromNow(6)},
		{"Сделать код-ревью паттернов", "TODO", "MEDIUM", 3, p(pPatterns), p(tagIDs[3]), daysFromNow(6)},

		// === TaskFlow Backend (комфортный дедлайн +21) =============================
		{"Добавить seed-данные в БД", "DONE", "MEDIUM", 2, p(pBackend), p(tagIDs[0]), ""},
		{"Настроить Telegram уведомления", "IN_PROGRESS", "HIGH", 5, p(pBackend), p(tagIDs[1]), daysFromNow(3)},
		{"Реализовать AI-советник", "TODO", "HIGH", 8, p(pBackend), p(tagIDs[0]), daysFromNow(7)},
		{"Добавить экспорт задач в CSV", "TODO", "MEDIUM", 3, p(pBackend), p(tagIDs[0]), daysFromNow(12)},
		{"Оптимизировать SQLite запросы", "TODO", "LOW", 3, p(pBackend), p(tagIDs[0]), daysFromNow(15)},
		{"Написать README", "TODO", "LOW", 2, p(pBackend), p(tagIDs[2]), daysFromNow(18)},
		{"Реализовать веб-диаграмму Ганта", "TODO", "MEDIUM", 5, p(pBackend), p(tagIDs[0]), daysFromNow(10)},

		// === Мобильное приложение (длинный горизонт +45) ===========================
		{"Настроить React Native проект", "DONE", "HIGH", 3, p(pMobile), p(tagIDs[0]), ""},
		{"Реализовать экран списка задач", "IN_PROGRESS", "HIGH", 8, p(pMobile), p(tagIDs[0]), daysFromNow(10)},
		{"Реализовать экран pomodoro", "TODO", "HIGH", 8, p(pMobile), p(tagIDs[0]), daysFromNow(15)},
		{"Интеграция с REST API", "TODO", "HIGH", 13, p(pMobile), p(tagIDs[1]), daysFromNow(25)},
		{"Push-уведомления (FCM)", "TODO", "MEDIUM", 5, p(pMobile), p(tagIDs[0]), daysFromNow(30)},
		{"Тестирование на iOS/Android", "TODO", "MEDIUM", 5, p(pMobile), p(tagIDs[3]), daysFromNow(38)},
		{"Подготовить App Store метаданные", "TODO", "LOW", 3, p(pMobile), p(tagIDs[2]), daysFromNow(42)},

		// === Личные дела (без дедлайна проекта, мелкие задачи) =====================
		{"Записаться к врачу", "TODO", "HIGH", 1, p(pPersonal), p(tagIDs[1]), daysFromNow(2)},
		{"Оплатить счета", "TODO", "MEDIUM", 1, p(pPersonal), nil, daysFromNow(-2)}, // OVERDUE
		{"Купить продукты", "DONE", "LOW", 1, p(pPersonal), nil, ""},
		{"Записаться в спортзал", "TODO", "LOW", 1, p(pPersonal), nil, daysFromNow(7)},

		// === Архив Q1 (DONE проект — задачи тоже DONE) =============================
		{"Подключить SQLite", "DONE", "HIGH", 5, p(pArchive), p(tagIDs[0]), ""},
		{"Базовая модель Task", "DONE", "MEDIUM", 3, p(pArchive), p(tagIDs[0]), ""},
		{"Первый прототип TUI", "DONE", "HIGH", 8, p(pArchive), p(tagIDs[0]), ""},

		// === R&D: AI Assistant (длинные эпики +60) =================================
		{"Сравнить локальные LLM (gemma/llama/qwen)", "IN_PROGRESS", "HIGH", 5, p(pRD), p(tagIDs[2]), daysFromNow(7)},
		{"Бенчмарк latency на M1/M2", "TODO", "MEDIUM", 3, p(pRD), p(tagIDs[2]), daysFromNow(14)},
		{"Спроектировать prompt-pipeline", "TODO", "HIGH", 8, p(pRD), p(tagIDs[0]), daysFromNow(20)},
		{"Реализовать function calling", "TODO", "HIGH", 13, p(pRD), p(tagIDs[0]), daysFromNow(35)},
		{"Eval-set: 50 кейсов планирования", "TODO", "MEDIUM", 8, p(pRD), p(tagIDs[3]), daysFromNow(45)},
		{"Презентация результатов R&D", "TODO", "MEDIUM", 5, p(pRD), p(tagIDs[3]), daysFromNow(55)},
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
		{"🚀 SaaS MVP: roadmap", "30 дней до запуска. Discovery ✓, Design в работе, Backend начат, Frontend следом, QA + Launch на финише.", p(pSaaS)},
		{"🚀 SaaS MVP: метрики успеха", "Цели запуска: 100 регистраций в первую неделю, 10 платных подписок до конца месяца, NPS ≥ 40 от beta-юзеров.", p(pSaaS)},
		{"Архитектура системы", "Проект использует Clean Architecture: domain, repository, service, TUI слои.", p(pBackend)},
		{"Паттерны GoF", "Реализованы 11 паттернов: Стратегия, Наблюдатель, Декоратор, Фабрика, Одиночка, Команда, Адаптер, Фасад, Шаблонный метод, Итератор, Компоновщик, Состояние, Заместитель.", p(pPatterns)},
		{"Формула ETA", "ETA = (remaining / avg_tasks_per_session) * avg_pomodoro_min", p(pPatterns)},
		{"Telegram Bot", "Токен передаётся через TELEGRAM_BOT_TOKEN env var. Chat ID через TELEGRAM_CHAT_ID. Тоггл уведомлений — на вкладке Settings.", p(pBackend)},
		{"Gantt-планировщик", "Использует LLM (Ollama) для расстановки задач по календарю. Если LLM недоступна — детерминированный fallback по приоритету и SP.", p(pBackend)},
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
		content      string
		daysFromNow  int
		hour, minute int
		projectID    *int64
	}{
		{"Сдать курсовой проект", 7, 10, 0, p(pPatterns)},
		{"Встреча с куратором", 3, 14, 0, p(pPatterns)},
		{"Презентация проекта", 5, 12, 0, p(pPatterns)},
		{"Демо TaskFlow для команды", 14, 11, 0, p(pBackend)},
		{"Sync по мобильному приложению", 21, 15, 0, p(pMobile)},
		// SaaS MVP
		{"Investor pitch session", 14, 10, 0, p(pSaaS)},
		{"Beta-тест с 5 клиентами", 24, 16, 0, p(pSaaS)},
		{"ProductHunt запуск 🚀", 30, 9, 0, p(pSaaS)},
	}
	for _, r := range reminders {
		pID := sql.NullInt64{}
		if r.projectID != nil {
			pID = sql.NullInt64{Int64: *r.projectID, Valid: true}
		}
		_, err := tx.Exec(
			`INSERT INTO reminders(user_id,project_id,content,reminder_time,status) VALUES(?,?,?,?,?)`,
			userID, pID, r.content, dateTimeFromNow(r.daysFromNow, r.hour, r.minute), "PENDING",
		)
		if err != nil {
			return fmt.Errorf("seed reminder %q: %w", r.content, err)
		}
	}

	sessions := []struct {
		projectID         *int64
		duration          int
		state             string
		startDayOffset    int
		startHour, startM int
		finishDayOffset   int // -1 если не завершалось
		finishHour, finM  int
		remaining         int
	}{
		// Дизайн паттерны — мало сессий, много работы => высокий риск
		{p(pPatterns), 25, "COMPLETED", -16, 10, 0, -16, 10, 25, 0},
		{p(pPatterns), 25, "COMPLETED", -16, 11, 0, -16, 11, 25, 0},
		{p(pPatterns), 25, "CANCELLED", -14, 14, 0, -1, 0, 0, 1200},
		// TaskFlow Backend — хорошая история сессий
		{p(pBackend), 25, "COMPLETED", -15, 9, 0, -15, 9, 25, 0},
		{p(pBackend), 25, "COMPLETED", -15, 10, 0, -15, 10, 25, 0},
		{p(pBackend), 25, "COMPLETED", -13, 10, 0, -13, 10, 25, 0},
		{p(pBackend), 50, "COMPLETED", -12, 9, 0, -12, 9, 50, 0},
		{p(pBackend), 25, "CANCELLED", -10, 14, 0, -1, 0, 0, 1800},
		// Мобильное приложение — только началось
		{p(pMobile), 25, "COMPLETED", -8, 11, 0, -8, 11, 25, 0},
		{p(pMobile), 25, "COMPLETED", -3, 16, 0, -3, 16, 25, 0},
		// R&D — экспериментальные сессии
		{p(pRD), 50, "COMPLETED", -5, 14, 0, -5, 14, 50, 0},
		// SaaS MVP — продуктивная команда
		{p(pSaaS), 50, "COMPLETED", -7, 10, 0, -7, 10, 50, 0},
		{p(pSaaS), 50, "COMPLETED", -7, 14, 0, -7, 14, 50, 0},
		{p(pSaaS), 50, "COMPLETED", -5, 9, 0, -5, 9, 50, 0},
		{p(pSaaS), 25, "COMPLETED", -4, 11, 0, -4, 11, 25, 0},
		{p(pSaaS), 50, "COMPLETED", -3, 14, 0, -3, 14, 50, 0},
		{p(pSaaS), 25, "COMPLETED", -2, 10, 0, -2, 10, 25, 0},
		{p(pSaaS), 50, "COMPLETED", -1, 14, 0, -1, 14, 50, 0},
	}
	for _, s := range sessions {
		pID := sql.NullInt64{}
		if s.projectID != nil {
			pID = sql.NullInt64{Int64: *s.projectID, Valid: true}
		}
		start := dateTimeFromNow(s.startDayOffset, s.startHour, s.startM)
		finish := sql.NullString{}
		if s.finishDayOffset != -1 {
			finish = sql.NullString{
				String: dateTimeFromNow(s.finishDayOffset, s.finishHour, s.finM),
				Valid:  true,
			}
		}
		_, err := tx.Exec(
			`INSERT INTO pomodoro_sessions(user_id,project_id,start_time,work_duration,finish_time,remaining_time,state) VALUES(?,?,?,?,?,?,?)`,
			userID, pID, start, s.duration, finish, s.remaining, s.state,
		)
		if err != nil {
			return fmt.Errorf("seed session: %w", err)
		}
	}

	return tx.Commit()
}
