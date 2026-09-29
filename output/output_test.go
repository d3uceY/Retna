package output

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/d3uceY/Retna/color"
)

func TestColumnsAlignsStyledCells(t *testing.T) {
	// The middle column is the one that has to line up. Styled cells carry ANSI
	// escapes, so a naive byte count would push the following column around.
	rows := [][]string{
		{Pass.Render("PASS"), "alpha", "end"},
		{Fail.Render("FAIL"), "b", "end"},
		{Pass.Render("PASS"), "gamma", "end"},
	}
	out := Columns(rows)

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}

	// "PASS" or "FAIL" plus two spaces, then the middle column padded to
	// "alpha", then the two space separator.
	const wantPrefix = 4 + 2 + 5 + 2
	for i, line := range lines {
		at := strings.Index(line, "end")
		if at < 0 {
			t.Fatalf("line %d has no marker: %q", i, line)
		}
		if got := lipgloss.Width(line[:at]); got != wantPrefix {
			t.Errorf("line %d puts column three at %d, want %d (%q)", i, got, wantPrefix, line)
		}
	}
}

func TestTableHasHeaderAndRule(t *testing.T) {
	out := Table(
		[]string{"Color", "Ratio"},
		[][]string{
			{"#000000", "21.00:1"},
			{"#FFFFFF", "1.00:1"},
		},
	)
	if !strings.Contains(out, "Color") || !strings.Contains(out, "Ratio") {
		t.Error("table is missing its header")
	}
	if !strings.Contains(out, "─") {
		t.Error("table is missing its rule")
	}
	if !strings.Contains(out, "#000000") || !strings.Contains(out, "1.00:1") {
		t.Error("table is missing its rows")
	}
}

func TestSwatchCarriesTheColor(t *testing.T) {
	c, err := color.Parse("#3498DB")
	if err != nil {
		t.Fatal(err)
	}
	if got := lipgloss.Width(Swatch(c)); got != 2 {
		t.Errorf("swatch width = %d, want 2", got)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, convertView()); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("JSON output should end with a newline")
	}

	var decoded struct {
		Input  string            `json:"input"`
		Values map[string]string `json:"values"`
	}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded.Input != "#1A73E8" {
		t.Errorf("input = %q", decoded.Input)
	}
	if decoded.Values["hex"] != "#1A73E8" {
		t.Errorf("hex = %q", decoded.Values["hex"])
	}
}

func convertView() ConvertJSON {
	return ConvertJSON{
		Input:  "#1A73E8",
		Values: map[string]string{"hex": "#1A73E8"},
	}
}

func TestNewColorJSON(t *testing.T) {
	c, err := color.Parse("#3498DB")
	if err != nil {
		t.Fatal(err)
	}
	view := NewColorJSON("#3498DB", c)

	if view.Hex != "#3498DB" {
		t.Errorf("hex = %q", view.Hex)
	}
	if view.RGB != [3]int{52, 152, 219} {
		t.Errorf("rgb = %v", view.RGB)
	}
	if view.Alpha != 1 {
		t.Errorf("alpha = %v", view.Alpha)
	}
	if view.Luminance != 0.283 {
		t.Errorf("luminance = %v, want 0.283", view.Luminance)
	}
	if view.HSL[0] != 204.0719 || view.HSL[1] != 0.6987 {
		t.Errorf("hsl = %v", view.HSL)
	}
	if view.OKLCH[0] != 0.6531 {
		t.Errorf("oklch = %v", view.OKLCH)
	}

	// color(srgb ...) components are the same sRGB values scaled to 0..1.
	for i, want := range []float64{52, 152, 219} {
		if math.Abs(view.SRGB[i]-want/255) > 1e-4 {
			t.Errorf("srgb = %v, want it to match rgb %v", view.SRGB, view.RGB)
			break
		}
	}

	// The remaining gamuts all contain sRGB, so the color cannot fall outside
	// 0..1 in any of them.
	for name, group := range map[string][3]float64{
		"display-p3":   view.DisplayP3,
		"a98-rgb":      view.A98RGB,
		"prophoto-rgb": view.ProPhoto,
		"rec2020":      view.Rec2020,
	} {
		for i, v := range group {
			if v < -1e-6 || v > 1+1e-6 {
				t.Errorf("%s[%d] = %v, outside 0..1", name, i, v)
			}
		}
	}

	// Every group has to be present and finite so consumers can rely on them.
	for name, group := range map[string][3]float64{
		"hsl": view.HSL, "hsv": view.HSV, "hwb": view.HWB,
		"lab": view.Lab, "lch": view.LCH, "oklab": view.OKLab, "oklch": view.OKLCH,
		"srgb": view.SRGB, "display-p3": view.DisplayP3,
		"a98-rgb": view.A98RGB, "prophoto-rgb": view.ProPhoto, "rec2020": view.Rec2020,
	} {
		for i, v := range group {
			if v != v {
				t.Errorf("%s[%d] is NaN", name, i)
			}
		}
	}
}

func TestColorJSONOmitsEmptyAlgorithmBlocks(t *testing.T) {
	raw, err := json.Marshal(ContrastJSON{Foreground: "#000000", Background: "#FFFFFF", Contrast: 21})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "wcag") || strings.Contains(string(raw), "apca") {
		t.Errorf("unset algorithm blocks should be omitted, got %s", raw)
	}
}
