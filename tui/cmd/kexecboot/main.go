package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RPDevs-Builds/kexecboot.xyz/tui/internal/menu"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := menu.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
		os.Exit(1)
	}

	fmt.Println("kexecboot TUI exited.")
}
