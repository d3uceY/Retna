package main

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/contrast"
	"github.com/d3uceY/Retna/output"
)

func parseColor(s string) (color.Color, error) { return color.Parse(s) }

func run(t *testing.T, fn func(*bytes.Buffer) (int, error)) (string, int, error) {
	t.Helper()
	var buf bytes.Buffer
	code, err := fn(&buf)
	return buf.String(), code, err
}

func TestRunContrastSingle(t *testing.T) {
	out, code, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "#777", background: "#fff"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	for _, want := range []string{"Contrast", "#777777", "#FFFFFF", "4.48:1", "AA Normal", "AA Large"} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q\n%s", want, out)
		}
	}
}

func TestRunContrastTable(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{
			foreground: "#777",
			against:    []string{"#fff,#000", "#f5f5f5"},
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"#FFFFFF", "#000000", "#F5F5F5", "Ratio", "AA", "AAA"} {
		if !strings.Contains(out, want) {
			t.Errorf("table is missing %q\n%s", want, out)
		}
	}
}

func TestRunContrastMinExitCode(t *testing.T) {
	cases := []struct {
		min  float64
		want int
	}{
		{4.5, 1},
		{4.4, 0},
		{4.47, 0},
		{4.479, 1},
	}
	for _, tc := range cases {
		out, code, err := run(t, func(w *bytes.Buffer) (int, error) {
			return runContrast(w, contrastOptions{
				foreground: "#777", background: "#fff", min: tc.min, minSet: true,
			})
		})
		if err != nil {
			t.Fatalf("min %v: %v", tc.min, err)
		}
		if code != tc.want {
			t.Errorf("min %v gave exit code %d, want %d", tc.min, code, tc.want)
		}
		if tc.want == 1 && !strings.Contains(out, "FAIL") {
			t.Errorf("min %v should print FAIL\n%s", tc.min, out)
		}
		if tc.want == 0 && !strings.Contains(out, "PASS") {
			t.Errorf("min %v should print PASS\n%s", tc.min, out)
		}
	}
}

func TestRunContrastJSON(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "#777", background: "#fff", json: true})
	})
	if err != nil {
		t.Fatal(err)
	}

	var decoded output.ContrastJSON
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if decoded.Foreground != "#777777" || decoded.Background != "#FFFFFF" {
		t.Errorf("colors = %s on %s", decoded.Foreground, decoded.Background)
	}
	if math.Abs(decoded.Contrast-4.478089453577214) > 1e-12 {
		t.Errorf("contrast = %v", decoded.Contrast)
	}
	if decoded.WCAG == nil {
		t.Fatal("wcag block is missing")
	}
	if decoded.WCAG.AA.Normal || !decoded.WCAG.AA.Large {
		t.Errorf("aa = %+v", decoded.WCAG.AA)
	}
	if decoded.WCAG.AAA.Normal || decoded.WCAG.AAA.Large {
		t.Errorf("aaa = %+v", decoded.WCAG.AAA)
	}
}

func TestRunContrastAgainstJSON(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "#777", against: []string{"#fff", "#000"}, json: true})
	})
	if err != nil {
		t.Fatal(err)
	}

	var decoded output.ContrastListJSON
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if decoded.Foreground != "#777777" {
		t.Errorf("foreground = %q", decoded.Foreground)
	}
	if len(decoded.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(decoded.Results))
	}
	if decoded.Results[0].Background != "#FFFFFF" || decoded.Results[1].Background != "#000000" {
		t.Errorf("backgrounds = %q, %q", decoded.Results[0].Background, decoded.Results[1].Background)
	}
}

func TestRunContrastAPCA(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "black", background: "white", algorithm: "apca"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "APCA") || !strings.Contains(out, "106.04") {
		t.Errorf("APCA output is wrong\n%s", out)
	}
	if strings.Contains(out, "WCAG") {
		t.Errorf("only APCA was requested\n%s", out)
	}
}

func TestRunContrastAll(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "black", background: "white", all: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "APCA") || !strings.Contains(out, "WCAG") {
		t.Errorf("--all should run both algorithms\n%s", out)
	}
}

