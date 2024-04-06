package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/eslam-allam/spring-initializer-go/models/mainModel"
	"github.com/eslam-allam/spring-initializer-go/service/files"
	"github.com/eslam-allam/spring-initializer-go/service/term"
)

var logger *log.Logger = log.Default()

func main() {
	tmpDir := os.TempDir()
    f, err := tea.LogToFile(path.Join(tmpDir, constants.LogFileName), "Main loop")
	if err != nil {
		fmt.Printf("Failed to start logger: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	targetDirectory := "."

	args := os.Args[1:]

	if len(args) > 0 {
		targetDirectory, err = files.ExpandAndMakeDir(args[0])
		if err != nil {
			logger.Printf("Error making directory: %v", err)
			os.Exit(1)
		}
	}

	p := tea.NewProgram(mainModel.New(
		mainModel.WithSpinner(spinner.New(spinner.WithSpinner(spinner.Dot))),
		mainModel.WithTargetDir(targetDirectory),
	), tea.WithAltScreen(), tea.WithMouseCellMotion())

	colorUpdate := term.ApplyColors(constants.ForegroundColour, constants.BackgroundColour)

	if _, err := p.Run(); err != nil {
		logger.Printf("Error occurred in main loop: %v", err)
		defer os.Exit(1)
	}

	term.ResetColors(colorUpdate)
}
