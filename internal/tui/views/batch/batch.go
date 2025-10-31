package batch

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/shawnvan/bl4/pkg/logger"
)

// BatchView represents the batch processing view
type BatchView struct {
	layout       *tview.Flex
	inputArea    *tview.Flex
	controlArea  *tview.Flex
	resultsArea  *tview.Flex
	statusArea   *tview.Flex

	// Input components
	inputField   *tview.TextArea
	fileSelector *tview.DropDown
	addButton    *tview.Button
	clearButton  *tview.Button
	loadButton   *tview.Button

	// Control components
	startButton    *tview.Button
	cancelButton   *tview.Button
	saveButton     *tview.Button
	progressBar    *tview.TextView
	statusText     *tview.TextView

	// Results components
	resultsTable   *tview.Table
	resultsInfo    *tview.TextView

	// Status components
	statusBar      *tview.TextView
	batchIDText    *tview.TextView

	// Data
	codes          []string
	currentBatchID string
	isProcessing   bool
	navigateToSingle func()
}

// NewBatchView creates a new batch view
func NewBatchView() *BatchView {
	bv := &BatchView{
		codes:        make([]string, 0),
		isProcessing: false,
	}

	bv.createLayout()
	bv.setupEventHandlers()

	return bv
}

// GetLayout returns the main layout
func (bv *BatchView) GetLayout() tview.Primitive {
	return bv.layout
}

// SetNavigateToSingle sets the navigation callback
func (bv *BatchView) SetNavigateToSingle(callback func()) {
	bv.navigateToSingle = callback
}

// Refresh refreshes the current view
func (bv *BatchView) Refresh() {
	bv.updateStatus("🔄 View refreshed")
	bv.updateResultsInfo()
}

// createLayout creates the main layout structure
func (bv *BatchView) createLayout() {
	bv.layout = tview.NewFlex()
	bv.layout.SetDirection(tview.FlexRow)

	// Create sections
	bv.createInputArea()
	bv.createControlArea()
	bv.createResultsArea()
	bv.createStatusArea()

	// Add sections to layout
	bv.layout.AddItem(bv.inputArea, 0, 1, false)
	bv.layout.AddItem(bv.controlArea, 0, 1, false)
	bv.layout.AddItem(bv.resultsArea, 0, 1, true)
	bv.layout.AddItem(bv.statusArea, 1, 1, false)
}

// createInputArea creates the input section
func (bv *BatchView) createInputArea() {
	bv.inputArea = tview.NewFlex()
	bv.inputArea.SetDirection(tview.FlexColumn)
	bv.inputArea.SetBorder(true)
	bv.inputArea.SetTitle(" Input Codes ")

	// Left side - input field
	inputContainer := tview.NewFlex()
	inputContainer.SetDirection(tview.FlexRow)

	inputLabel := tview.NewTextView()
	inputLabel.SetText("Enter serial codes (one per line):")
	inputLabel.SetBorder(false)

	bv.inputField = tview.NewTextArea()
	bv.inputField.SetPlaceholder("@Ugy3L+2}TYg...\n@AnotherCode...\n@YetAnotherCode...")
	bv.inputField.SetBorder(true)

	inputContainer.AddItem(inputLabel, 1, 1, false)
	inputContainer.AddItem(bv.inputField, 0, 1, true)

	// Right side - file operations
	fileContainer := tview.NewFlex()
	fileContainer.SetDirection(tview.FlexRow)

	fileLabel := tview.NewTextView()
	fileLabel.SetText("File Operations:")
	fileLabel.SetBorder(false)

	bv.fileSelector = tview.NewDropDown()
	bv.fileSelector.SetLabel("Load from file: ")
	bv.fileSelector.SetOptions([]string{"Browse files...", "recent1.txt", "recent2.txt"}, nil)

	bv.loadButton = tview.NewButton("Load File")
	bv.loadButton.SetSelectedFunc(bv.handleLoadFile)

	clearFileButton := tview.NewButton("Clear File")
	clearFileButton.SetSelectedFunc(func() {
		bv.fileSelector.SetCurrentOption(0)
	})

	fileContainer.AddItem(fileLabel, 1, 1, false)
	fileContainer.AddItem(bv.fileSelector, 1, 1, false)
	fileContainer.AddItem(tview.NewBox(), 1, 1, false)
	fileContainer.AddItem(bv.loadButton, 1, 1, false)
	fileContainer.AddItem(clearFileButton, 1, 1, false)

	// Input controls
	inputControls := tview.NewFlex()
	bv.addButton = tview.NewButton("Add Codes")
	bv.addButton.SetSelectedFunc(bv.handleAddCodes)

	bv.clearButton = tview.NewButton("Clear All")
	bv.clearButton.SetSelectedFunc(bv.handleClearAll)

	inputControls.AddItem(bv.addButton, 0, 1, false)
	inputControls.AddItem(bv.clearButton, 0, 1, false)

	// Assemble input area
	bv.inputArea.AddItem(inputContainer, 0, 3, true)
	bv.inputArea.AddItem(fileContainer, 0, 1, false)
	bv.inputArea.AddItem(inputControls, 1, 1, false)
}