func TestRunContrastAlphaIsComposited(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "rgba(0, 0, 0, 0.5)", background: "#ffffff"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#808080") {
		t.Errorf("half transparent black should composite to #808080\n%s", out)
	}
	if !strings.Contains(out, "composited over white") {
		t.Errorf("the alpha decision should be reported\n%s", out)
	}
}

func TestRunContrastErrors(t *testing.T) {
	cases := []contrastOptions{
		{foreground: "#777"},
		{foreground: "nope", background: "#fff"},
		{foreground: "#777", background: "nope"},
		{foreground: "#777", background: "#fff", algorithm: "nope"},
	}
	for _, opts := range cases {
		if _, _, err := run(t, func(w *bytes.Buffer) (int, error) { return runContrast(w, opts) }); err == nil {
			t.Errorf("%+v should have failed", opts)
		}
	}
}

func TestSplitList(t *testing.T) {
	got := splitList([]string{"a,b", "c", " d , e "})
	want := []string{"a", "b", "c", "d", "e"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestSelectAlgorithms(t *testing.T) {
	one, err := selectAlgorithms("wcag", false)
	if err != nil || len(one) != 1 || one[0].Name() != "wcag" {
		t.Fatalf("got %v, %v", one, err)
	}

	every, err := selectAlgorithms("wcag", true)
	if err != nil || len(every) != len(contrast.Names()) {
		t.Fatalf("--all returned %d algorithms, want %d", len(every), len(contrast.Names()))
	}

	if _, err := selectAlgorithms("nope", false); err == nil {
		t.Error("an unknown algorithm should fail")
	}
}

func TestRunConvert(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runConvert(w, convertOptions{input: "#1a73e8", to: []string{"hsl,oklch"}})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "HSL") || !strings.Contains(out, "OKLCH") {
		t.Errorf("missing requested spaces\n%s", out)
	}
	if strings.Contains(out, "HWB") {
		t.Errorf("unrequested space was printed\n%s", out)
	}
}

func TestRunConvertDefaultsToAllSpaces(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runConvert(w, convertOptions{input: "rebeccapurple", to: []string{"all"}})
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, space := range []string{"HEX", "RGB", "RGBA", "HSL", "HSV", "HWB", "LAB", "LCH", "OKLAB", "OKLCH", "SRGB", "DISPLAY-P3", "A98-RGB", "PROPHOTO-RGB", "REC2020"} {
		if !strings.Contains(out, space) {
			t.Errorf("missing %s\n%s", space, out)
		}
	}
}

func TestRunConvertWideGamut(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runConvert(w, convertOptions{input: "red", to: []string{"display-p3", "rec2020"}})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "color(display-p3 0.9175 0.2003 0.1386)") {
		t.Errorf("display-p3 value is wrong\n%s", out)
	}
	if !strings.Contains(out, "color(rec2020 ") {
		t.Errorf("rec2020 value is missing\n%s", out)
	}
}

