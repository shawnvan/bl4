package common

// GetHelpText returns the help text for the TUI
func GetHelpText() string {
	return `BL4 Item Serial Code Codec - Help

NAVIGATION:
• Tab: Switch between input fields
• Arrow Keys: Navigate lists and tables
• Enter: Confirm action/Start processing
• Escape: Cancel current operation

PAGES:
• F1: Show this help
• F2: Show about information
• F5: Refresh current view
• Ctrl+Q: Quit application

SINGLE DECODE:
• Enter serial code starting with '@'
• Press Enter or click Decode to process
• Results show item details and structure

BATCH PROCESSING:
• Add multiple codes (one per line)
• Load codes from file
• Start batch processing with real-time progress
• Save results to file
• Click results to view in single decode mode

KEYBOARD SHORTCUTS:
• Ctrl+S: Save results (when available)
• Ctrl+C: Cancel processing
• Ctrl+N: New/Clear input
• Ctrl+O: Open file
• Ctrl+R: Refresh

TIPS:
• Serial codes should start with '@'
• Use batch processing for multiple codes
• Results can be exported in JSON, CSV, or TXT formats
• Processing happens in real-time with progress tracking`
}

// GetAboutText returns the about text for the TUI
func GetAboutText() string {
	return `BL4 Item Serial Code Codec v1.0

A comprehensive tool for decoding and encoding Borderlands 4
item serial codes with advanced analysis capabilities.

FEATURES:
• Real-time batch processing
• Multiple interface modes (CLI, TUI, API)
• Progress tracking and cancellation
• Export capabilities
• Pattern analysis tools
• Performance monitoring

TECHNOLOGY:
• Built with Go 1.21+
• TUI framework: tview
• API framework: Gin
• Logging: Logrus
• Testing: Testify

LICENSE:
MIT License - See LICENSE file for details

SOURCE:
https://github.com/shawnvan/bl4

For support and documentation, visit the project repository.`
}