// createControlArea creates the control section
func (bv *BatchView) createControlArea() {
	bv.controlArea = tview.NewFlex()
	bv.controlArea.SetDirection(tview.FlexColumn)
	bv.controlArea.SetBorder(true)
	bv.controlArea.SetTitle(" Batch Control ")

	// Progress section
	progressContainer := tview.NewFlex()
	progressContainer.SetDirection(tview.FlexColumn)

	progressLabel := tview.NewTextView()
	progressLabel.SetText("Progress:")
	progressLabel.SetBorder(false)

	bv.progressBar = tview.NewTextView()
	bv.progressBar.SetText("Ready to process")
	bv.progressBar.SetDynamicColors(true)

	bv.statusText = tview.NewTextView()
	bv.statusText.SetText("")
	bv.statusText.SetDynamicColors(true)

	progressContainer.AddItem(progressLabel, 1, 1, false)
	progressContainer.AddItem(bv.progressBar, 1, 1, false)
	progressContainer.AddItem(bv.statusText, 2, 1, false)

	// Control buttons
	buttonContainer := tview.NewFlex()

	bv.startButton = tview.NewButton("Start Batch")
	bv.startButton.SetSelectedFunc(bv.handleStartBatch)

	bv.cancelButton = tview.NewButton("Cancel")
	bv.cancelButton.SetSelectedFunc(bv.handleCancelBatch)
	bv.cancelButton.SetDisabled(true)

	bv.saveButton = tview.NewButton("Save Results")
	bv.saveButton.SetSelectedFunc(bv.handleSaveResults)
	bv.saveButton.SetDisabled(true)

	buttonContainer.AddItem(bv.startButton, 0, 1, false)
	buttonContainer.AddItem(bv.cancelButton, 0, 1, false)
	buttonContainer.AddItem(tview.NewBox(), 1, 1, false)
	buttonContainer.AddItem(bv.saveButton, 0, 1, false)

	// Batch info
	bv.batchIDText = tview.NewTextView()
	bv.batchIDText.SetText("Batch ID: Not started")
	bv.batchIDText.SetBorder(true)

	bv.controlArea.AddItem(progressContainer, 0, 1, false)
	bv.controlArea.AddItem(buttonContainer, 1, 1, false)
	bv.controlArea.AddItem(bv.batchIDText, 1, 1, false)
}

// createResultsArea creates the results section
func (bv *BatchView) createResultsArea() {
	bv.resultsArea = tview.NewFlex()
	bv.resultsArea.SetDirection(tview.FlexColumn)
	bv.resultsArea.SetBorder(true)
	bv.resultsArea.SetTitle(" Results ")

	// Results info
	bv.resultsInfo = tview.NewTextView()
	bv.resultsInfo.SetText("No results yet. Add codes and start processing.")
	bv.resultsInfo.SetBorder(true)
	bv.resultsInfo.SetDynamicColors(true)

	// Results table
	bv.resultsTable = tview.NewTable()
	bv.resultsTable.SetBorders(true)
	bv.resultsTable.SetFixed(1, 0) // Fix header row
	bv.resultsTable.SetSelectable(true, false)
	bv.resultsTable.SetSelectedFunc(func(row, column int) {
		bv.handleResultSelection(row)
	})

	// Set up table headers
	bv.setupTableHeaders()

	bv.resultsArea.AddItem(bv.resultsInfo, 2, 1, false)
	bv.resultsArea.AddItem(bv.resultsTable, 0, 1, true)
}

