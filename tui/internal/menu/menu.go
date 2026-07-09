package menu

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RPDevs-Builds/kexecboot.xyz/tui/internal/network"
	"github.com/RPDevs-Builds/kexecboot.xyz/tui/internal/payload"
)

type state int

const (
	stateMainMenu state = iota
	stateWifiList
	stateWifiInput
	statePayloadList
)

type endpointsMsg []payload.Endpoint
type ssidsMsg []string
type errMsg error
type successMsg struct{}

type Model struct {
	state        state
	choices      []string
	endpoints    []payload.Endpoint
	ssids        []string
	selectedSsid string
	cursor       int
	loading      bool
	err          error
	ti           textinput.Model
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	
	return Model{
		state:   stateMainMenu,
		choices: []string{"Connect to Wi-Fi", "Fetch netboot.xyz Menu", "Reboot", "Poweroff"},
		ti:      ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func fetchEndpoints() tea.Msg {
	endpoints, err := payload.FetchEndpoints("")
	if err != nil {
		return errMsg(err)
	}
	return endpointsMsg(endpoints)
}

func executePayload(e payload.Endpoint) tea.Cmd {
	return func() tea.Msg {
		err := payload.Execute(e)
		if err != nil {
			return errMsg(err)
		}
		return nil
	}
}

func scanWifi() tea.Msg {
	ssids, err := network.Scan()
	if err != nil {
		return errMsg(err)
	}
	return ssidsMsg(ssids)
}

func connectWifi(ssid, psk string) tea.Cmd {
	return func() tea.Msg {
		err := network.Connect(ssid, psk)
		if err != nil {
			return errMsg(err)
		}
		return successMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state == stateMainMenu {
				return m, tea.Quit
			}
		}

		if m.loading {
			return m, nil
		}

		// Handle Wi-Fi password input
		if m.state == stateWifiInput {
			switch msg.String() {
			case "enter":
				m.loading = true
				return m, connectWifi(m.selectedSsid, m.ti.Value())
			case "esc":
				m.state = stateWifiList
				m.ti.SetValue("")
				m.cursor = 0
				return m, nil
			default:
				m.ti, cmd = m.ti.Update(msg)
				return m, cmd
			}
		}

		// Handle lists navigation
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			choice := m.choices[m.cursor]

			if choice == "Back" {
				m.state = stateMainMenu
				m.endpoints = nil
				m.ssids = nil
				m.err = nil
				m.choices = []string{"Connect to Wi-Fi", "Fetch netboot.xyz Menu", "Reboot", "Poweroff"}
				m.cursor = 0
				return m, nil
			}

			switch m.state {
			case stateMainMenu:
				if choice == "Reboot" {
					exec.Command("reboot").Run()
					return m, tea.Quit
				} else if choice == "Poweroff" {
					exec.Command("poweroff").Run()
					return m, tea.Quit
				} else if choice == "Fetch netboot.xyz Menu" {
					m.loading = true
					m.err = nil
					return m, fetchEndpoints
				} else if choice == "Connect to Wi-Fi" {
					m.loading = true
					m.err = nil
					return m, scanWifi
				}
			case statePayloadList:
				endpoint := m.endpoints[m.cursor]
				m.loading = true
				return m, executePayload(endpoint)
			case stateWifiList:
				m.selectedSsid = choice
				m.state = stateWifiInput
				m.ti.Focus()
				m.ti.SetValue("")
				return m, textinput.Blink
			}
		}

	case endpointsMsg:
		m.loading = false
		m.endpoints = msg
		m.choices = nil
		for _, e := range msg {
			m.choices = append(m.choices, e.Name)
		}
		m.choices = append(m.choices, "Back")
		m.state = statePayloadList
		m.cursor = 0
		return m, nil

	case ssidsMsg:
		m.loading = false
		m.ssids = msg
		m.choices = nil
		for _, s := range msg {
			m.choices = append(m.choices, s)
		}
		m.choices = append(m.choices, "Back")
		m.state = stateWifiList
		m.cursor = 0
		return m, nil

	case successMsg:
		m.loading = false
		m.state = stateMainMenu
		m.choices = []string{"Connect to Wi-Fi", "Fetch netboot.xyz Menu", "Reboot", "Poweroff"}
		m.cursor = 0
		m.err = nil
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg
		// Return to main menu on error if we were fetching or scanning
		if m.state != stateWifiInput {
		    m.state = stateMainMenu
		    m.choices = []string{"Connect to Wi-Fi", "Fetch netboot.xyz Menu", "Reboot", "Poweroff"}
		    m.cursor = 0
		}
		return m, nil
	}
	
	// Update textinput even if no key was matched to handle blinking
	if m.state == stateWifiInput {
		m.ti, cmd = m.ti.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
var itemStyle = lipgloss.NewStyle().PaddingLeft(2)
var selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).MarginTop(1)
var loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Italic(true).MarginTop(1)
var promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

func (m Model) View() string {
	s := titleStyle.Render("kexecboot.xyz") + "\n\n"

	if m.state == stateWifiInput {
		s += promptStyle.Render(fmt.Sprintf("Enter PSK for %s:", m.selectedSsid)) + "\n"
		s += m.ti.View() + "\n\n(esc to go back)\n"
	} else {
		for i, choice := range m.choices {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
				s += fmt.Sprintf("%s %s\n", cursor, selectedItemStyle.Render(choice))
			} else {
				s += fmt.Sprintf("%s %s\n", cursor, itemStyle.Render(choice))
			}
		}
	}

	if m.loading {
		s += loadingStyle.Render("\nLoading...")
	} else if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("\nError: %v", m.err))
	}

	s += "\nPress ctrl+c to quit.\n"
	return s
}
