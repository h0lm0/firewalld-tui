package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var defaultZone = "restricted"

type model struct {
	ports      []string
	cursor     int
	adding     bool
	input      string
	zone       string
	errMessage string
}

func getZones() ([]string, error) {
	out, err := exec.Command("firewall-cmd", "--get-zones").Output()
	if err != nil {
		return nil, fmt.Errorf("firewall-cmd --get-zones failed: %w", err)
	}
	zones := strings.Fields(strings.TrimSpace(string(out)))
	return zones, nil
}

func isValidZone(zone string, allowed []string) bool {
	for _, z := range allowed {
		if z == zone {
			return true
		}
	}
	return false
}

func getPorts(zone string) ([]string, error) {
	out, err := exec.Command("sudo", "firewall-cmd", "--permanent", "--zone="+zone, "--list-ports").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get ports for zone %s: %w", zone, err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []string{}, nil
	}
	return strings.Split(raw, " "), nil
}

func removePort(zone, port string) error {
	cmd := exec.Command("sudo", "firewall-cmd", "--zone="+zone, "--remove-port="+port, "--permanent")
	if err := cmd.Run(); err != nil {
		return err
	}
	return exec.Command("sudo", "firewall-cmd", "--reload").Run()
}

func addPort(zone, port string) error {
	cmd := exec.Command("sudo", "firewall-cmd", "--zone="+zone, "--add-port="+port, "--permanent")
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a":
			m.adding = true
			m.input = ""
			m.errMessage = ""
			return m, nil
		case "enter":
			if m.adding {
				err := addPort(m.zone, m.input)
				if err != nil {
					m.errMessage = "Erreur: " + err.Error()
				} else {
					m.ports, _ = getPorts(m.zone)
					m.errMessage = ""
				}
				m.adding = false
				m.input = ""
			} else if len(m.ports) > 0 {
				err := removePort(m.zone, m.ports[m.cursor])
				if err != nil {
					m.errMessage = "Erreur: " + err.Error()
				} else {
					m.ports, _ = getPorts(m.zone)
					if m.cursor >= len(m.ports) {
						m.cursor = max(0, len(m.ports)-1)
					}
					m.errMessage = ""
				}
			}
			return m, nil
		case "esc":
			m.adding = false
			m.input = ""
			m.errMessage = ""
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.ports)-1 {
				m.cursor++
			}
			return m, nil
		default:
			if m.adding {
				if msg.Type == tea.KeyRunes {
					m.input += msg.String()
				} else if msg.Type == tea.KeyBackspace && len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Ports in zone: %s\n\n", m.zone))

	if m.errMessage != "" {
		b.WriteString("[!] " + m.errMessage + "\n\n")
	}

	for i, port := range m.ports {
		cursor := "  "
		if i == m.cursor {
			cursor = "→ "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, port))
	}

	if len(m.ports) == 0 {
		b.WriteString("No ports configured.\n")
	}

	b.WriteString("\n")

	if m.adding {
		b.WriteString(fmt.Sprintf("Add port (e.g. 443/tcp): %s\n", m.input))
		b.WriteString("[Enter = Add, Esc = Cancel]\n")
	} else {
		b.WriteString("[a = Add] [Enter = Remove selected] [q = Quit]\n")
	}

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	zone := defaultZone
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if strings.HasPrefix(arg, "--zone=") {
			zone = strings.TrimPrefix(arg, "--zone=")
		} else {
			fmt.Println("Usage: firewalld-tui [--zone=<name>]")
			os.Exit(1)
		}
	}

	allowedZones, err := getZones()
	if err != nil {
		fmt.Println("Erreur récupération zones:", err)
		os.Exit(1)
	}

	if !isValidZone(zone, allowedZones) {
		fmt.Printf("Zone non valide: %q\nZones disponibles: %s\n", zone, strings.Join(allowedZones, ", "))
		os.Exit(1)
	}

	ports, err := getPorts(zone)
	if err != nil {
		fmt.Println("Erreur chargement ports:", err)
		os.Exit(1)
	}

	m := model{
		ports: ports,
		zone:  zone,
	}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Erreur TUI:", err)
		os.Exit(1)
	}
}
