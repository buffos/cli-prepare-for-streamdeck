package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mainMenuModel struct {
	titleStyle  lipgloss.Style
	promptStyle lipgloss.Style
	choices     []string
	cursor      int
	selectedCmd int
}

func initialMainMenuModel() mainMenuModel {
	return mainMenuModel{
		choices: []string{
			"Prepare Media Folder",
			"Echo Command",
			"Quit",
		},
		cursor:      0,
		selectedCmd: -1,
		titleStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true),
		promptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
	}
}

func (m mainMenuModel) Init() tea.Cmd {
	return nil
}

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			m.cursor = (m.cursor - 1 + len(m.choices)) % len(m.choices)

		case tea.KeyDown:
			m.cursor = (m.cursor + 1) % len(m.choices)

		case tea.KeyEnter:
			m.selectedCmd = m.cursor
			switch m.selectedCmd {
			case 0: // Prepare Media Folder
				return initialModel(), nil
			case 1: // Echo Command
				return initialEchoModel(), nil
			case 2: // Quit
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m mainMenuModel) View() string {
	s := m.titleStyle.Render("StreamDeck CLI Tool") + "\n\n"
	s += m.promptStyle.Render("Select a command:") + "\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	return s
}

func main() {
	p := tea.NewProgram(initialMainMenuModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
