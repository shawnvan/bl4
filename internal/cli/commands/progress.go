package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressTracker displays a progress bar for batch operations
type ProgressTracker struct {
	bar         *progressbar.ProgressBar
	total       int
	current     int
	startTime   time.Time
	isRunning   bool
	stopChannel chan struct{}
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		total:       total,
		startTime:   time.Now(),
		stopChannel: make(chan struct{}),
	}
}

// Start begins displaying the progress bar
func (pt *ProgressTracker) Start() {
	pt.isRunning = true

	// Create progress bar with custom options
	pt.bar = progressbar.NewOptions64(
		int64(pt.total),
		progressbar.OptionSetDescription("Processing"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionFullWidth(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintf(os.Stderr, "\n")
		}),
	)

	// Start ticker for progress updates
	go pt.ticker()
}

// Stop stops the progress tracker
func (pt *ProgressTracker) Stop() {
	if !pt.isRunning {
		return
	}
	close(pt.stopChannel)
	pt.isRunning = false
	if pt.bar != nil {
		pt.bar.Finish()
	}
}

// Complete marks the progress bar as complete with statistics
func (pt *ProgressTracker) Complete(successful, failed int) {
	if pt.bar != nil {
		// Update final statistics
		pt.bar.Set64(int64(successful + failed))
		pt.bar.Finish()

		// Show final statistics
		duration := time.Since(pt.startTime)
		rate := float64(successful+failed) / duration.Seconds()

		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "✅ Batch processing completed!\n")
		fmt.Fprintf(os.Stderr, "📊 Statistics:\n")
		fmt.Fprintf(os.Stderr, "   Total items:     %d\n", pt.total)
		fmt.Fprintf(os.Stderr, "   Successful:      %d\n", successful)
		fmt.Fprintf(os.Stderr, "   Failed:          %d\n", failed)
		fmt.Fprintf(os.Stderr, "   Duration:        %v\n", duration.Round(time.Millisecond))
		fmt.Fprintf(os.Stderr, "   Rate:            %.2f items/sec\n", rate)

		if failed > 0 {
			fmt.Fprintf(os.Stderr, "   Success rate:     %.1f%%\n", float64(successful)/float64(pt.total)*100)
		}
	}
}

// Update updates the progress bar
func (pt *ProgressTracker) Update(current int) {
	pt.current = current
	if pt.bar != nil && pt.isRunning {
		pt.bar.Set64(int64(current))
	}
}

// ticker updates the progress bar periodically
func (pt *ProgressTracker) ticker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if pt.bar != nil {
				// Calculate elapsed time and rate
				elapsed := time.Since(pt.startTime)
				if elapsed > 0 {
					// Use simpler description without accessing Current field
					pt.bar.Describe(fmt.Sprintf("Processing item %d", pt.current))
				}
			}
		case <-pt.stopChannel:
			return
		}
	}
}

// IsTerminal checks if the output is a terminal
func IsTerminal() bool {
	// Simple terminal detection - assume we're in a terminal
	// This is a simplified version that doesn't require external dependencies
	return true
}

// SimpleProgressTracker provides a simple text-based progress indicator
type SimpleProgressTracker struct {
	total     int
	current   int
	startTime time.Time
}

// NewSimpleProgressTracker creates a simple progress tracker
func NewSimpleProgressTracker(total int) *SimpleProgressTracker {
	return &SimpleProgressTracker{
		total:     total,
		startTime: time.Now(),
	}
}

// Update updates the progress
func (spt *SimpleProgressTracker) Update(current int) {
	spt.current = current

	// Only update if we've made significant progress
	if current%10 == 0 || current == spt.total {
		percentage := float64(current) / float64(spt.total) * 100
		elapsed := time.Since(spt.startTime)
		rate := float64(current) / elapsed.Seconds()

		barWidth := 50
		completed := int(percentage / 100 * float64(barWidth))
	remaining := barWidth - completed

		bar := strings.Repeat("█", completed) + strings.Repeat("░", remaining)
		fmt.Printf("\r[%s] %d/%d (%.1f%%) %.1f items/sec", bar, current, spt.total, percentage, rate)

		if current == spt.total {
			fmt.Println() // New line when complete
		}
	}
}

// Complete marks the simple progress as complete
func (spt *SimpleProgressTracker) Complete(successful, failed int) {
	duration := time.Since(spt.startTime)
	rate := float64(spt.total) / duration.Seconds()

	fmt.Printf("\n✅ Batch processing completed!\n")
	fmt.Printf("📊 Statistics:\n")
	fmt.Printf("   Total items:     %d\n", spt.total)
	fmt.Printf("   Successful:      %d\n", successful)
	fmt.Printf("   Failed:          %d\n", failed)
	fmt.Printf("   Duration:        %v\n", duration.Round(time.Millisecond))
	fmt.Printf("   Rate:            %.2f items/sec\n", rate)

	if failed > 0 {
		fmt.Printf("   Success rate:     %.1f%%\n", float64(successful)/float64(spt.total)*100)
	}
}