package tui

import (
	"fmt"
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

// MinTerminalSize defines minimum required terminal dimensions
const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 20
)

// IsTerminalSizeAdequate checks if terminal is large enough for TUI
func IsTerminalSizeAdequate(width, height int) bool {
	return width >= MinTerminalWidth && height >= MinTerminalHeight
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

// PanelDimensions holds calculated panel dimensions
type PanelDimensions struct {
	LeftWidth  int
	RightWidth int
	UseSingle  bool
}

// CalculatePanelDimensions calculates optimal panel dimensions for dual-panel layout
func CalculatePanelDimensions(windowWidth int) PanelDimensions {
	minLeftWidth := 30  // Minimum width for left panel
	minRightWidth := 40 // Minimum width for right panel

	leftWidth := int(float64(windowWidth) * 0.4) // 40% for left panel
	if leftWidth < minLeftWidth {
		leftWidth = minLeftWidth
	}

	rightWidth := windowWidth - leftWidth - 3 // Remaining for right panel (minus border)
	if rightWidth < minRightWidth {
		rightWidth = minRightWidth
		leftWidth = windowWidth - rightWidth - 3
		if leftWidth < minLeftWidth {
			// If we can't fit both panels properly, use single panel
			return PanelDimensions{UseSingle: true}
		}
	}

	return PanelDimensions{
		LeftWidth:  leftWidth,
		RightWidth: rightWidth,
		UseSingle:  false,
	}
}

// RenderTerminalTooSmallMessage renders the standard "terminal too small" message
func RenderTerminalTooSmallMessage(styles *CommonStyles, width, height int) string {
	return styles.Header.Render("Terminal too small") +
		"\n\nMinimum size: 80x20" +
		fmt.Sprintf("\nCurrent size: %dx%d", width, height) +
		"\n\nPress q to quit"
}

// CreateBorderedPanel creates a panel with consistent border styling
func CreateBorderedPanel(width, height int, borderColor string) lipgloss.Style {
	adjustedWidth := max(20, width)   // Ensure minimum width for border
	adjustedHeight := max(10, height) // Ensure minimum height

	return lipgloss.NewStyle().
		Width(adjustedWidth).
		Height(adjustedHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1)
}

// GetThemeColors returns appropriate colors for current theme
func GetThemeColors() (headerColor, borderColor, helpColor, selectedFg, selectedBg, scrollColor string) {
	if isDarkMode() {
		mocha := catppuccin.Mocha
		return mocha.Mauve().Hex, mocha.Surface2().Hex, mocha.Subtext1().Hex,
			mocha.Base().Hex, mocha.Blue().Hex, mocha.Subtext1().Hex
	}

	latte := catppuccin.Latte
	return latte.Mauve().Hex, latte.Surface2().Hex, latte.Subtext1().Hex,
		latte.Base().Hex, latte.Blue().Hex, latte.Subtext1().Hex
}

// RenderHeader renders a standard TUI header with theme-appropriate styling
func RenderHeader(title string, windowWidth int) string {
	headerColor, _, _, _, _, _ := GetThemeColors()
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(headerColor)).
		Width(windowWidth).
		Align(lipgloss.Center).
		MarginBottom(1)

	return headerStyle.Render(title)
}

// RenderHelpText renders navigation help text with consistent styling
func RenderHelpText(helpText string, maxWidth int) string {
	_, _, helpColor, _, _, _ := GetThemeColors()
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(helpColor)).
		Italic(true).
		Width(max(10, maxWidth))

	return helpStyle.Render(helpText)
}

// RenderScrollIndicator renders a scroll position indicator
func RenderScrollIndicator(current, total, maxWidth int) string {
	_, _, _, _, _, scrollColor := GetThemeColors()
	scrollInfo := fmt.Sprintf("[%d/%d]", current, total)

	scrollStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(scrollColor)).
		Align(lipgloss.Right).
		Width(max(10, maxWidth))

	return scrollStyle.Render(scrollInfo)
}

// ApplySelectionStyle applies selection styling to text
func ApplySelectionStyle(text string, isSelected bool, maxWidth int) string {
	if isSelected {
		_, _, _, selectedFg, selectedBg, _ := GetThemeColors()
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(selectedFg)).
			Background(lipgloss.Color(selectedBg)).
			Bold(true).
			Width(max(10, maxWidth))
		return style.Render("> " + text)
	}

	style := lipgloss.NewStyle().Width(max(10, maxWidth))
	return style.Render("  " + text)
}

// TruncateText truncates text to fit within maxWidth, adding ellipsis if needed
func TruncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	if maxWidth <= 3 {
		return text[:max(1, maxWidth)]
	}

	return text[:maxWidth-3] + "..."
}
