package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/config"
	domain "taskflow/internal/domain"
)

// NotificationToggler — то, что Settings вызывает при переключении уведомлений.
// Реализуется ReminderCoordinator-ом (метод SetEnabled).
type NotificationToggler interface {
	SetEnabled(bool)
	IsEnabled() bool
	IsConfigured() bool
}

type settingsModel struct {
	svcs    Services
	user    *domain.User
	toggler NotificationToggler
	status  string
	width   int
	height  int
}

func newSettingsModel(svcs Services, user *domain.User, toggler NotificationToggler) settingsModel {
	return settingsModel{svcs: svcs, user: user, toggler: toggler}
}

func (m settingsModel) reload() tea.Cmd {
	return func() tea.Msg { return settingsLoadMsg{} }
}

type settingsLoadMsg struct{}

func (m settingsModel) Init() tea.Cmd { return nil }

// form у settings нет — но интерфейс activeHasForm ожидает его наличие у других экранов.
// Поле не нужно, но чтобы быть совместимыми с паттерном, оставляем nil.
// (В app.go проверка идёт явно по типу.)

func (m settingsModel) Update(msg tea.Msg) (settingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case settingsLoadMsg:
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "t", " ", "enter":
			return m.toggleNotifications()
		}
	}
	return m, nil
}

func (m settingsModel) toggleNotifications() (settingsModel, tea.Cmd) {
	cfg := config.Instance()
	cfg.NotificationsEnabled = !cfg.NotificationsEnabled
	if err := config.Save(cfg); err != nil {
		m.status = errorStyle.Render("Ошибка сохранения: " + err.Error())
		// откат
		cfg.NotificationsEnabled = !cfg.NotificationsEnabled
		return m, nil
	}
	if m.toggler != nil {
		m.toggler.SetEnabled(cfg.NotificationsEnabled)
	}
	if cfg.NotificationsEnabled {
		m.status = "✅ Telegram-уведомления включены."
	} else {
		m.status = "🔕 Telegram-уведомления выключены."
	}
	return m, nil
}

func (m settingsModel) View() string {
	cfg := config.Instance()

	tokenView := "(не задан)"
	if cfg.TelegramBotToken != "" {
		tokenView = maskToken(cfg.TelegramBotToken)
	}
	chatView := "(не задан)"
	if cfg.TelegramChatID != 0 {
		chatView = strconv.FormatInt(cfg.TelegramChatID, 10)
	}

	stateLabel := "🔕 ВЫКЛЮЧЕНЫ"
	stateStyle := lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
	if cfg.NotificationsEnabled {
		stateLabel = "🔔 ВКЛЮЧЕНЫ"
		stateStyle = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	}

	configured := m.toggler != nil && m.toggler.IsConfigured()
	configuredView := "❌ нет (укажите token и chat_id)"
	if configured {
		configuredView = "✅ да"
	}

	body := strings.Builder{}
	body.WriteString(titleStyle.Render("Settings — Уведомления") + "\n\n")
	body.WriteString("Telegram уведомления:  " + stateStyle.Render(stateLabel) + "\n")
	body.WriteString(fmt.Sprintf("Подключение настроено: %s\n\n", configuredView))
	body.WriteString(fmt.Sprintf("Bot token : %s\n", tokenView))
	body.WriteString(fmt.Sprintf("Chat ID   : %s\n", chatView))
	body.WriteString(fmt.Sprintf("Конфиг    : %s\n\n", config.Path()))
	body.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(
		"Token и chat_id задаются через ENV (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID)\n" +
			"либо вручную в файле конфига. ENV имеет приоритет.\n",
	))

	if m.status != "" {
		body.WriteString("\n" + m.status)
	}

	help := statusBarStyle.Render("t/space/enter toggle  tab switch  q quit")
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(body.String()),
		help,
	)
}

func (m settingsModel) setSize(w, h int) settingsModel {
	m.width, m.height = w, h
	return m
}

func maskToken(t string) string {
	if len(t) <= 8 {
		return strings.Repeat("*", len(t))
	}
	return t[:4] + strings.Repeat("*", len(t)-8) + t[len(t)-4:]
}
