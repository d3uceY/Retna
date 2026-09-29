// Package output renders results for the terminal and for JSON consumers.
package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/d3uceY/Retna/color"
)

// The shared palette. Every command pulls its colors from here.
const (
	ColAccent = "#5F9EA0"
	ColText   = "#CCCCCC"
	ColDim    = "#666666"
	ColPass   = "#7EC8A4"
	ColFail   = "#CC6666"
	ColWarn   = "#C8A96E"
)

var (
	Accent = lipgloss.NewStyle().Foreground(lipgloss.Color(ColAccent))
	Text   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColText))
	Dim    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColDim))
	Pass   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColPass))
	Fail   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColFail))
	Warn   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColWarn))
)

// Heading renders a section title.
func Heading(s string) string { return Accent.Bold(true).Render(s) }

// Label renders secondary text.
func Label(s string) string { return Dim.Render(s) }

// Verdict renders a pass or fail label.
func Verdict(ok bool) string {
	if ok {
		return Pass.Render("PASS")
	}
	return Fail.Render("FAIL")
}

// Swatch renders a two-cell block filled with the color.
func Swatch(c color.Color) string {
	r, g, b := c.Channels()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))).
		Render("  ")
}

// Columns renders rows padded to a common width with no header.
func Columns(rows [][]string) string {
	return render(rows, columnWidths(rows))
}

// Table renders a header, a rule and then the rows.
func Table(headers []string, rows [][]string) string {
	widths := columnWidths(append([][]string{headers}, rows...))
	styled := make([]string, len(headers))
	for i, h := range headers {
		styled[i] = Accent.Render(h)
	}

	var b strings.Builder
	b.WriteString(render([][]string{styled}, widths))

	rule := make([]string, len(widths))
	for i, w := range widths {
		rule[i] = strings.Repeat("─", w)
	}
	b.WriteString(Dim.Render(strings.Join(rule, "  ")))
	b.WriteString("\n")

	return b.String() + render(rows, widths)
}

// columnWidths measures every column with ANSI escapes discounted, which is
// what keeps colored cells aligned.
func columnWidths(rows [][]string) []int {
	var widths []int
	for _, row := range rows {
		if len(row) > len(widths) {
			widths = append(widths, make([]int, len(row)-len(widths))...)
		}
	}
	for _, row := range rows {
		for i, cell := range row {
			if w := lipgloss.Width(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func render(rows [][]string, widths []int) string {
	var b strings.Builder
	for _, row := range rows {
		cells := make([]string, len(widths))
		for i := range widths {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			pad := widths[i] - lipgloss.Width(cell)
			if pad < 0 {
				pad = 0
			}
			cells[i] = cell + strings.Repeat(" ", pad)
		}
		b.WriteString(strings.TrimRight(strings.Join(cells, "  "), " "))
		b.WriteString("\n")
	}
	return b.String()
}
