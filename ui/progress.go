package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxWidth = 40
)

const (
	QuitReasonSuccess = iota
	QuitReasonError
	QuitReasonInterrupt
)

type ProgressWriter struct {
	total      int
	current    int
	onProgress func(float64)
}

var p *tea.Program

type incrMsg struct{}
type reqQuitMsg struct{}

type cancelFromOuterMsg struct{}

func (pw *ProgressWriter) Incr(val int) (int, error) {
	pw.current += val
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.current) / float64(pw.total))
	}
	return val, nil
}

var HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
var dotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#299E9C")).Render

type progressModel struct {
	progress progress.Model

	length  int
	current int

	padding int
	title   string
	clear   bool

	quitting   bool
	titleStyle lipgloss.Style
}

func Progress(title string, length int, updates chan bool, padding int, clear bool, headless bool) int {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#B24652"))

	if headless {
		for {
			if length == 0 {
				return QuitReasonSuccess
			}
			_, ok := <-updates
			if !ok {
				return QuitReasonSuccess
			}
			length--
		}
	} else {

		m := progressModel{
			progress:   progress.New(progress.WithGradient("#3A1F38", "#F18749")),
			length:     length,
			current:    0,
			padding:    padding,
			title:      title,
			clear:      clear,
			titleStyle: titleStyle,
		}

		wg := sync.WaitGroup{}

		p = tea.NewProgram(m)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if p == nil {
					continue
				}
				_, ok := <-updates
				if !ok {
					p.Send(cancelFromOuterMsg{})
					return
				}
				p.Send(incrMsg{})
			}
		}()

		if _, err := p.Run(); err != nil {
			fmt.Println("error running program:", err)
			os.Exit(1)
		}
		close(updates)

		wg.Wait()
		return QuitReasonSuccess
	}

}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case reqQuitMsg:
		m.quitting = true
		return m, tea.Quit

	case cancelFromOuterMsg:
		m.clear = true
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		m.clear = true
		m.quitting = true
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - m.padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case incrMsg:
		var cmds []tea.Cmd
		m.current++

		if m.current == m.length {
			cmds = append(cmds, tea.Sequence(finalPause(), func() tea.Msg {
				return reqQuitMsg{}
			}))
		}

		cmds = append(cmds, m.progress.SetPercent(float64(m.current)/float64(m.length)))
		return m, tea.Batch(cmds...)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m progressModel) View() string {
	if m.quitting {
		if m.clear {
			return ""
		} else {
			return dotStyle(m.title) + "\n"
		}
	}
	pad := strings.Repeat(" ", m.padding)
	return "\n" + m.titleStyle.Render(m.title) + " " +
		pad + m.progress.View() + "\n\n" +
		pad + HelpStyle("Press any key to quit")
}

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}