// createStatusArea creates the status bar
func (bv *BatchView) createStatusArea() {
	bv.statusArea = tview.NewFlex()
	bv.statusArea.SetDirection(tview.FlexColumn)

	bv.statusBar = tview.NewTextView()
	bv.statusBar.SetDynamicColors(true)
	bv.statusBar.SetText("Ready | F1: Help | Ctrl+S: Save | Ctrl+C: Cancel | Enter: Start")

	bv.statusArea.AddItem(bv.statusBar, 1, 1, false)
}

// setupTableHeaders sets up the results table headers
func (bv *BatchView) setupTableHeaders() {
	headers := []string{"#", "Serial Code", "Status", "Level", "Type", "Manufacturer", "Result"}

	for col, header := range headers {
		cell := tview.NewTableCell(header)
		cell.SetAlign(tview.AlignCenter)
		cell.SetSelectable(false)
		bv.resultsTable.SetCell(0, col, cell)
	}
}

// setupEventHandlers sets up keyboard shortcuts
func (bv *BatchView) setupEventHandlers() {
	// Keyboard shortcuts simplified for now
}

// Event handlers
func (bv *BatchView) handleAddCodes() {
	text := strings.TrimSpace(bv.inputField.GetText())
	if text == "" {
		bv.updateStatus("❌ No codes to add")
		return
	}

	lines := strings.Split(text, "\n")
	addedCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, "@") {
			bv.codes = append(bv.codes, line)
			addedCount++
		}
	}

	if addedCount > 0 {
		bv.inputField.SetText("", true)
		bv.updateResultsInfo()
		bv.updateStatus(fmt.Sprintf("✅ Added %d codes to batch", addedCount))
	} else {
		bv.updateStatus("⚠️  No valid codes found (codes should start with '@')")
	}
}

func (bv *BatchView) handleClearAll() {
	bv.codes = make([]string, 0)
	bv.inputField.SetText("", true)
	bv.clearResults()
	bv.updateResultsInfo()
	bv.updateStatus("🗑️  Cleared all codes")
}

func (bv *BatchView) handleLoadFile() {
	// For now, simulate file loading
	// In a real implementation, this would open a file browser
	sampleCodes := []string{
		"@Ugy3L+2}TYg==",
		"@AnotherSampleCode==",
		"@YetAnotherCode==",
	}

	bv.inputField.SetText(strings.Join(sampleCodes, "\n"), true)
	bv.updateStatus("📁 Sample codes loaded (file browser not implemented)")
}

func (bv *BatchView) handleStartBatch() {
	if len(bv.codes) == 0 {
		bv.updateStatus("❌ No codes to process")
		return
	}

	bv.isProcessing = true
	bv.currentBatchID = fmt.Sprintf("batch_%d", time.Now().Unix())

	// Update UI state
	bv.startButton.SetDisabled(true)
	bv.cancelButton.SetDisabled(false)
	bv.saveButton.SetDisabled(true)

	bv.batchIDText.SetText(fmt.Sprintf("Batch ID: %s", bv.currentBatchID))
	bv.updateStatus("🔄 Starting batch processing...")
	bv.clearResults()

	// Start processing in goroutine
	go bv.simulateBatchProcessing()
}

func (bv *BatchView) handleCancelBatch() {
	if !bv.isProcessing {
		return
	}

	bv.isProcessing = false
	bv.updateStatus("⏹️  Cancelling batch processing...")

	// Update UI state
	bv.startButton.SetDisabled(false)
	bv.cancelButton.SetDisabled(true)

	bv.statusText.SetText("[red]Cancelled[/red]")
	bv.progressBar.SetText("[red]Processing cancelled[/red]")
}

func (bv *BatchView) handleSaveResults() {
	if bv.resultsTable.GetRowCount() <= 1 {
		bv.updateStatus("❌ No results to save")
		return
	}

	bv.updateStatus("💾 Saving results... (not implemented)")
	// In a real implementation, this would save to file
}

