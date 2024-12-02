package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/buffos/cli-prepare-for-streamdeck/config"
)

// DirectoryInfo provides more details about a directory
type DirectoryInfo struct {
	Name      string
	Path      string
	Size      int64
	FileCount int
	IsDir     bool
}

// listDirectoriesWithDetails returns detailed information about directories in the given path
func listDirectoriesWithDetails(path string) ([]DirectoryInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirInfos []DirectoryInfo
	// Add parent directory option
	if path != "/" && path != filepath.VolumeName(path)+"\\"+string(filepath.Separator) {
		parentPath := filepath.Dir(path)
		dirInfos = append(dirInfos, DirectoryInfo{
			Name:  "..",
			Path:  parentPath,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		dirInfo := DirectoryInfo{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			// Count files and subdirectories
			subEntries, err := os.ReadDir(fullPath)
			if err == nil {
				dirInfo.FileCount = len(subEntries)
			}
		} else {
			dirInfo.Size = info.Size()
		}

		dirInfos = append(dirInfos, dirInfo)
	}

	return dirInfos, nil
}

type model struct {
	// Current step in the process
	step int

	// Input fields
	pathInput    textinput.Model
	mediaTypeIdx int
	oscPrefixIdx int
	oscPrefix    textinput.Model
	borderColor  textinput.Model
	borderWidth  textinput.Model

	// Folder selection
	currentPath   string
	availableDirs []DirectoryInfo
	dirSelectIdx  int

	// Styling
	titleStyle  lipgloss.Style
	promptStyle lipgloss.Style
	errorStyle  lipgloss.Style
	detailStyle lipgloss.Style

	// Final collected data
	searchPath   string
	mediaType    config.MediaType
	oscPrefixStr string
	colorStr     string
	widthStr     string

	// Configuration
	config *config.Config

	// Error handling
	err error

	// Processing state
	done bool
}

func initialModel() model {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		cfg = &config.DefaultConfig
	}

	currentPath, err := os.Getwd()
	if err != nil {
		currentPath = "."
	}

	availableDirs, err := listDirectoriesWithDetails(currentPath)
	if err != nil {
		availableDirs = []DirectoryInfo{}
	}

	pathInput := textinput.New()
	pathInput.Placeholder = "Enter full path to media folder"
	pathInput.Focus()

	oscPrefix := textinput.New()
	oscPrefix.Placeholder = "Enter custom OSC prefix (e.g. /streamdeck/custom)"

	borderColor := textinput.New()
	borderColor.Placeholder = fmt.Sprintf("Enter border color (default: %s)", cfg.BorderColor)
	borderColor.SetValue(cfg.BorderColor)

	borderWidth := textinput.New()
	borderWidth.Placeholder = fmt.Sprintf("Enter border width (default: %d)", cfg.BorderWidth)
	borderWidth.SetValue(strconv.Itoa(cfg.BorderWidth))

	return model{
		step:          0,
		pathInput:     pathInput,
		oscPrefix:     oscPrefix,
		borderColor:   borderColor,
		borderWidth:   borderWidth,
		mediaTypeIdx:  0,
		oscPrefixIdx:  0,
		currentPath:   currentPath,
		availableDirs: availableDirs,
		dirSelectIdx:  0,
		titleStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true),
		promptStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		errorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
		detailStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")),
		config:        cfg,
		// reset final data
		searchPath:   "",
		mediaType:    config.MediaType(0),
		oscPrefixStr: "",
		colorStr:     "",
		widthStr:     "",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nosonar
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.done {
				return initialModel(), nil
			}
			return m, tea.Quit

		case tea.KeyEnter:
			if m.done {
				return initialMainMenuModel(), nil
			}
			return m.handleEnter()

		case tea.KeyUp:
			return m.handleUp()

		case tea.KeyDown:
			return m.handleDown()
		}
	}

	// Handle input for path and custom OSC prefix
	if !m.done && m.step == 0 {
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	} else if !m.done && m.step == 2 && m.oscPrefixIdx == len(m.config.OscPrefixOptions)-1 {
		m.oscPrefix, cmd = m.oscPrefix.Update(msg)
		return m, cmd
	} else if !m.done && m.step == 3 {
		m.borderColor, cmd = m.borderColor.Update(msg)
		return m, cmd
	} else if !m.done && m.step == 4 {
		m.borderWidth, cmd = m.borderWidth.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string { //nosonar
	if m.err != nil {
		return m.errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.done {
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s",
			m.titleStyle.Render("Media Preparation Complete"),
			m.promptStyle.Render("Successfully processed media files."),
			m.promptStyle.Render(fmt.Sprintf("Path: %s", m.searchPath)),
			m.promptStyle.Render("Press Enter to return to main menu"),
		)
	}

	switch m.step {
	case 0: // Directory selection
		s := m.titleStyle.Render("StreamDeck Media Preparation") + "\n\n" //nosonar
		s += m.promptStyle.Render(fmt.Sprintf("Select a folder in %s:", m.currentPath)) + "\n"

		for i, dir := range m.availableDirs {
			cursor := " "
			if m.dirSelectIdx == i {
				cursor = ">"
			}

			// Format directory details
			if dir.IsDir {
				if dir.Name == ".." {
					s += fmt.Sprintf("%s %s (Parent Directory)\n", cursor, dir.Name)
				} else {
					s += fmt.Sprintf("%s %s (Directories: %d)\n",
						cursor,
						dir.Name,
						dir.FileCount,
					)
				}
			} else {
				s += fmt.Sprintf("%s %s (Size: %s)\n",
					cursor,
					dir.Name,
					formatFileSize(dir.Size),
				)
			}
		}

		if len(m.availableDirs) == 0 {
			s += m.errorStyle.Render("No directories found in current path.")
		}

		// Show details of selected item
		if len(m.availableDirs) > 0 {
			selected := m.availableDirs[m.dirSelectIdx]
			s += "\n" + m.detailStyle.Render(fmt.Sprintf("Selected: %s", selected.Path))
		}

		return s

	case 1: // Media type selection
		mediaTypes := []string{"Image", "Video", "Audio"}
		s := m.titleStyle.Render("StreamDeck Media Preparation") + "\n\n"
		s += m.promptStyle.Render("Select Media Type:") + "\n"

		for i, mt := range mediaTypes {
			cursor := " "
			if m.mediaTypeIdx == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, mt)
		}
		return s

	case 2: // OSC Prefix selection or input
		s := m.titleStyle.Render("StreamDeck Media Preparation") + "\n\n"

		if m.oscPrefixIdx == len(m.config.OscPrefixOptions)-1 {
			// Show custom input field
			s += m.promptStyle.Render("Enter Custom OSC Prefix:") + "\n"
			s += m.oscPrefix.View()
		} else {
			// Show selection list
			s += m.promptStyle.Render("Select OSC Prefix:") + "\n"
			for i, opt := range m.config.OscPrefixOptions {
				cursor := " "
				if m.oscPrefixIdx == i {
					cursor = ">"
				}
				s += fmt.Sprintf("%s %s (%s)\n", cursor, opt.Name, opt.Prefix)
			}
		}
		return s

	case 3: // Border Color input
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			m.titleStyle.Render("StreamDeck Media Preparation"),
			m.promptStyle.Render(fmt.Sprintf("Enter border color (default: %s):", m.config.BorderColor)),
			m.borderColor.View(),
		)

	case 4: // Border Width input
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			m.titleStyle.Render("StreamDeck Media Preparation"),
			m.promptStyle.Render(fmt.Sprintf("Enter border width (default: %d):", m.config.BorderWidth)),
			m.borderWidth.View(),
		)

	case 5: // Confirmation and processing
		return fmt.Sprintf(
			"%s\n\n%s\n\nPath: %s\nMedia Type: %s\nOSC Prefix: %s\nBorder Color: %s\nBorder Width: %s",
			m.titleStyle.Render("Confirm Details"),
			m.promptStyle.Render("Press Enter to process or Esc to cancel"),
			m.searchPath,
			[]string{"Image", "Video", "Audio"}[m.mediaType],
			m.oscPrefixStr,
			m.colorStr,
			m.widthStr,
		)

	default:
		return "Something went wrong"
	}
}

