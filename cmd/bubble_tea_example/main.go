package main

// Working through the bubble tea tutorial on github. Then I will expand and add https://github.com/charmbracelet/bubbletea
// we will build a simple TODO app
import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string         // items on the todo list
	cursor   int              // which todo list item our cursor is pointing at
	selected map[int]struct{} // which todo items are selected
}

// initial state
func initialModel() model {
	return model{
		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// just return a `nil`
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// is it a key press?
	case tea.KeyMsg:

		switch msg.String() {

		// these keys should exist
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
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}

	}
	// return the updated model to bubble tea runtime for processing
	// note that we're not returning a command
	return m, nil
}

// render the UI
func (m model) View() string {
	// the header
	s := "What whould we buy at the market?\n\n"

	// iterate over our choices
	for i, choice := range m.choices {

		// is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // our cursor
		}

		// is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected
		}

		// render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// the footer
	s += "\nPress 'q' to quit"

	// send to the ui
	return s
}

// now we just need to pass our model to tea.NewProgram
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
