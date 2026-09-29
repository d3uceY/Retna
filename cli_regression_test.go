package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/contrast"
	"github.com/d3uceY/Retna/output"
)

// TestRunContrastMinNeedsWCAG covers the bug where --min with a non-ratio
// algorithm found nothing to compare and reported a pass at 0.00:1 with a zero
// exit code, which would turn a CI check green on nothing.
func TestRunContrastMinNeedsWCAG(t *testing.T) {
	_, code, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{
			foreground: "black", background: "white",
			algorithm: "apca", min: 4.5, minSet: true,
		})
	})
	if err == nil {
		t.Fatalf("--min without wcag should fail, got exit code %d", code)
	}
	if !strings.Contains(err.Error(), "wcag") {
		t.Errorf("the error should name wcag, got %v", err)
	}
}

// TestRunContrastMinStillWorksWithAll keeps the guard from being too broad:
// --all includes wcag, so --min still has something to compare.
func TestRunContrastMinStillWorksWithAll(t *testing.T) {
	out, code, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{
			foreground: "black", background: "white",
			all: true, min: 4.5, minSet: true,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("21:1 meets 4.5:1, exit code = %d", code)
	}
	if !strings.Contains(out, "PASS") {
		t.Errorf("expected a pass\n%s", out)
	}
}

func TestRunContrastMinMustBePositiveFinite(t *testing.T) {
	for _, min := range []float64{0, -1, -0.5, math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, _, err := run(t, func(w *bytes.Buffer) (int, error) {
			return runContrast(w, contrastOptions{
				foreground: "black", background: "white", min: min, minSet: true,
			})
		})
		if err == nil {
			t.Errorf("--min %v should have been rejected", min)
		}
	}
}

func TestRunFixRejectsBadSuggestCount(t *testing.T) {
	for _, suggest := range []int{0, -1} {
		var buf bytes.Buffer
		err := runFix(&buf, "#777777", fixOptions{
			background: "#ffffff", level: "AA", suggest: suggest,
		})
		if err == nil {
			t.Errorf("--suggest %d should have been rejected", suggest)
		}
	}
}

func TestRunFixRejectsNonFiniteTarget(t *testing.T) {
	for _, target := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0, 0.5} {
		var buf bytes.Buffer
		err := runFix(&buf, "#777777", fixOptions{
			background: "#ffffff", target: target, targetSet: true, suggest: 3,
		})
		if err == nil {
			t.Errorf("--target %v should have been rejected", target)
		}
	}
}

// TestRunReadableCompositesBackground checks that the color shown is the one
// the ratios were measured against. The measurement composites alpha, so the
// report has to as well.
func TestRunReadableCompositesBackground(t *testing.T) {
	// The two inputs are not quite the same color: 0x80 is 128/255, a hair
	// under a half, so it composites one step darker than an exact 0.5.
	cases := map[string]string{
		"rgba(0, 0, 0, 0.5)": "#808080",
		"#00000080":          "#7F7F7F",
	}
	for in, want := range cases {
		out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
			return 0, runReadable(w, in, readableOptions{level: "AA"})
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Background "+want) {
			t.Errorf("input %s should report %s\n%s", in, want, out)
		}
		if strings.Contains(out, "#00000080") {
			t.Errorf("input %s should not show the raw translucent color\n%s", in, out)
		}
		if !strings.Contains(out, "composited over white") {
			t.Errorf("input %s should say the alpha was resolved\n%s", in, out)
		}
	}
}

func TestRunReadableJSONReportsCompositedBackground(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runReadable(w, "rgba(0, 0, 0, 0.5)", readableOptions{level: "AA", json: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"background": "#808080"`) {
		t.Errorf("JSON should report the composited background\n%s", out)
	}
}

func TestRunReadableOpaqueIsNotNoted(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runReadable(w, "#3498db", readableOptions{level: "AA"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "composited over white") {
		t.Errorf("nothing was translucent, so there should be no note\n%s", out)
	}
}

// TestRunFixRatioMatchesTheHexItPrints covers a mismatch where the ratio came
// from the unquantized OKLab color while the hex came from the quantized one.
// Measured apart, those straddle the target, so the command could suggest a hex
// that fails the very check it reported as passing.
func TestRunFixRatioMatchesTheHexItPrints(t *testing.T) {
	bg, err := color.Parse("#ffffff")
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"#777777", "#FFFFFF", "#3498db", "#00ff00", "#808080"} {
		out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
			return 0, runFix(w, input, fixOptions{
				background: "#ffffff", level: "AA", suggest: 5, json: true,
			})
		})
		if err != nil {
			t.Fatalf("input %s: %v", input, err)
		}

		var decoded output.FixJSON
		if err := json.Unmarshal([]byte(out), &decoded); err != nil {
			t.Fatalf("input %s: invalid JSON: %v", input, err)
		}
		if len(decoded.Suggestions) == 0 {
			t.Fatalf("input %s produced no suggestions", input)
		}

		for _, suggestion := range decoded.Suggestions {
			fg, err := color.Parse(suggestion.Hex)
			if err != nil {
				t.Fatal(err)
			}
			measured := contrast.WCAG{}.Calculate(fg, bg).Value
			if math.Abs(measured-suggestion.Contrast) > 1e-9 {
				t.Errorf("input %s: %s is quoted at %.6f but measures %.6f",
					input, suggestion.Hex, suggestion.Contrast, measured)
			}
			if measured < 4.5 {
				t.Errorf("input %s: %s is quoted as passing but measures %.6f",
					input, suggestion.Hex, measured)
			}
		}
	}
}

func TestRunFixRejectsUnreachableTarget(t *testing.T) {
	for _, target := range []float64{21.000001, 25, 100} {
		var buf bytes.Buffer
		err := runFix(&buf, "#777777", fixOptions{
			background: "#ffffff", target: target, targetSet: true, suggest: 3,
		})
		if err == nil {
			t.Errorf("--target %v is above the 21:1 maximum and should be rejected", target)
		}
	}
}

// TestReadColorFileStripsBOM covers a Windows papercut: editors write a UTF-8
// BOM without asking, and it used to end up glued to the first color.
func TestReadColorFileStripsBOM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "palette.txt")
	if err := os.WriteFile(path, []byte("\ufeff#fff\n#000\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readColorFile(newRootCmd(), path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"#fff", "#000"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestReadColorFileNamesTheFile(t *testing.T) {
	_, err := readColorFile(newRootCmd(), filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("a missing file should fail")
	}
	if !strings.Contains(err.Error(), "missing.txt") {
		t.Errorf("the error should name the file, got %v", err)
	}
}

func TestReadColorFileSkipsCommentsAndBlanks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "palette.txt")
	content := "#fff\n\n// a comment\n#000, #111\n   \n#222\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readColorFile(newRootCmd(), path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"#fff", "#000", "#111", "#222"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// TestRunConvertRejectsNonFiniteColor routes the numeric guard through a
// command, so the error surfaces the way a user would see it.
func TestRunConvertRejectsNonFiniteColor(t *testing.T) {
	var buf bytes.Buffer
	err := runConvert(&buf, convertOptions{input: "rgb(NaN, 0, 0)", to: []string{"hex"}})
	if err == nil {
		t.Fatal("a NaN component should be rejected")
	}
	if buf.Len() != 0 {
		t.Errorf("nothing should be printed on failure, got %q", buf.String())
	}
}
