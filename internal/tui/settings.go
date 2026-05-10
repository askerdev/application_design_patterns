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
	cursor  int
	width   int
	height  int
}

func newSettingsModel(svcs Services, user *domain.User, toggler NotificationToggler) settingsModel {
	return settingsModel{svcs: svcs, user: user, toggler: toggler, cursor: 0}
}

func (m settingsModel) reload() tea.Cmd {
	return func() tea.Msg { return settingsLoadMsg{} }
}

type settingsLoadMsg struct{}

func (m settingsModel) Init() tea.Cmd { return nil }

func (m settingsModel) Update(msg tea.Msg) (settingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case settingsLoadMsg:
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}
		case " ", "enter":
			if m.cursor == 0 {
				return m.toggleNotifications()
			} else {
				return m.toggleTheme()
			}
		}
	}
	return m, nil
}

func (m settingsModel) toggleNotifications() (settingsModel, tea.Cmd) {
	cfg := config.Instance()
	cfg.NotificationsEnabled = !cfg.NotificationsEnabled
	if err := config.Save(cfg); err != nil {
		m.status = errorStyle.Render("Ошибка сохранения: " + err.Error())
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

func (m settingsModel) toggleTheme() (settingsModel, tea.Cmd) {
	cfg := config.Instance()
	if cfg.Theme == "dark" {
		cfg.Theme = "light"
	} else {
		cfg.Theme = "dark"
	}
	if err := config.Save(cfg); err != nil {
		m.status = errorStyle.Render("Ошибка сохранения: " + err.Error())
		if cfg.Theme == "dark" {
			cfg.Theme = "light"
		} else {
			cfg.Theme = "dark"
		}
		return m, nil
	}
	lipgloss.SetHasDarkBackground(cfg.Theme == "dark")
	applyTerminalTheme(cfg.Theme == "dark")
	if cfg.Theme == "dark" {
		m.status = "🌙 Тёмная тема включена."
	} else {
		m.status = "☀️ Светлая тема включена."
	}
	return m, tea.WindowSize()
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

	themeLabel := "🌙 ТЁМНАЯ"
	themeStyle := lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
	if cfg.Theme == "light" {
		themeLabel = "☀️ СВЕТЛАЯ"
		themeStyle = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	}

	configured := m.toggler != nil && m.toggler.IsConfigured()
	configuredView := "❌ нет (укажите token и chat_id)"
	if configured {
		configuredView = "✅ да"
	}

	cursorNotif := "  "
	cursorTheme := "  "
	if m.cursor == 0 {
		cursorNotif = "> "
	} else if m.cursor == 1 {
		cursorTheme = "> "
	}

	body := strings.Builder{}
	body.WriteString(titleStyle.Render("Settings — Настройки") + "\n\n")
	body.WriteString(cursorNotif + "Telegram уведомления:  " + stateStyle.Render(stateLabel) + "\n")
	body.WriteString(cursorTheme + "Тема интерфейса:       " + themeStyle.Render(themeLabel) + "\n\n")
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

	help := statusBarStyle.Render("↑/↓ move  space/enter toggle  tab switch  q quit")
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