func (bv *BatchView) handleResultSelection(row int) {
	if row <= 0 || row >= bv.resultsTable.GetRowCount() {
		return
	}

	// Get the serial code from the selected row
	codeCell := bv.resultsTable.GetCell(row, 1)
	if codeCell != nil {
		code := codeCell.Text

		// Navigate to single decode view with this code
		if bv.navigateToSingle != nil {
			bv.navigateToSingle()
		}

		bv.updateStatus(fmt.Sprintf("🔍 Opened '%s' in single decode view", code))
	}
}

// simulateBatchProcessing simulates the batch processing
func (bv *BatchView) simulateBatchProcessing() {
	total := len(bv.codes)

	for i, code := range bv.codes {
		if !bv.isProcessing {
			break
		}

		// Update progress
		progress := float64(i+1) / float64(total) * 100
		bv.updateProgress(progress, i+1, total)

		// Simulate processing time
		time.Sleep(500 * time.Millisecond)

		// Add result to table
		bv.addResultToTable(i+1, code, "Success")

		// Update status
		bv.statusText.SetText(fmt.Sprintf("Processing: %s", code))
	}

	if bv.isProcessing {
		bv.completeBatch()
	}
}

// addResultToTable adds a result to the results table
func (bv *BatchView) addResultToTable(index int, code, status string) {
	row := index

	// Index
	cell := tview.NewTableCell(fmt.Sprintf("%d", index))
	cell.SetAlign(tview.AlignCenter)
	bv.resultsTable.SetCell(row, 0, cell)

	// Serial Code
	codeCell := tview.NewTableCell(code)
	codeCell.SetSelectable(true)
	bv.resultsTable.SetCell(row, 1, codeCell)

	// Status
	statusCell := tview.NewTableCell(status)
	statusCell.SetAlign(tview.AlignCenter)
	bv.resultsTable.SetCell(row, 2, statusCell)

	// Simulated data
	bv.resultsTable.SetCell(row, 3, tview.NewTableCell("24"))
	bv.resultsTable.SetCell(row, 4, tview.NewTableCell("Pistol"))
	bv.resultsTable.SetCell(row, 5, tview.NewTableCell("Maliwan"))
	bv.resultsTable.SetCell(row, 6, tview.NewTableCell("24, 0, 1, 50| 2, 3379|| {76} {2} {3}|"))
}

// completeBatch completes the batch processing
func (bv *BatchView) completeBatch() {
	bv.isProcessing = false

	// Update UI state
	bv.startButton.SetDisabled(false)
	bv.cancelButton.SetDisabled(true)
	bv.saveButton.SetDisabled(false)

	bv.progressBar.SetText("[green]✅ Processing completed[/green]")
	bv.statusText.SetText("All codes processed successfully")
	bv.updateStatus(fmt.Sprintf("✅ Batch completed: %d codes processed", len(bv.codes)))
}

// Helper methods
func (bv *BatchView) updateProgress(percentage float64, current, total int) {
	bar := ""
	barWidth := 30
	filled := int(percentage / 100 * float64(barWidth))

	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	bv.progressBar.SetText(fmt.Sprintf("[green]%s[white] %.1f%% (%d/%d)", bar, percentage, current, total))
}

func (bv *BatchView) updateStatus(message string) {
	bv.statusBar.SetText(message)
	logger.Sugar().Infof("TUI Batch: %s", message)
}

func (bv *BatchView) updateResultsInfo() {
	if len(bv.codes) == 0 {
		bv.resultsInfo.SetText("No codes yet. Add codes to start processing.")
	} else {
		bv.resultsInfo.SetText(fmt.Sprintf("[yellow]%d[white] codes ready for processing", len(bv.codes)))
	}
}

func (bv *BatchView) clearResults() {
	// Clear all rows except header
	for bv.resultsTable.GetRowCount() > 1 {
		bv.resultsTable.RemoveRow(1)
	}
}

func (bv *BatchView) showHelp() {
	// Show help modal (simplified implementation)
	bv.updateStatus("📖 Help displayed - Batch processing interface for BL4 serial codes")
	logger.Sugar().Info("TUI Batch Help displayed")
}