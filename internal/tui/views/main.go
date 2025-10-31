package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/shawnvan/bl4/internal/tui/views/batch"
	"github.com/shawnvan/bl4/internal/tui/views/common"
)

// MainView represents the main TUI interface
type MainView struct {
	app        *tview.Application
	layout     *tview.Flex
	pages      *tview.Pages
	statusBar  *tview.TextView
	helpModal  *tview.Modal
	aboutModal *tview.Modal
	batchView  *batch.BatchView

	// Pages
	singleDecodePage *tview.Flex
	batchDecodePage   *tview.Flex
	encodePage       *tview.Flex
	analyzePage      *tview.Flex
}

// NewMainView creates a new main view
func NewMainView(app *tview.Application) *MainView {
	mv := &MainView{
		app:   app,
		pages: tview.NewPages(),
	}

	mv.createLayout()
	mv.createPages()
	mv.createStatusAndModals()
	mv.setupEventHandlers()

	return mv
}

// GetLayout returns the main layout
func (mv *MainView) GetLayout() tview.Primitive {
	return mv.layout
}

// createLayout creates the main layout structure
func (mv *MainView) createLayout() {
	mv.layout = tview.NewFlex()
	mv.layout.SetDirection(tview.FlexRow)

	// Create main content area
	mv.layout.AddItem(mv.pages, 0, 1, true)

	// Create status bar
	mv.statusBar = tview.NewTextView()
	mv.statusBar.SetDynamicColors(true)
	mv.statusBar.SetRegions(true)
	mv.layout.AddItem(mv.statusBar, 0, 1, false)
}

// createPages creates all the pages
func (mv *MainView) createPages() {
	// Single decode page
	mv.singleDecodePage = mv.createSingleDecodePage()
	mv.pages.AddPage("Single Decode", mv.singleDecodePage, true, true)

	// Batch decode page
	mv.batchDecodePage = mv.createBatchDecodePage()
	mv.pages.AddPage("Batch Decode", mv.batchDecodePage, true, false)

	// Encode page
	mv.encodePage = mv.createEncodePage()
	mv.pages.AddPage("Encode", mv.encodePage, true, false)

	// Analysis page
	mv.analyzePage = mv.createAnalyzePage()
	mv.pages.AddPage("Analysis", mv.analyzePage, true, false)

	// Show first page by default
	mv.pages.SwitchToPage("Single Decode")
}

// createSingleDecodePage creates the single decode page
func (mv *MainView) createSingleDecodePage() *tview.Flex {
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.SetBorder(true)
	flex.SetTitle(" Single Decode ")

	// Input section
	inputSection := tview.NewFlex()
	inputSection.SetDirection(tview.FlexColumn)
	inputSection.SetBorder(false)
	inputSection.SetTitle(" Input ")

	inputLabel := tview.NewTextView()
	inputLabel.SetText("Serial Code:")
	inputLabel.SetBorder(false)

	inputField := tview.NewInputField()
	inputField.SetPlaceholder("Enter BL4 serial code (@U...)")
	inputField.SetBorder(true)

	decodeButton := tview.NewButton("Decode")
	decodeButton.SetSelectedFunc(func() {
		mv.handleSingleDecode(inputField.GetText())
	})

	clearButton := tview.NewButton("Clear")
	clearButton.SetSelectedFunc(func() {
		inputField.SetText("")
	})

	buttonsFlex := tview.NewFlex()
	buttonsFlex.AddItem(decodeButton, 0, 1, false)
	buttonsFlex.AddItem(clearButton, 0, 1, false)

	inputSection.AddItem(inputLabel, 0, 1, false)
	inputSection.AddItem(inputField, 0, 1, true)
	inputSection.AddItem(buttonsFlex, 0, 1, false)

	// Results section
	resultsSection := tview.NewFlex()
	resultsSection.SetDirection(tview.FlexColumn)
	resultsSection.SetBorder(false)
	resultsSection.SetTitle(" Results ")

	resultsView := tview.NewTextView()
	resultsView.SetBorder(true)
	resultsView.SetScrollable(true)
	resultsView.SetDynamicColors(true)

	resultsSection.AddItem(resultsView, 0, 1, true)

	flex.AddItem(inputSection, 0, 1, false)
	flex.AddItem(resultsSection, 0, 1, true)

	return flex
}