func (m *model) handleUp() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0: // Directory selection
		if len(m.availableDirs) > 0 {
			m.dirSelectIdx = (m.dirSelectIdx - 1 + len(m.availableDirs)) % len(m.availableDirs)
		}
	case 1:
		m.mediaTypeIdx = (m.mediaTypeIdx - 1 + 3) % 3
	case 2:
		if m.oscPrefixIdx != len(m.config.OscPrefixOptions)-1 || !m.oscPrefix.Focused() {
			m.oscPrefixIdx = (m.oscPrefixIdx - 1 + len(m.config.OscPrefixOptions)) % len(m.config.OscPrefixOptions)
		}
	}
	return m, nil
}

func (m *model) handleDown() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0: // Directory selection
		if len(m.availableDirs) > 0 {
			m.dirSelectIdx = (m.dirSelectIdx + 1) % len(m.availableDirs)
		}
	case 1:
		m.mediaTypeIdx = (m.mediaTypeIdx + 1) % 3
	case 2:
		if m.oscPrefixIdx != len(m.config.OscPrefixOptions)-1 || !m.oscPrefix.Focused() {
			m.oscPrefixIdx = (m.oscPrefixIdx + 1) % len(m.config.OscPrefixOptions)
		}
	}
	return m, nil
}

func (m *model) handleEnter() (tea.Model, tea.Cmd) { //nosonar
	switch m.step {
	case 0: // Directory selection
		if len(m.availableDirs) > 0 {
			selected := m.availableDirs[m.dirSelectIdx]

			// Handle parent directory navigation
			if selected.Name == ".." {
				// Move to parent directory
				m.currentPath = selected.Path
				newDirs, err := listDirectoriesWithDetails(m.currentPath)
				if err != nil {
					m.err = fmt.Errorf("error reading directory: %v", err)
					return m, nil
				}
				m.availableDirs = newDirs
				m.dirSelectIdx = 0
				return m, nil
			}

			// If selected is a directory, navigate into it
			if selected.IsDir {
				m.currentPath = selected.Path
				newDirs, err := listDirectoriesWithDetails(m.currentPath)
				if err != nil {
					m.err = fmt.Errorf("error reading directory: %v", err)
					return m, nil
				}
				m.availableDirs = newDirs
				m.dirSelectIdx = 0
				return m, nil
			}

			// If selected is a file, the search path is the one set by the directory
			// it belongs to
			m.searchPath = filepath.Dir(selected.Path)

			// Validate path
			if _, err := os.Stat(m.searchPath); os.IsNotExist(err) {
				m.err = fmt.Errorf("path does not exist: %s", m.searchPath)
				return m, nil
			}

			m.step++
			return m, nil
		} else {
			m.err = fmt.Errorf("no directories available to select")
			return m, nil
		}

	case 1: // Media type selection
		m.mediaType = config.MediaType(m.mediaTypeIdx)
		m.step++
		return m, nil

	case 2: // OSC Prefix selection or input
		if m.oscPrefixIdx == len(m.config.OscPrefixOptions)-1 {
			// Custom prefix
			prefix := m.oscPrefix.Value()
			if prefix == "" {
				m.err = fmt.Errorf("OSC prefix cannot be empty")
				return m, nil
			}
			if !strings.HasPrefix(prefix, "/") {
				prefix = "/" + prefix
			}
			m.oscPrefixStr = prefix
		} else {
			// Selected prefix
			m.oscPrefixStr = m.config.OscPrefixOptions[m.oscPrefixIdx].Prefix
		}
		m.step++
		m.borderColor.Focus()
		return m, nil

	case 3: // Border Color
		color := m.borderColor.Value()
		if color == "" {
			color = m.config.BorderColor
		}
		if !strings.HasPrefix(color, "#") || len(color) != 7 {
			m.err = fmt.Errorf("invalid color format. Must be in format #RRGGBB")
			return m, nil
		}
		m.colorStr = color
		m.step++
		m.borderWidth.Focus()
		return m, nil

	case 4: // Border Width
		width := m.borderWidth.Value()
		if width == "" {
			width = strconv.Itoa(m.config.BorderWidth)
		}
		_, err := strconv.Atoi(width)
		if err != nil {
			m.err = fmt.Errorf("border width must be a number")
			return m, nil
		}
		m.widthStr = width
		m.step++
		return m, nil

	case 5: // Process files
		width, _ := strconv.Atoi(m.widthStr)
		err := processMediaFiles(m.searchPath, m.mediaType, m.oscPrefixStr, m.colorStr, width)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.done = true
		return m, nil
	}

	return m, nil
}

// formatFileSize converts file size in bytes to human-readable format
func formatFileSize(size int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
