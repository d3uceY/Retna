package output

import (
	"encoding/json"
	"io"
	"math"

	"github.com/d3uceY/Retna/color"
)

// LevelJSON is the pass state of the normal and large text rules.
type LevelJSON struct {
	Normal bool `json:"normal"`
	Large  bool `json:"large"`
}

// WCAGJSON is the WCAG verdict block.
type WCAGJSON struct {
	AA  LevelJSON `json:"aa"`
	AAA LevelJSON `json:"aaa"`
}

// ContrastJSON is one foreground and background measurement.
type ContrastJSON struct {
	Foreground string    `json:"foreground"`
	Background string    `json:"background"`
	Contrast   float64   `json:"contrast"`
	WCAG       *WCAGJSON `json:"wcag,omitempty"`
	APCA       *float64  `json:"apca,omitempty"`
}

// ContrastListJSON is one foreground measured against many backgrounds.
type ContrastListJSON struct {
	Foreground string         `json:"foreground"`
	Results    []ContrastJSON `json:"results"`
}

// ColorJSON is the structured breakdown of a single color. Angles are degrees,
// saturation style channels are 0..1, and Lab lightness is 0..100 while OKLab
// lightness is 0..1.
type ColorJSON struct {
	Input     string     `json:"input"`
	Hex       string     `json:"hex"`
	Alpha     float64    `json:"alpha"`
	RGB       [3]int     `json:"rgb"`
	HSL       [3]float64 `json:"hsl"`
	HSV       [3]float64 `json:"hsv"`
	HWB       [3]float64 `json:"hwb"`
	Lab       [3]float64 `json:"lab"`
	LCH       [3]float64 `json:"lch"`
	OKLab     [3]float64 `json:"oklab"`
	OKLCH     [3]float64 `json:"oklch"`
	Luminance float64    `json:"luminance"`
}

// ConvertJSON is the result of converting one color into named spaces.
type ConvertJSON struct {
	Input  string            `json:"input"`
	Values map[string]string `json:"values"`
}

// CandidateJSON is one suggested text color.
type CandidateJSON struct {
	Color    string  `json:"color"`
	Contrast float64 `json:"contrast"`
	Pass     bool    `json:"pass"`
}

// ReadableJSON is the result of asking for readable text colors.
type ReadableJSON struct {
	Background string          `json:"background"`
	Level      string          `json:"level"`
	Minimum    float64         `json:"minimum"`
	Candidates []CandidateJSON `json:"candidates"`
}

// FixSuggestionJSON is one passing color the fix command found.
type FixSuggestionJSON struct {
	Hex      string  `json:"hex"`
	Contrast float64 `json:"contrast"`
	Pass     bool    `json:"pass"`
}

// FixJSON is the result of fixing a color for a target ratio.
type FixJSON struct {
	Input       string              `json:"input"`
	Background  string              `json:"background"`
	Target      float64             `json:"target"`
	Contrast    float64             `json:"contrast"`
	Pass        bool                `json:"pass"`
	Suggestions []FixSuggestionJSON `json:"suggestions"`
}

// WriteJSON encodes v as indented JSON followed by a newline.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// NewColorJSON builds the structured view of a color.
func NewColorJSON(input string, c color.Color) ColorJSON {
	r, g, b := c.Channels()
	hslH, hslS, hslL := color.ToHSL(c)
	hsvH, hsvS, hsvV := color.ToHSV(c)
	hwbH, hwbW, hwbB := color.ToHWB(c)
	labL, labA, labB := color.ToLab(c)
	lchL, lchC, lchH := color.ToLCH(c)
	okL, okA, okB := color.ToOKLab(c)
	oklchL, oklchC, oklchH := color.ToOKLCH(c)

	return ColorJSON{
		Input:     input,
		Hex:       c.Hex(),
		Alpha:     round4(c.A),
		RGB:       [3]int{int(r), int(g), int(b)},
		HSL:       [3]float64{round4(hslH), round4(hslS), round4(hslL)},
		HSV:       [3]float64{round4(hsvH), round4(hsvS), round4(hsvV)},
		HWB:       [3]float64{round4(hwbH), round4(hwbW), round4(hwbB)},
		Lab:       [3]float64{round4(labL), round4(labA), round4(labB)},
		LCH:       [3]float64{round4(lchL), round4(lchC), round4(lchH)},
		OKLab:     [3]float64{round4(okL), round4(okA), round4(okB)},
		OKLCH:     [3]float64{round4(oklchL), round4(oklchC), round4(oklchH)},
		Luminance: round4(c.Luminance()),
	}
}

func round4(v float64) float64 { return math.Round(v*1e4) / 1e4 }
