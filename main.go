package main

import (
	"fmt"
	"log"
	"time"

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

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)
	return s
}

type movieItem struct {
	title, year string
}

func (i movieItem) Title() string       { return i.title }
func (i movieItem) Description() string { return i.year }
func (i movieItem) FilterValue() string { return i.title }

type model struct {
	movies      []*providers.Movie
	width       int
	height      int
	list        list.Model
	styles      *Styles
	answerField textinput.Model
	state       string
	loading     bool
	initialized bool
}

func New() *model {
	styles := DefaultStyles()
	answerField := textinput.New()
	answerField.Placeholder = "Enter movie title"
	answerField.Focus()

	// Initialize the list with an empty delegate
	emptyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)

	return &model{
		answerField: answerField,
		styles:      styles,
		state:       "input",
		list:        emptyList,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.initialized {
			m.initialized = true
			return m, tea.Batch(
				func() tea.Msg {
					time.Sleep(100 * time.Millisecond) // Small delay to ensure window size is set
					return tea.WindowSizeMsg{Width: m.width, Height: m.height}
				},
			)
		}

	case tea.KeyMsg:
		switch m.state {
		case "input":
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				if !m.loading {
					m.loading = true
					query := m.answerField.Value()
					return m, tea.Batch(
						func() tea.Msg {
							return searchMsg(query)
						},
						tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
							return tickMsg{}
						}),
					)
				}
			}
		case "list":
			switch msg.String() {
			case "q", "esc":
				m.state = "input"
				m.answerField.SetValue("")
				return m, nil
			}
		}

	case searchMsg:
		return m, func() tea.Msg {
			start := time.Now()
			flixHQ := providers.NewFlixHQProvider()
			catalogue, err := flixHQ.Scrape(string(msg))
			log.Printf("Scrape took %v", time.Since(start))
			return searchResultMsg{movies: catalogue, err: err}
		}

	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.state = "input"
			m.answerField.SetValue(fmt.Sprintf("Error: %v", msg.err))
		} else {
			start := time.Now()
			m.movies = msg.movies
			items := make([]list.Item, len(m.movies))
			for i, kino := range m.movies {
				items[i] = movieItem{title: kino.Title, year: kino.Year}
			}

			m.list.SetItems(items)
			m.list.SetSize(m.width, m.height-10)
			m.list.Title = "Search Results"
			m.state = "list"

			log.Printf("List initialization took %v", time.Since(start))
		}
		return m, nil

	case tickMsg:
		if m.loading {
			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return tickMsg{}
			})
		}
	}

	if m.state == "input" {
		m.answerField, cmd = m.answerField.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if !m.initialized {
		return "Initializing..."
	}
	var content string
	switch m.state {
	case "input":
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			"What movie would you like to watch?",
			m.styles.InputField.Render(m.answerField.View()),
		)
		if m.loading {
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				"Searching...",
			)
		}
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

type searchMsg string

type searchResultMsg struct {
	movies []*providers.Movie
	err    error
}

type tickMsg struct{}

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
