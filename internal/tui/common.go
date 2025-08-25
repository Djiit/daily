package tui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-isatty"
)

// CommonStyles contains shared styling for TUI components
type CommonStyles struct {
	Base          lipgloss.Style
	Header        lipgloss.Style
	SectionHeader lipgloss.Style
	Selected      lipgloss.Style
	Unselected    lipgloss.Style
	Help          lipgloss.Style
	Border        lipgloss.Style
	ScrollInfo    lipgloss.Style
	StatusBar     lipgloss.Style
	Description   lipgloss.Style
	URL           lipgloss.Style
	Tags          lipgloss.Style
	Time          lipgloss.Style
}

// isDarkMode detects if the terminal is using a dark theme
func isDarkMode() bool {
	// Check for explicit dark mode environment variables
	if theme := os.Getenv("THEME"); theme == "dark" {
		return true
	}
	if theme := os.Getenv("TERMINAL_THEME"); theme == "dark" {
		return true
	}

	// Check environment variables that indicate dark mode
	if colorScheme := os.Getenv("COLORFGBG"); colorScheme != "" {
		// COLORFGBG format is usually "foreground;background"
		// Dark themes typically have light foreground on dark background
		parts := strings.Split(colorScheme, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			// Background colors like 0-7 (especially 0, 1, 8) indicate dark themes
			return bg == "0" || bg == "1" || bg == "8"
		}
	}

	// Default to light mode if we can't determine
	return false
}

// NewCommonStyles creates a new set of common TUI styles
func NewCommonStyles() *CommonStyles {
	isDark := isDarkMode()

	if isDark {
		// Use Catppuccin Mocha colors for dark mode
		mocha := catppuccin.Mocha
		return &CommonStyles{
			Base: lipgloss.NewStyle().
				Padding(1),
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(mocha.Mauve().Hex)).
				Align(lipgloss.Center).
				MarginBottom(1),
			SectionHeader: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(mocha.Green().Hex)).
				MarginTop(1).
				MarginBottom(1),
			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(mocha.Base().Hex)).
				Background(lipgloss.Color(mocha.Blue().Hex)),
			Unselected: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext0().Hex)),
			Help: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext1().Hex)).
				Italic(true),
			Border: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(mocha.Surface2().Hex)),
			ScrollInfo: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext1().Hex)).
				Align(lipgloss.Right),
			StatusBar: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Text().Hex)).
				Background(lipgloss.Color(mocha.Surface0().Hex)).
				PaddingLeft(1).
				PaddingRight(1).
				Align(lipgloss.Center),
			Description: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext0().Hex)).
				Italic(true).
				MarginLeft(2),
			URL: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Sapphire().Hex)).
				Underline(true).
				MarginLeft(2),
			Tags: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Peach().Hex)).
				Italic(true).
				MarginLeft(2),
			Time: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext1().Hex)),
		}
	}

	// Light mode colors (default - Catppuccin Latte)
	latte := catppuccin.Latte
	return &CommonStyles{
		Base: lipgloss.NewStyle().
			Padding(1),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(latte.Mauve().Hex)).
			Align(lipgloss.Center).
			MarginBottom(1),
		SectionHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(latte.Green().Hex)).
			MarginTop(1).
			MarginBottom(1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(latte.Base().Hex)).
			Background(lipgloss.Color(latte.Blue().Hex)),
		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext0().Hex)),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext1().Hex)).
			Italic(true),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(latte.Surface2().Hex)),
		ScrollInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext1().Hex)).
			Align(lipgloss.Right),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Text().Hex)).
			Background(lipgloss.Color(latte.Surface0().Hex)).
			PaddingLeft(1).
			PaddingRight(1).
			Align(lipgloss.Center),
		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext0().Hex)).
			Italic(true).
			MarginLeft(2),
		URL: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Sapphire().Hex)).
			Underline(true).
			MarginLeft(2),
		Tags: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Peach().Hex)).
			Italic(true).
			MarginLeft(2),
		Time: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext1().Hex)),
	}
}

// IsTerminalCapable checks if the current environment supports TUI
func IsTerminalCapable() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// OpenURL opens the given URL in the default browser
func OpenURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// Navigation helpers

// ClampCursor ensures cursor stays within bounds
func ClampCursor(cursor, min, max int) int {
	if cursor < min {
		return min
	}
	if cursor > max {
		return max
	}
	return cursor
}

// UpdateViewport updates viewport offset to keep cursor visible
func UpdateViewport(cursor, viewportOffset, viewportHeight, totalItems int) int {
	if viewportHeight <= 0 {
		return viewportOffset
	}

	// Ensure cursor is visible in viewport
	if cursor < viewportOffset {
		viewportOffset = cursor
	} else if cursor >= viewportOffset+viewportHeight {
		viewportOffset = cursor - viewportHeight + 1
	}

	// Ensure viewport doesn't exceed bounds
	viewportOffset = max(0, viewportOffset)
	maxOffset := max(0, totalItems-viewportHeight)
	viewportOffset = min(viewportOffset, maxOffset)

	return viewportOffset
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
