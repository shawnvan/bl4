package app

import (
	"fmt"

	"github.com/shawnvan/bl4/internal/tui/views"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/rivo/tview"
)

// Run starts the TUI application
func Run() error {
	logger.Sugar().Info("Starting BL4 TUI application")

	// Create application
	app := tview.NewApplication()
	loadingText := tview.NewTextView()
	loadingText.SetText("BL4 TUI Loading...")
	app.SetRoot(loadingText, true)

	// Create main view
	mainView := views.NewMainView(app)

	// Simplified keyboard shortcuts - disabled for now
	// app.SetInputCapture(func(event *tview.EventKey) bool {
	// 	switch event.Key() {
	// 	case tview.KeyCtrlQ:
	// 		app.Stop()
	// 		return true
	// 	case tview.KeyF1:
	// 		mainView.ShowHelp()
	// 		return true
	// 	case tview.KeyF5:
	// 		mainView.Refresh()
	// 		return true
	// 	}
	// 	return false
	// })

	// Set root view
	app.SetRoot(mainView.GetLayout(), true)

	// Start application
	if err := app.Run(); err != nil {
		return fmt.Errorf("failed to run TUI application: %w", err)
	}

	return nil
}