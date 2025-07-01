package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const zone = "restricted"

type model struct {
	ports       []string
	cursor      int
	addingPort  bool
	inputBuffer string
	err         error
}

func initialModel() model {
	ports, err := listPorts()
	return model{ports: ports, err: err}
}

func listPorts() ([]string, error) {
	out, err := exec.Command("sudo", "firewall-cmd", "--permanent", "--zone="+zone, "--list-ports").Output()
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return []string{}, nil
	}
	return strings.Split(trimmed, " "), nil
}

func addPort(port string) error {
	cmd := exec.Command("sudo", "firewall-cmd", "--zone="+zone, "--add-port="+port, "--permanent")
	if err := cmd.Run(); err != nil {
		return err
	}
	return exec.Command("sudo", "firewall-cmd", "--reload").Run()
}

func removePort(port string) error {
	cmd := exec.Command("sudo", "firewall-cmd", "--zone="+zone, "--remove-port="+port, "--permanent")
	if err := cmd.Run(); err != nil {
		return err
	}
	return exec.Command("sudo", "firewall-cmd", "--reload").Run()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.ports)-1 {
				m.cursor++
			}

		case "enter":
			if m.addingPort {
				if strings.TrimSpace(m.inputBuffer) != "" {
					err := addPort(strings.TrimSpace(m.inputBuffer))
					m.err = err
					m.addingPort = false
					m.inputBuffer = ""
					m.ports, _ = listPorts()
				}
			} else if len(m.ports) > 0 {
				err := removePort(m.ports[m.cursor])
				m.err = err
				m.ports, _ = listPorts()
				if m.cursor >= len(m.ports) && m.cursor > 0 {
					m.cursor--
				}
			}

		case "a":
			m.addingPort = true
			m.inputBuffer = ""

		case "esc":
			m.addingPort = false
			m.inputBuffer = ""

		case "backspace":
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			}

		default:
			if m.addingPort {
				m.inputBuffer += msg.String()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("firewalld TUI - Zone: " + zone)
	b.WriteString(title + "\n\n")

	if m.addingPort {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Add new port (e.g., 8080/tcp): ") + m.inputBuffer + "\n")
		b.WriteString("[Enter] to submit, [Esc] to cancel\n")
		return b.String()
	}

	if len(m.ports) == 0 {
		b.WriteString("No ports configured.\n")
	} else {
		for i, port := range m.ports {
			cursor := "  "
			if m.cursor == i {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("➤ ")
				port = lipgloss.NewStyle().Bold(true).Render(port)
			}
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, port))
		}
		b.WriteString("\n[↑↓] Navigate  [Enter] Delete  [a] Add  [q] Quit\n")
	}

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Italic(true)
		b.WriteString("\nError: " + errStyle.Render(m.err.Error()) + "\n")
	}

	return b.String()
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