// createBatchDecodePage creates the batch decode page
func (mv *MainView) createBatchDecodePage() *tview.Flex {
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.SetBorder(true)
	flex.SetTitle(" Batch Decode ")

	// Create batch view
	mv.batchView = batch.NewBatchView()

	// Set up batch view callback for navigation
	mv.batchView.SetNavigateToSingle(func() {
		mv.pages.SwitchToPage("Single Decode")
	})

	flex.AddItem(mv.batchView.GetLayout(), 0, 1, true)

	return flex
}

// createEncodePage creates the encode page
func (mv *MainView) createEncodePage() *tview.Flex {
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.SetBorder(true)
	flex.SetTitle(" Encode ")

	content := tview.NewTextView()
	content.SetText("Encode functionality coming soon...\n\nThis page will allow you to:\n• Encode structured item data to BL4 serial codes\n• Preview results before encoding\n• Save encoded codes to files\n• Batch encode multiple items")
	content.SetBorder(true)
	content.SetScrollable(true)

	flex.AddItem(content, 0, 1, true)

	return flex
}

// createAnalyzePage creates the analysis page
func (mv *MainView) createAnalyzePage() *tview.Flex {
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.SetBorder(true)
	flex.SetTitle(" Analysis ")

	content := tview.NewTextView()
	content.SetText("Analysis tools coming soon...\n\nThis page will provide:\n• Pattern analysis of item codes\n• Statistical analysis of batch results\n• Item distribution charts\n• Performance metrics\n• Export options for analysis data")
	content.SetBorder(true)
	content.SetScrollable(true)

	flex.AddItem(content, 0, 1, true)

	return flex
}

// createStatusAndModals creates status bar and modal dialogs
func (mv *MainView) createStatusAndModals() {
	// Help modal
	mv.helpModal = tview.NewModal()
	mv.helpModal.SetText(common.GetHelpText())
	mv.helpModal.AddButtons([]string{"Close"})
	mv.helpModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		mv.app.SetRoot(mv.layout, true)
	})

	// About modal
	mv.aboutModal = tview.NewModal()
	mv.aboutModal.SetText(common.GetAboutText())
	mv.aboutModal.AddButtons([]string{"Close"})
	mv.aboutModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		mv.app.SetRoot(mv.layout, true)
	})
}

// setupEventHandlers sets up event handlers
func (mv *MainView) setupEventHandlers() {
	mv.updateStatusBar()
}

// handleSingleDecode handles single decode operation
func (mv *MainView) handleSingleDecode(code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		mv.showStatus("❌ Please enter a serial code")
		return
	}

	mv.showStatus("🔄 Decoding serial code...")

	// In a real implementation, this would call the actual decode logic
	// For now, we'll simulate it
	go func() {
		// Simulate processing time
		time.Sleep(1 * time.Second)

		// Show success message
		mv.showStatus("✅ Decode completed")

		// In real implementation, you would parse the results and display them
		// For single decode page
		resultsSection := mv.singleDecodePage.GetItem(1).(*tview.Flex)
		resultsView := resultsSection.GetItem(0).(*tview.TextView)

		resultsView.SetText(fmt.Sprintf("Decode Results for:\n%s\n\n[Results would be displayed here]\n", code))
	}()
}

// showStatus updates the status bar
func (mv *MainView) showStatus(message string) {
	mv.statusBar.SetText(message)
}

// updateStatusBar updates the status bar with current info
func (mv *MainView) updateStatusBar() {
	currentPage := mv.GetPageName()
	statusText := fmt.Sprintf("BL4 Item Serial Code Codec | %s | F1: Help | F5: Refresh | Ctrl+Tab: Switch Page | Ctrl+Q: Quit", currentPage)
	mv.statusBar.SetText(statusText)
}

// ShowHelp displays the help modal
func (mv *MainView) ShowHelp() {
	mv.app.SetRoot(mv.helpModal, true)
}

// ShowAbout displays the about modal
func (mv *MainView) ShowAbout() {
	mv.app.SetRoot(mv.aboutModal, true)
}

// Refresh refreshes the current view
func (mv *MainView) Refresh() {
	// Simplified refresh - always refresh batch view if available
	if mv.batchView != nil {
		mv.batchView.Refresh()
	}

	mv.showStatus("🔄 View refreshed")
}

// SwitchToNextPage switches to the next page (simplified)
func (mv *MainView) SwitchToNextPage() {
	// Simplified page switching - just show status
	mv.showStatus("🔄 Page switching (simplified)")
	mv.updateStatusBar()
}

// GetPageName returns the current page name (simplified)
func (mv *MainView) GetPageName() string {
	// Simplified - return current context
	return "Batch Processing"
}