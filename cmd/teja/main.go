package main

import (
	"log"
	"strings"

	"teja/internal/config"
	"teja/internal/ollama"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var userPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}).Align(lipgloss.Right).Padding(2).Width(80).Bold(true)

func main() {
	p := tea.NewProgram(ModelNew())
	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
	}
}

type model struct {
	Conversation []ollama.Message
	Cursor       int
	Input        textarea.Model
	rend         *glamour.TermRenderer
	Config       config.Config
}

func ModelNew() model {
	i := textarea.New()
	i.Focus()
	i.Prompt = "  "
	i.SetWidth(80)
	i.SetHeight(5)
	i.FocusedStyle.CursorLine = lipgloss.NewStyle()
	i.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"})
	i.ShowLineNumbers = false

	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		log.Fatalln(err)
	}

	return model{
		Conversation: []ollama.Message{},
		Cursor:       0,
		Input:        i,
		rend:         renderer,
		Config:       config.Read(),
	}

}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

type Stream struct {
	s  string
	ch <-chan string
}

func receiveOne(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		m, ok := <-ch
		if !ok {
			return Stream{s: m, ch: nil}
		}
		return Stream{s: m, ch: ch}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case Stream:
		if len(m.Conversation) == 0 {
			return m, cmd
		}
		m.Conversation[len(m.Conversation)-1].Content += msg.s
		if msg.ch != nil {
			return m, receiveOne(msg.ch)
		}
		return m, cmd

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.Cursor > 0 {
				m.Cursor -= 2
			}
			return m, cmd

		case tea.KeyDown:
			if m.Cursor+2 < len(m.Conversation) {
				m.Cursor += 2
			}
			return m, cmd

		case tea.KeyEnter:
			prompt := m.Input.Value()
			if strings.TrimSpace(prompt) == "" {
				return m, cmd
			}
			m.Input.Reset()
			m.Conversation = append(m.Conversation, ollama.Message{Role: "user", Content: prompt})
			m.Conversation = append(m.Conversation, ollama.Message{Role: "assistant", Content: ""})

			ch := ollama.Chat(m.Config, m.Conversation)

			m.Cursor = len(m.Conversation) - 2
			return m, receiveOne(ch)

		}

		switch msg.String() {
		case "ctrl+r":
			m.Conversation = []ollama.Message{}
			m.Cursor = 0
		}
	}

	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s string

	if len(m.Conversation) > 0 {
		s += userPromptStyle.Render(m.Conversation[m.Cursor].Content)

		md, err := m.rend.Render(m.Conversation[m.Cursor+1].Content)
		if err != nil {
			log.Fatalln(err)
		}

		s += md
	}

	s += m.Input.View()

	return s
}
