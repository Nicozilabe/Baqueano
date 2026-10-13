package ui

import "github.com/charmbracelet/lipgloss"

// Theme colors - aligned with the project website palette (docs/styles.css)
var (
	// Primary interface color - website accent (green)
	PrimaryColor = "#4ade80" // flag: --accent
	// PrimaryDimColor is the darker accent, used for borders and hover states
	PrimaryDimColor = "#22c55e" // flag: --accent-dim

	// Link / highlight color
	LinkColor = "#7dd3fc" // flag: --link

	// Foreground colors
	TextColor    = "#e6edf3" // flag: --text
	TextDimColor = "#9fb0c0" // flag: --text-dim

	// Surface and border colors
	SurfaceColor  = "#111823" // flag: --bg-soft
	ElevatedColor = "#161f2c" // flag: --bg-elev
	BorderColor   = "#232f3f" // flag: --border

	// Semantic colors
	SecondaryColor = TextDimColor // Gray (dim text)
	ErrorColor     = "#f87171"    // Red
	SuccessColor   = PrimaryColor // Green
)

// Styles struct centralizes all lipgloss styles
type Styles struct {
	// Layout
	App    lipgloss.Style
	Header lipgloss.Style

	// Search styles
	SearchFocused   lipgloss.Style
	SearchUnfocused lipgloss.Style

	// Table styles
	TableFocused   lipgloss.Style
	TableUnfocused lipgloss.Style
	Selected       lipgloss.Style

	// Info and help styles
	SortInfo lipgloss.Style
	HelpText lipgloss.Style

	// Error and confirmation styles
	Error     lipgloss.Style
	ErrorText lipgloss.Style

	// Form styles (for add/edit forms)
	FormTitle     lipgloss.Style
	FormField     lipgloss.Style
	FormHelp      lipgloss.Style
	FormContainer lipgloss.Style
	Label         lipgloss.Style
	FocusedLabel  lipgloss.Style
	HelpSection   lipgloss.Style
}

// NewStyles creates a new Styles struct with the given terminal width
func NewStyles(width int) Styles {
	return Styles{
		// Main app container
		App: lipgloss.NewStyle().
			Padding(1),

		// Header style
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimaryColor)).
			Bold(true).
			Align(lipgloss.Center),

		// Search styles
		SearchFocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)).
			Padding(0, 1),

		SearchUnfocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(SecondaryColor)).
			Padding(0, 1),

		// Table styles
		TableFocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)),

		TableUnfocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(SecondaryColor)),

		// Style for selected items
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0b0f14")).
			Background(lipgloss.Color(PrimaryColor)).
			Bold(false),

		// Info styles
		SortInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color(SecondaryColor)),

		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(SecondaryColor)),

		// Error style
		Error: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ErrorColor)).
			Padding(1, 2),

		// Error text style (no border, just red text)
		ErrorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ErrorColor)).
			Bold(true),

		// Form styles
		FormTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0b0f14")).
			Background(lipgloss.Color(PrimaryColor)).
			Padding(0, 1),

		FormField: lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimaryColor)),

		FormHelp: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextDimColor)),

		FormContainer: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)).
			Padding(1, 2),

		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color(SecondaryColor)),

		FocusedLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimaryColor)),

		HelpSection: lipgloss.NewStyle().
			Padding(0, 2),
	}
}

// Application ASCII title
const asciiTitle = `
 _____ _____ __ __ _____
|   __|   __|  |  |     |
|__   |__   |     | | | |
|_____|_____|__|__|_|_|_|
`
