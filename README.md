# TaskFlow

Менеджер задач с TUI-интерфейсом. Go + SQLite + Bubble Tea.

## Запуск

```bash
go build -o taskflow ./cmd/taskflow
./taskflow              # TUI
./taskflow seed         # заполнить БД демо-данными (31 запись)
./taskflow task add "Задача" --priority HIGH
./taskflow task list
./taskflow project add "Проект"
./taskflow project list
```

## Структура проекта

```
cmd/taskflow/           — точка входа, Cobra CLI
internal/
  app/                  — Facade
  config/               — Singleton (конфигурация)
  db/                   — Singleton (соединение с БД), миграции, seed
  domain/               — доменные объекты и паттерны
  notifications/        — Observer, Adapter (Telegram)
  pomodoro/             — State machine
  repository/sqlite/    — репозитории SQLite
  tui/                  — Bubble Tea TUI
docs/diagrams/          — BPMN, ERD, диаграммы классов и состояний
```

---

## Паттерны проектирования

### Порождающие (Creational)

| Паттерн | Файл | Описание |
|---------|------|----------|
| **Singleton** | `internal/config/config.go` | `config.Instance()` — один экземпляр конфигурации через `sync.Once` |
| **Singleton** | `internal/db/db.go` | `db.Instance()` — одно соединение с SQLite через `sync.Once` |
| **Factory** | `internal/domain/factory.go` | `EntityFactory` создаёт Task, Project, Note, Reminder, Tag с дефолтными значениями |

### Структурные (Structural)

| Паттерн | Файл | Описание |
|---------|------|----------|
| **Facade** | `internal/app/facade.go` | `AppFacade` — единая точка входа для CLI-команд (AddTask, ListTasks, AddProject, ListProjects) |
| **Proxy** | `internal/domain/proxy_cache.go` | `CachingTaskRepo` — прозрачное кэширование поверх `TaskRepository` |
| **Decorator** | `internal/domain/decorator_display.go` | `PriorityDecorator`, `OverdueDecorator` оборачивают `Renderer` для форматирования задач в TUI |
| **Composite** | `internal/domain/composite.go` | `WorkItem` — дерево: `ProjectNode` содержит `TaskLeaf` и `NoteLeaf`; используется в деталях проекта |
| **Adapter** | `internal/notifications/telegram/adapter.go` | `ClientAdapter` адаптирует `Client` (Telegram HTTP) к интерфейсу `MessageSender` |

### Поведенческие (Behavioral)

| Паттерн | Файл | Описание |
|---------|------|----------|
| **Strategy** | `internal/domain/strategy_filter.go` | `FilterStrategy` + `ByStatusFilter`, `ByTagFilter`, `ByDateFilter` — взаимозаменяемые алгоритмы фильтрации задач |
| **Observer** | `internal/notifications/telegram/notifier.go` | `ReminderCoordinator` уведомляет `TelegramNotifier` при наступлении времени напоминания |
| **Command** | `internal/domain/command_task.go` | `CompleteTaskCommand`, `DeleteTaskCommand` + `CommandHistory` — undo в TUI (клавиша `u`) |
| **Template Method** | `internal/domain/template_report.go` | `Generate(r Report)` — шаблонный метод; конкретные отчёты `TaskCountReport`, `ProjectETAReport` реализуют `fetchData()` и `formatContent()` |
| **Iterator** | `internal/domain/iterator.go` | `SliceIterator[T]` — обход коллекции задач без раскрытия внутренней структуры |
| **State** | `internal/pomodoro/machine.go` | `PomodoroMachine` — состояния IDLE/RUNNING/PAUSED/COMPLETED/CANCELLED; переходы через `Start()`, `Pause()`, `Resume()`, `Complete()`, `Cancel()` |

---

## Требования этапа 1 (выделение элементов)

### Стратегии поведения объектов

Интерфейс `FilterStrategy` (`internal/domain/strategy_filter.go`):

```
FilterStrategy
  ├── ByStatusFilter   — фильтр по статусу задачи (TODO / DONE)
  ├── ByTagFilter      — фильтр по тегу
  └── ByDateFilter     — фильтр по дате дедлайна (просроченные)
```

Стратегия подставляется в `TaskService.Filter(strategy)` без изменения остального кода.

### Ключевой объект с состояниями

`PomodoroSession` / `PomodoroMachine` (`internal/pomodoro/machine.go`, `internal/domain/pomodoro_entity.go`):

```
IDLE ──Start()──► RUNNING ──Pause()──► PAUSED
                     │                   │
                  Cancel()            Resume() ──► RUNNING
                     │                Cancel()
                  Complete()              │
                     │                   ▼
                     ▼               CANCELLED
                 COMPLETED
```

COMPLETED и CANCELLED — терминальные состояния. Любой переход из них возвращает ошибку.

### Синглетные классы

| Класс | Файл | Гарантия единственности |
|-------|------|------------------------|
| `Config` | `internal/config/config.go` | `sync.Once`, `Instance()` — единственный экземпляр на процесс |
| `*sql.DB` | `internal/db/db.go` | `sync.Once`, `Instance()` — единственное соединение с БД |

---

## Математическая модель

**ETA проекта** (`internal/domain/stats_eta.go`):

```
ETA (мин) = (оставшиеся задачи / среднее задач за сессию) × средняя длина сессии
```

Используется в `ProjectETAReport` для оценки времени завершения проекта на основе истории Pomodoro-сессий.

---

## Диаграммы

| Файл | Содержание |
|------|-----------|
| `docs/diagrams/bp1_reminder_notification.bpmn` | БП1: Жизненный цикл напоминания (Пользователь → TaskFlow → Telegram) |
| `docs/diagrams/bp2_pomodoro_session.bpmn` | БП2: Жизненный цикл Pomodoro-сессии (старт/пауза/отмена/завершение) |
| `docs/diagrams/bp3_template_method.bpmn` | БП3: Шаблонный метод — Generate() делегирует fetchData/formatContent субклассам |
| `docs/diagrams/erd.drawio` | ERD: 7 сущностей с атрибутами и внешними ключами |
| `docs/diagrams/state_machine.drawio` | Машина состояний: PomodoroSession |
| `docs/diagrams/class_diagram.drawio` | Диаграмма классов: все 11 паттернов |

---

## БД

SQLite, 7 таблиц: `users`, `projects`, `tags`, `tasks`, `notes`, `reminders`, `pomodoro_sessions`.

```bash
./taskflow seed   # вставляет 31 запись (4 проекта, 4 тега, 10 задач, 4 заметки, 4 напоминания, 5 сессий)
```
