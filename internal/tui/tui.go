package tui

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/slashformotion/radioboat/internal/players"
	"github.com/slashformotion/radioboat/internal/urls"
)

var width int
var height int

type model struct {
	savedTracks   []string
	stations      []*urls.Station
	cursor        int
	player        players.RadioPlayer
	help          help.Model
	dj            Dj
	trackFilePath string
	mb            *MessageBox
}

type Dj struct {
	currentStation string
	muted          bool
	volume         int
	currentTrack   string
}

func HeaderToString(currentStation string, trackName string, volume int, muted bool) string {
	var mutedStr string
	if muted {
		mutedStr = fmt.Sprintf("Muted(%d)", volume)
	} else {
		mutedStr = fmt.Sprintf("Volume %d", volume)
	}
	statusStr := header_status_s.Render(currentStation)
	volumeStr := header_volume_s.Render(mutedStr)
	centerStr := header_center_s.Copy().
		Width(width - lipgloss.Width(statusStr) - lipgloss.Width(volumeStr) - 3). // -3 because of the doc margin
		Render(trackName)
	s := lipgloss.JoinHorizontal(lipgloss.Top, statusStr, centerStr, volumeStr)
	s += "\n\n"
	return s
}

func (m model) Init() tea.Cmd {
	return tea.Batch(CmdTickerMessageBox, CmdTickerTrackname)
}

func (m model) Update(tmsg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := tmsg.(type) {
	case TickMessageBox:
		return m, m.mb.Update(tmsg)
	case SaveTrackMsg:
		if msg.err != nil {
			m.mb.append(
				NewMessageFromErr(msg.err),
			)
			return m, nil
		} else {
			m.mb.append(
				NewMessage(fmt.Sprintf("Just saved track name %q to %q", msg.trackName, m.trackFilePath)),
			)
		}
	case Tick:
		m.dj.currentTrack = m.player.NowPlaying()
		return m, CmdTickerTrackname
	case tea.WindowSizeMsg:
		width = msg.Width
		height = msg.Height
		m.help.Width = msg.Width

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
			m.dj.currentStation = m.stations[m.cursor].Name

		case key.Matches(msg, DefaultKeyMap.VolumeUp):
			m.player.IncVolume()
			m.dj.volume = m.player.Volume()
		case key.Matches(msg, DefaultKeyMap.VolumeDown):
			m.player.DecVolume()
			m.dj.volume = m.player.Volume()
		case key.Matches(msg, DefaultKeyMap.SaveTrack):
			trackName := m.player.NowPlaying()
			for _, s := range m.savedTracks {
				if trackName == s {
					return m, nil
				}
			}
			m.savedTracks = append(m.savedTracks, trackName)
			return m, CmdSaveTrack(m.trackFilePath, trackName)
		}
	}

	return m, nil
}

func (m model) View() string {
	s := HeaderToString(m.dj.currentStation, m.dj.currentTrack, m.dj.volume, m.dj.muted)

	// Iterate over our choices
	for i, station := range m.stations {

		cursor := " " // no cursor
		name := station.Name
		if m.cursor == i {
			cursor = ">" // cursor!
			name = list_selected_s.Render(station.Name)
		}
		if m.dj.currentStation == station.Name {
			name = list_selected_s.Copy().Italic(true).Bold(false).Render(station.Name)
		}

		s += fmt.Sprintf("%s %s\n", cursor, name)
	}
	s += m.mb.View()
	helpView := m.help.View(DefaultKeyMap)
	s += "\n\n" + helpView

	return docStyle.Render(s)
}
func InitialModel(p players.RadioPlayer, stations []*urls.Station, volume int, trackFilePath string) model {
	m := model{
		player:   p,
		stations: stations,
		dj: Dj{
			currentStation: "Not Playing",
			volume:         volume,
		},
		help:          help.New(),
		trackFilePath: trackFilePath,
		mb:            new(MessageBox),
	}
	m.player.Init()
	m.player.SetVolume(volume)
	m.dj.volume = m.player.Volume()
	m.dj.muted = m.player.IsMute()

	// m.help.ShowAll = true
	return m
}

// wait 1 sec and then send a Tick
func CmdTickerTrackname() tea.Msg {
	time.Sleep(time.Second)
	return Tick{}
}

// tea.Msg send by CmdTickerTrackname
type Tick struct{}

func CmdSaveTrack(trackFilePath, track string) tea.Cmd {
	return func() tea.Msg {
		var msg SaveTrackMsg = SaveTrackMsg{err: nil}
		if track == "" {
			return msg
		}
		trackFile, err := os.OpenFile(trackFilePath, os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			msg.err = err
			return msg
		}
		_, err = fmt.Fprintf(trackFile, "%s\n", track)
		if err != nil {
			msg.err = err
			return msg
		}
		msg.err = trackFile.Close()
		msg.trackName = track
		return msg
	}
}

// tea.Msg send by CmdSaveTrack
type SaveTrackMsg struct {
	err       error
	trackName string
}
