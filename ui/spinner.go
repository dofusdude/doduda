package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error

	incomingUpdates chan string
	lastUpdate      string
	title           string
	newline         bool
}

var TitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#B24652"))
var DotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#299E9C"))

func Spinner(title string, feedbacks chan string, newline bool, headless bool) {
	if headless {
		for feedback := range feedbacks {
			log.Print(TitleStyle.Render(title), "step", feedback)
		}
	} else {
		p := tea.NewProgram(initialModel(title, feedbacks, newline))
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func initialModel(title string, feedbacks chan string, newline bool) model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = DotStyle
	return model{spinner: s, incomingUpdates: feedbacks, title: title, newline: newline}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	default:
		select {
		case update, ok := <-m.incomingUpdates:
			if !ok {
				m.quitting = true
				return m, tea.Quit
			}
			m.lastUpdate = update
		default:
			// do nothing
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	var prefix string
	if m.newline {
		prefix = "\n"
	}
	str := fmt.Sprintf("%s%s %s %s\n", prefix, TitleStyle.Render(m.title), m.spinner.View(), m.lastUpdate)
	if m.quitting {
		return ""
	}
	return str
}
