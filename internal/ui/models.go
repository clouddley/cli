package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(1).
		PaddingLeft(4).
		PaddingRight(4).
		PaddingBottom(1)

	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))
)

// EnvironmentChoice represents the environment selection
type EnvironmentChoice struct {
	Name        string
	Description string
}

// InstanceType represents an EC2 instance type with pricing
type InstanceType struct {
	Type        string
	VCPUs       string
	Memory      string
	Disk        string
	MonthlyCost string
}

// EnvironmentModel for selecting development vs production
type EnvironmentModel struct {
	choices  []EnvironmentChoice
	cursor   int
	selected bool
}

func NewEnvironmentModel() EnvironmentModel {
	return EnvironmentModel{
		choices: []EnvironmentChoice{
			{"Development/Test", "T-family instances optimized for burstable workloads"},
			{"Production", "M-family (memory) and C-family (compute) instances for production workloads"},
		},
	}
}

func (m EnvironmentModel) Init() tea.Cmd {
	return nil
}

func (m EnvironmentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m EnvironmentModel) View() string {
	s := titleStyle.Render("Select Environment") + "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			choice.Name = selectedStyle.Render(choice.Name)
		} else {
			choice.Name = normalStyle.Render(choice.Name)
		}

		s += fmt.Sprintf("%s %s\n   %s\n\n", cursor, choice.Name, choice.Description)
	}

	s += "\nPress q to quit.\n"
	return s
}

func (m EnvironmentModel) Selected() int {
	if m.selected {
		return m.cursor
	}
	return -1
}

// InstanceSelectionModel for selecting instance types
type InstanceSelectionModel struct {
	table    table.Model
	selected bool
}

func NewInstanceSelectionModel(instances []InstanceType, title string) InstanceSelectionModel {
	columns := []table.Column{
		{Title: "Type", Width: 12},
		{Title: "vCPUs", Width: 8},
		{Title: "Memory", Width: 10},
		{Title: "Disk", Width: 10},
		{Title: "Monthly Cost", Width: 15},
	}

	rows := make([]table.Row, len(instances))
	for i, instance := range instances {
		rows[i] = table.Row{
			instance.Type,
			instance.VCPUs,
			instance.Memory,
			instance.Disk,
			instance.MonthlyCost,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return InstanceSelectionModel{
		table: t,
	}
}

func (m InstanceSelectionModel) Init() tea.Cmd {
	return nil
}

func (m InstanceSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.selected = true
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m InstanceSelectionModel) View() string {
	title := titleStyle.Render("Select Instance Type")
	return title + "\n\n" + m.table.View() + "\n\nPress enter to select, q to quit.\n"
}

func (m InstanceSelectionModel) Selected() int {
	if m.selected {
		return m.table.Cursor()
	}
	return -1
}

// SSHKeySelectionModel for selecting SSH keys
type SSHKeySelectionModel struct {
	choices  []string
	cursor   int
	selected bool
}

func NewSSHKeySelectionModel(keyTypes []string) SSHKeySelectionModel {
	return SSHKeySelectionModel{
		choices: keyTypes,
	}
}

func (m SSHKeySelectionModel) Init() tea.Cmd {
	return nil
}

func (m SSHKeySelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SSHKeySelectionModel) View() string {
	s := titleStyle.Render("Multiple SSH Keys Found - Select One") + "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		text := choice
		if m.cursor == i {
			cursor = ">"
			text = selectedStyle.Render(choice)
		} else {
			text = normalStyle.Render(choice)
		}

		s += fmt.Sprintf("%s %s\n", cursor, text)
	}

	s += "\nPress enter to select, q to quit.\n"
	return s
}

func (m SSHKeySelectionModel) Selected() int {
	if m.selected {
		return m.cursor
	}
	return -1
}

// FormatOutput formats success output with styling
func FormatOutput(title, content string) string {
	titleStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#04B575")).
		Render(title)
	
	return fmt.Sprintf("%s\n%s", titleStyled, content)
}

// FormatError formats error output with styling
func FormatError(content string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF5F87")).
		Render(content)
}

// InstanceDetailsTable represents a table for displaying instance details
type InstanceDetailsTable struct {
	rows [][]string
}

// NewInstanceDetailsTable creates a new instance details table
func NewInstanceDetailsTable() *InstanceDetailsTable {
	return &InstanceDetailsTable{
		rows: make([][]string, 0),
	}
}

// AddRow adds a key-value row to the table
func (t *InstanceDetailsTable) AddRow(key, value string) {
	t.rows = append(t.rows, []string{key, value})
}

// Render renders the table with styling
func (t *InstanceDetailsTable) Render() string {
	if len(t.rows) == 0 {
		return ""
	}

	// Calculate max width for keys
	maxKeyWidth := 0
	for _, row := range t.rows {
		if len(row[0]) > maxKeyWidth {
			maxKeyWidth = len(row[0])
		}
	}

	// Add padding
	maxKeyWidth += 2

	var result strings.Builder
	
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Width(maxKeyWidth).
		AlignHorizontal(lipgloss.Right)
	
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))

	for _, row := range t.rows {
		styledKey := keyStyle.Render(row[0] + ":")
		styledValue := valueStyle.Render(row[1])
		result.WriteString(fmt.Sprintf("%s %s\n", styledKey, styledValue))
	}

	return result.String()
}

