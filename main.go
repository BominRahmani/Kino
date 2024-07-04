package main

import (
	"log"

	"github.com/bominrahmani/kino/providers"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

type movieItem struct {
	title, year string
}

func (i movieItem) Title() string       { return i.title }
func (i movieItem) Description() string { return i.year }
func (i movieItem) FilterValue() string { return i.title }

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)
	return s
}

type model struct {
	movies      []*providers.Movie
	width       int
	height      int
	index       int
	list        list.Model
	styles      *Styles
	answerField textinput.Model
	loading     bool
	state       string
}

func New() *model {
	styles := DefaultStyles()
	answerField := textinput.New()
	answerField.Placeholder = ""
	answerField.Focus()
	return &model{
		answerField: answerField,
		styles:      styles,
		state:       "input",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			query := m.answerField.Value()
			m.answerField.SetValue("Searching")
			catalogue, err := providers.Scrape(query)
			if err != nil {
				log.Printf("Error searching movies: %v", err)
				return m, nil
			}
			m.movies = catalogue
			// iterate through movies to make a list
			items := make([]list.Item, len(catalogue))
			for i, kino := range catalogue {
				items[i] = movieItem{title: kino.Title, year: kino.Year}
			}
			m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height-10)
			m.list.Title = "Search Results"
			m.state = "list"
			return m, nil
		case "list":
			switch msg.String() {
			case "q", "esc":
				m.state = "input"
				m.answerField.SetValue("")
				return m, nil
			}
		}
	}
	if m.state == "input" {
		m.answerField, cmd = m.answerField.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	//m.answerField, cmd = m.answerField.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}
	var content string
	switch m.state {
	case "input":
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			"What movie would you like to watch?",
			m.styles.InputField.Render(m.answerField.View()),
		)
	case "list":
		content = m.list.View()
	}
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func main() {
	m := New()
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal("err: %w", err)
	}
	defer f.Close()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
