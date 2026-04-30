package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	domain "taskflow/internal/domain"
)

// GanttServer — singleton HTTP сервер, отдающий страницу с диаграммой Ганта.
// Запускается лениво при первом обращении из TUI; повторные старты — no-op.
type GanttServer struct {
	mu       sync.Mutex
	planner  domain.GanttPlanner
	projects domain.ProjectService

	addr    string // настроенный адрес (например ":8080")
	bound   string // фактический адрес после Listen (например "[::]:8080")
	running bool
	userID  int64

	// Кэш последнего сгенерированного плана (по projectID).
	plansMu sync.RWMutex
	plans   map[int64]*domain.GanttPlan
}

func NewGanttServer(planner domain.GanttPlanner, projects domain.ProjectService, addr string, userID int64) *GanttServer {
	if addr == "" {
		addr = ":8080"
	}
	return &GanttServer{
		planner:  planner,
		projects: projects,
		addr:     addr,
		userID:   userID,
		plans:    make(map[int64]*domain.GanttPlan),
	}
}

// extractPort возвращает порт из адреса вида ":8080" / "[::]:8080" / "localhost:8080".
func extractPort(addr string) string {
	if _, port, err := net.SplitHostPort(addr); err == nil {
		return port
	}
	return ""
}

// Addr возвращает человекочитаемый адрес сервера для отображения в UI.
func (s *GanttServer) Addr() string {
	port := extractPort(s.bound)
	if port == "" {
		port = extractPort(s.addr)
	}
	if port == "" {
		return "localhost"
	}
	return "localhost:" + port
}

// URL возвращает URL для проекта (для открытия в браузере).
func (s *GanttServer) URL(projectID int64) string {
	port := extractPort(s.bound)
	if port == "" {
		port = extractPort(s.addr)
	}
	if port == "" {
		port = "8080"
	}
	return fmt.Sprintf("http://localhost:%s/?project_id=%d", port, projectID)
}

// IsRunning возвращает, запущен ли сервер.
func (s *GanttServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// SetPlan кэширует план, чтобы web-страница могла его получить (без повторного дёргания LLM).
func (s *GanttServer) SetPlan(plan *domain.GanttPlan) {
	if plan == nil || plan.Project == nil {
		return
	}
	s.plansMu.Lock()
	s.plans[plan.Project.ID] = plan
	s.plansMu.Unlock()
}

func (s *GanttServer) getPlan(projectID int64) *domain.GanttPlan {
	s.plansMu.RLock()
	defer s.plansMu.RUnlock()
	return s.plans[projectID]
}

// Start запускает HTTP-сервер в горутине. Повторный вызов — no-op.
func (s *GanttServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/gantt", s.handleAPIGantt)

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.bound = ln.Addr().String()
	s.running = true
	s.mu.Unlock()

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		_ = srv.Serve(ln)
	}()
	return nil
}

func (s *GanttServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("project_id")
	pid, _ := strconv.ParseInt(pidStr, 10, 64)

	tmpl, err := template.New("gantt").Parse(ganttHTML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, _ := s.projects.List(s.userID)
	// Если userID не задан / пользователь ещё не создан — возьмём проекты из кэша планов.
	if len(projects) == 0 {
		s.plansMu.RLock()
		for _, p := range s.plans {
			if p.Project != nil {
				projects = append(projects, p.Project)
			}
		}
		s.plansMu.RUnlock()
	}

	type projOpt struct {
		ID   int64
		Name string
	}
	opts := make([]projOpt, len(projects))
	for i, p := range projects {
		opts[i] = projOpt{ID: p.ID, Name: p.Name}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, map[string]any{
		"Projects":  opts,
		"ProjectID": pid,
	})
}

func (s *GanttServer) handleAPIGantt(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("project_id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || pid <= 0 {
		http.Error(w, "invalid project_id", http.StatusBadRequest)
		return
	}

	// Сначала пробуем кэш (положенный TUI после генерации).
	plan := s.getPlan(pid)
	if plan == nil {
		// Иначе генерируем здесь же.
		p, perr := s.planner.Plan(pid)
		if perr != nil {
			http.Error(w, perr.Error(), http.StatusInternalServerError)
			return
		}
		plan = p
		s.SetPlan(plan)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(plan)
}

// OpenBrowser кросс-платформенно открывает URL в браузере по умолчанию.
func OpenBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
