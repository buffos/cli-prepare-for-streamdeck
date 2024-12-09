package main

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type echoModel struct {
	textInput   textinput.Model
	err         error
	titleStyle  lipgloss.Style
	promptStyle lipgloss.Style
	outputStyle lipgloss.Style
	output      string
	done        bool
}

func initialEchoModel() echoModel {
	ti := textinput.New()
	ti.Placeholder = "Enter text to echo"
	ti.Focus()

	return echoModel{
		textInput:   ti,
		titleStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true),
		promptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		outputStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")),
	}
}

func (m echoModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m echoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.done {
				return initialMainMenuModel(), nil
			}
			return m, tea.Quit

		case tea.KeyEnter:
			if m.done {
				return initialMainMenuModel(), nil
			}

			text := m.textInput.Value()
			if text == "" {
				m.err = errors.New("text cannot be empty")
				return m, nil
			}
			m.output = text
			m.done = true
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.titleStyle.Width(msg.Width)
	}

	if !m.done {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m echoModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(
			fmt.Sprintf("Error: %v", m.err),
		)
	}

	if m.done {
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			m.titleStyle.Render("Echo Command"),
			m.outputStyle.Render(m.output),
			m.promptStyle.Render("Press Enter to return to main menu"),
		)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.titleStyle.Render("Echo Command"),
		m.promptStyle.Render("Enter text to echo:"),
		m.textInput.View(),
	)
}
