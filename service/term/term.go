package term

import (
	"log"
	"os"

	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/muesli/termenv"
)

var logger *log.Logger = log.Default()

type termUpdate struct {
	pForeground termenv.Color
	pBackground termenv.Color
	changed     bool
}

func ApplyColors(foreground, background string) termUpdate {
	output := termenv.DefaultOutput()

	currentBackground := output.BackgroundColor()
	currentForeground := output.ForegroundColor()
	colorSet := false
	_, inTmux := os.LookupEnv("TMUX")
	logger.Printf("Is in tmux: %v", inTmux)
	if !inTmux {
		output.SetBackgroundColor(termenv.RGBColor(constants.BackgroundColour))
		output.SetForegroundColor(termenv.RGBColor(constants.ForegroundColour))
		colorSet = true
	}
	return termUpdate{changed: colorSet, pForeground: currentForeground, pBackground: currentBackground}
}

func ResetColors(update termUpdate) {
    output := termenv.DefaultOutput()
    if update.changed {
        output.SetBackgroundColor(update.pBackground)
        output.SetForegroundColor(update.pForeground)
    }
}