// ShowBanner displays the Clouddley CLI banner
func ShowBanner() string {
	banner := titleStyle.Render("Clouddley")
	
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render("The Backend Infrastructure platform for your compute.")
	
	return fmt.Sprintf("%s\n%s\n", banner, subtitle)
}

// LoadingModel represents a loading spinner
type LoadingModel struct {
	spinner  string
	frame    int
	message  string
	finished bool
}

// NewLoadingModel creates a new loading model
func NewLoadingModel(message string) LoadingModel {
	return LoadingModel{
		message: message,
		frame:   0,
		spinner: "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏",
	}
}

func (m LoadingModel) Init() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

type TickMsg struct {
	Time time.Time
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		if !m.finished {
			m.frame++
			spinnerRunes := []rune(m.spinner)
			if m.frame >= len(spinnerRunes) {
				m.frame = 0
			}
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return TickMsg{Time: t}
			})
		}
	case tea.KeyMsg:
		// Don't quit on any key during loading
		return m, nil
	}
	return m, nil
}

func (m LoadingModel) View() string {
	if m.finished {
		return ""
	}
	
	spinnerRunes := []rune(m.spinner)
	spinnerChar := string(spinnerRunes[m.frame])
	
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)
	
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))
	
	return fmt.Sprintf("%s %s", 
		spinnerStyle.Render(spinnerChar), 
		messageStyle.Render(m.message))
}

// Finish stops the loading animation
func (m *LoadingModel) Finish() {
	m.finished = true
}

// ConfirmationModel for yes/no prompts
type ConfirmationModel struct {
	message  string
	selected bool // true = yes, false = no
	choice   int  // 0 = yes, 1 = no
	answered bool
}

// NewConfirmationModel creates a new confirmation model
func NewConfirmationModel(message string) ConfirmationModel {
	return ConfirmationModel{
		message:  message,
		selected: true, // Default to yes
		choice:   0,
		answered: false,
	}
}

func (m ConfirmationModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.choice = 0
			m.selected = true
		case "right", "l":
			m.choice = 1
			m.selected = false
		case "enter", " ":
			m.answered = true
			return m, tea.Quit
		case "y", "Y":
			m.choice = 0
			m.selected = true
			m.answered = true
			return m, tea.Quit
		case "n", "N":
			m.choice = 1
			m.selected = false
			m.answered = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ConfirmationModel) View() string {
	s := titleStyle.Render("Clouddley Dashboard Integration") + "\n\n"
	s += m.message + "\n\n"

	// Options with proper styling
	var yesOption, noOption string
	
	if m.choice == 0 {
		yesOption = fmt.Sprintf("[ %s ]", selectedStyle.Render("Yes"))
		noOption = fmt.Sprintf("[ %s ]", normalStyle.Render("No"))
	} else {
		yesOption = fmt.Sprintf("[ %s ]", normalStyle.Render("Yes"))
		noOption = fmt.Sprintf("[ %s ]", selectedStyle.Render("No"))
	}

	s += fmt.Sprintf("%s    %s\n\n", yesOption, noOption)
	s += "Use arrow keys or y/n to select, enter to confirm, q to quit.\n"
	return s
}

func (m ConfirmationModel) Selected() bool {
	if m.answered {
		return m.selected
	}
	return false // Default to false if not answered
}

func (m ConfirmationModel) Answered() bool {
	return m.answered
}