// TestRunContrastAcceptsColorFunction checks the whole pipeline: a color()
// input is parsed, converted to sRGB and then measured.
func TestRunContrastAcceptsColorFunction(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return runContrast(w, contrastOptions{foreground: "color(display-p3 1 0 0)", background: "white"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#FF0000") {
		t.Errorf("the P3 red primary should land on sRGB red\n%s", out)
	}
	if !strings.Contains(out, "4.00:1") {
		t.Errorf("expected a 4.00:1 ratio\n%s", out)
	}
}

func TestRunConvertJSON(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runConvert(w, convertOptions{input: "#1a73e8", to: []string{"hex", "hsl"}, json: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded output.ConvertJSON
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Input != "#1A73E8" {
		t.Errorf("input = %q", decoded.Input)
	}
	if decoded.Values["hex"] != "#1A73E8" {
		t.Errorf("hex = %q", decoded.Values["hex"])
	}
	if len(decoded.Values) != 2 {
		t.Errorf("got %d values, want 2", len(decoded.Values))
	}
}

func TestRunConvertErrors(t *testing.T) {
	var buf bytes.Buffer
	if err := runConvert(&buf, convertOptions{input: "nope", to: []string{"all"}}); err == nil {
		t.Error("a bad color should fail")
	}
	if err := runConvert(&buf, convertOptions{input: "#fff", to: []string{"cmyk"}}); err == nil {
		t.Error("a bad space should fail")
	}
}

func TestResolveSpaces(t *testing.T) {
	got, err := resolveSpaces([]string{"hsl,hsl", "all"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(color.Spaces) {
		t.Fatalf("got %v, want every space", got)
	}
	// An explicitly requested space keeps its place at the front.
	if got[0] != "hsl" {
		t.Errorf("requested order was not preserved: %v", got)
	}

	empty, err := resolveSpaces(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != len(color.Spaces) {
		t.Errorf("no --to should print every space, got %v", empty)
	}
}

func TestRunInspect(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runInspect(w, "#3498db", false)
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"#3498DB", "rgb(52, 152, 219)", "hsl(204.07, 69.87%, 53.14%)", "Relative luminance", "0.2830"} {
		if !strings.Contains(out, want) {
			t.Errorf("inspect is missing %q\n%s", want, out)
		}
	}
}

func TestRunInspectJSON(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runInspect(w, "#3498db", true)
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded output.ColorJSON
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Hex != "#3498DB" || decoded.RGB != [3]int{52, 152, 219} {
		t.Errorf("decoded = %+v", decoded)
	}
	if decoded.Luminance != 0.283 {
		t.Errorf("luminance = %v", decoded.Luminance)
	}
}

func TestRunInspectBadColor(t *testing.T) {
	var buf bytes.Buffer
	if err := runInspect(&buf, "nope", false); err == nil {
		t.Error("a bad color should fail")
	}
}

func TestRunReadable(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runReadable(w, "#3498db", readableOptions{level: "AA"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#000000") || !strings.Contains(out, "6.66:1") {
		t.Errorf("missing the black candidate\n%s", out)
	}
	if !strings.Contains(out, "#FFFFFF") || !strings.Contains(out, "3.15:1") {
		t.Errorf("missing the white candidate\n%s", out)
	}
	// Black outranks white here, so it has to come first.
	if strings.Index(out, "#000000") > strings.Index(out, "#FFFFFF") {
		t.Errorf("candidates are not sorted by contrast\n%s", out)
	}
}

func TestRunReadableBest(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runReadable(w, "#3498db", readableOptions{level: "AA", best: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#000000") {
		t.Errorf("expected black as the best choice\n%s", out)
	}
	if strings.Contains(out, "#FFFFFF") {
		t.Errorf("--best should print one candidate\n%s", out)
	}
}

func TestRunReadableJSON(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runReadable(w, "#3498db", readableOptions{level: "AAA", json: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded output.ReadableJSON
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Level != "AAA" || decoded.Minimum != 7 {
		t.Errorf("decoded = %+v", decoded)
	}
	if len(decoded.Candidates) != 2 {
		t.Fatalf("got %d candidates", len(decoded.Candidates))
	}
	for _, candidate := range decoded.Candidates {
		if candidate.Color == "#000000" && candidate.Pass {
			t.Error("black does not reach 7:1 on #3498DB, it should not pass")
		}
	}
}

func TestRunReadableBadLevel(t *testing.T) {
	var buf bytes.Buffer
	if err := runReadable(&buf, "#fff", readableOptions{level: "AAA+"}); err == nil {
		t.Error("an unknown level should fail")
	}
}

func TestLevelMin(t *testing.T) {
	cases := map[string]float64{"AA": 4.5, "aa": 4.5, " AAA ": 7}
	for level, want := range cases {
		got, err := levelMin(level)
		if err != nil {
			t.Fatalf("levelMin(%q): %v", level, err)
		}
		if got != want {
			t.Errorf("levelMin(%q) = %v, want %v", level, got, want)
		}
	}
	if _, err := levelMin("A"); err == nil {
		t.Error("levelMin should reject an unknown level")
	}
}

func TestRunFix(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runFix(w, "#777777", fixOptions{background: "#ffffff", level: "AA", suggest: 3})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#777777") || !strings.Contains(out, "4.48:1") {
		t.Errorf("fix should report the original\n%s", out)
	}
	if !strings.Contains(out, "Suggested colors") {
		t.Errorf("fix should suggest colors\n%s", out)
	}
}

func TestRunFixAlreadyPasses(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runFix(w, "#000000", fixOptions{background: "#ffffff", level: "AA", suggest: 3})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Already passes") {
		t.Errorf("expected an already passing message\n%s", out)
	}
	if strings.Contains(out, "Suggested colors") {
		t.Errorf("nothing should be suggested\n%s", out)
	}
}

func TestRunFixJSON(t *testing.T) {
	out, _, err := run(t, func(w *bytes.Buffer) (int, error) {
		return 0, runFix(w, "#777777", fixOptions{background: "#ffffff", level: "AA", suggest: 3, json: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded output.FixJSON
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Target != 4.5 || decoded.Pass {
		t.Errorf("decoded = %+v", decoded)
	}
	if len(decoded.Suggestions) != 3 {
		t.Fatalf("got %d suggestions, want 3", len(decoded.Suggestions))
	}
	for _, suggestion := range decoded.Suggestions {
		if !suggestion.Pass || suggestion.Contrast < 4.5 {
			t.Errorf("suggestion %+v does not meet the target", suggestion)
		}
	}
	if decoded.Suggestions[0].Hex != "#767676" {
		t.Errorf("the smallest fix should be #767676, got %s", decoded.Suggestions[0].Hex)
	}
}

func TestRunFixErrors(t *testing.T) {
	cases := []fixOptions{
		{background: "nope"},
		{background: "#fff", level: "A"},
		{background: "#fff", target: 0.5, targetSet: true},
	}
	for _, opts := range cases {
		var buf bytes.Buffer
		if err := runFix(&buf, "#777777", opts); err == nil {
			t.Errorf("%+v should have failed", opts)
		}
	}
}

// TestPassingColorsKeepsHueAndChroma checks the fix stays recognisable: only
// lightness moves, and the very first suggestion is the smallest change.
func TestPassingColorsKeepsHueAndChroma(t *testing.T) {
	start, err := parseColor("#1A73E8")
	if err != nil {
		t.Fatal(err)
	}
	bg, err := parseColor("#FFFFFF")
	if err != nil {
		t.Fatal(err)
	}
	got := passingColors(start, bg, 4.5, 3)
	if len(got) != 3 {
		t.Fatalf("got %d suggestions, want 3", len(got))
	}
	if got[0].ratio > 4.6 {
		t.Errorf("the first suggestion should be close to the target, got %s at %.2f:1", got[0].hex, got[0].ratio)
	}
	for i := 1; i < len(got); i++ {
		if got[i].ratio <= got[i-1].ratio {
			t.Errorf("suggestions should get more contrasty: %v", got)
		}
	}

	if got := passingColors(start, bg, 4.5, 0); got != nil {
		t.Errorf("a count of zero should suggest nothing, got %v", got)
	}
}

func TestRootCommandSurface(t *testing.T) {
	root := newRootCmd()
	names := map[string]bool{}
	for _, cmd := range root.Commands() {
		names[cmd.Name()] = true
	}
	for _, want := range []string{"contrast", "convert", "inspect", "palette", "readable", "fix", "version"} {
		if !names[want] {
			t.Errorf("root is missing the %q command", want)
		}
	}

	var palette *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "palette" {
			palette = cmd
		}
	}
	if palette == nil || len(palette.Commands()) != 1 || palette.Commands()[0].Name() != "contrast" {
		t.Error("palette should have a contrast subcommand")
	}
}

func TestVersionCommand(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), version) {
		t.Errorf("version output = %q, want it to contain %q", buf.String(), version)
	}
}
