package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/slashformotion/radioboat/internal/players"
	"github.com/slashformotion/radioboat/internal/urls"
)

type model struct {
	stations []*urls.Station
	cursor   int
	player   players.RadioPlayer
	help     help.Model
	dj       Dj
}

type Dj struct {
	current string
	muted   bool
	volume  int
}

func (dj Dj) ToString() string {
	if dj.muted {
		return fmt.Sprintf(" %s - Muted(%d)", strings.Title(dj.current), dj.volume)
	} else {
		return fmt.Sprintf(" %s - Volume %d", strings.Title(dj.current), dj.volume)
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		//lets change the width of evry thing we need
		header_s = header_s.Width(msg.Width)

	case tea.KeyMsg:

		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Up):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
		case key.Matches(msg, DefaultKeyMap.Down):
			m.cursor++
			if m.cursor >= len(m.stations) {
				m.cursor = len(m.stations) - 1
			}
		case key.Matches(msg, DefaultKeyMap.ToggleMute):
			m.player.ToggleMute()
			m.dj.muted = m.player.IsMute()
		case key.Matches(msg, DefaultKeyMap.Play):
			m.player.Play(m.stations[m.cursor].Url)
			m.dj.current = m.stations[m.cursor].Name

		case key.Matches(msg, DefaultKeyMap.VolumeUp):
			m.player.IncVolume()
			m.dj.volume = m.player.Volume()
		case key.Matches(msg, DefaultKeyMap.VolumeDown):
			m.player.DecVolume()
			m.dj.volume = m.player.Volume()
		}
	}

	return m, nil
}
func (m model) View() string {
	s := header_s.Render(m.dj.ToString())
	s += "\n\n"

	// Iterate over our choices
	for i, station := range m.stations {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		name := station.Name
		if m.cursor == i {
			cursor = ">" // cursor!
			name = list_selected_s.Render(station.Name)
		}
		if m.dj.current == station.Name {
			name = list_selected_s.Copy().Italic(true).Bold(false).Render(station.Name)
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, name)
	}
	helpView := m.help.View(DefaultKeyMap)
	s += "\n\n" + helpView

	return docStyle.Render(s)
}
func InitialModel(p players.RadioPlayer, stations []*urls.Station, volume int) model {
	m := model{
		player:   p,
		stations: stations,
		dj: Dj{
			current: "Not Playing",
			volume:  volume,
		},
		help: help.New(),
	}
	m.player.Init()
	m.player.SetVolume(volume)
	m.dj.volume = m.player.Volume()
	m.dj.muted = m.player.IsMute()

	// m.help.ShowAll = true
	return m
}
