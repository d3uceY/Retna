package contrast

import (
	"math"
	"strings"
	"testing"

	"github.com/d3uceY/Retna/color"
)

func mustParse(t *testing.T, s string) color.Color {
	t.Helper()
	c, err := color.Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return c
}

func TestWCAGRatio(t *testing.T) {
	cases := []struct {
		fg, bg string
		want   float64
	}{
		{"#000000", "#FFFFFF", 21},
		{"#FFFFFF", "#000000", 21},
		{"#FFFFFF", "#FFFFFF", 1},
		{"#000000", "#000000", 1},
		{"#777777", "#FFFFFF", 4.478089453577214},
		{"#767676", "#FFFFFF", 4.5426},
		{"#FF0000", "#FFFFFF", 3.9984198000000005},
	}
	for _, tc := range cases {
		got := WCAG{}.Calculate(mustParse(t, tc.fg), mustParse(t, tc.bg))
		if math.Abs(got.Value-tc.want) > 1e-3 {
			t.Errorf("WCAG %s on %s = %.6f, want %.6f", tc.fg, tc.bg, got.Value, tc.want)
		}
		if got.Algorithm != "wcag" {
			t.Errorf("Algorithm = %q, want wcag", got.Algorithm)
		}
	}
}

// TestWCAGSymmetry pins down that swapping the two colors cannot change the
// ratio, which is what the WCAG formula guarantees.
func TestWCAGSymmetry(t *testing.T) {
	pairs := [][2]string{
		{"#777777", "#FFFFFF"},
		{"#1A73E8", "#F5F5F5"},
		{"#000000", "#3498DB"},
	}
	for _, pair := range pairs {
		a := WCAG{}.Calculate(mustParse(t, pair[0]), mustParse(t, pair[1]))
		b := WCAG{}.Calculate(mustParse(t, pair[1]), mustParse(t, pair[0]))
		if math.Abs(a.Value-b.Value) > 1e-12 {
			t.Errorf("%s and %s gave %.12f and %.12f", pair[0], pair[1], a.Value, b.Value)
		}
	}
}

func TestWCAGChecks(t *testing.T) {
	checks := map[string]bool{
		"AA Normal":  false,
		"AA Large":   true,
		"AAA Normal": false,
		"AAA Large":  false,
	}
	result := WCAG{}.Calculate(mustParse(t, "#777777"), mustParse(t, "#FFFFFF"))
	if len(result.Checks) != len(checks) {
		t.Fatalf("got %d checks, want %d", len(result.Checks), len(checks))
	}
	for _, check := range result.Checks {
		want, ok := checks[check.Name]
		if !ok {
			t.Fatalf("unexpected check %q", check.Name)
		}
		if check.Pass != want {
			t.Errorf("check %q = %v, want %v", check.Name, check.Pass, want)
		}
	}
}

func TestWCAGChecksAllPassOnBlackAndWhite(t *testing.T) {
	result := WCAG{}.Calculate(mustParse(t, "#000000"), mustParse(t, "#FFFFFF"))
	for _, check := range result.Checks {
		if !check.Pass {
			t.Errorf("check %q failed at 21:1", check.Name)
		}
	}
}

// TestAPCAReferenceValues uses the published APCA figures for black on white
// and white on black.
func TestAPCAReferenceValues(t *testing.T) {
	blackOnWhite := APCA{}.Calculate(mustParse(t, "#000000"), mustParse(t, "#FFFFFF"))
	if math.Abs(blackOnWhite.Value-106.04) > 0.05 {
		t.Errorf("APCA black on white = %.4f, want about 106.04", blackOnWhite.Value)
	}

	whiteOnBlack := APCA{}.Calculate(mustParse(t, "#FFFFFF"), mustParse(t, "#000000"))
	if math.Abs(whiteOnBlack.Value+107.89) > 0.05 {
		t.Errorf("APCA white on black = %.4f, want about -107.89", whiteOnBlack.Value)
	}
}

// TestAPCAPolarity covers the part of APCA that surprises people: the sign of
// Lc says which color is the text, and equal colors score nothing.
func TestAPCAPolarity(t *testing.T) {
	equal := APCA{}.Calculate(mustParse(t, "#FFFFFF"), mustParse(t, "#FFFFFF"))
	if equal.Value != 0 {
		t.Errorf("equal colors gave %.4f, want 0", equal.Value)
	}

	darkOnLight := APCA{}.Calculate(mustParse(t, "#111111"), mustParse(t, "#EEEEEE"))
	if darkOnLight.Value <= 0 {
		t.Errorf("dark text on a light background should be positive, got %.4f", darkOnLight.Value)
	}

	lightOnDark := APCA{}.Calculate(mustParse(t, "#EEEEEE"), mustParse(t, "#111111"))
	if lightOnDark.Value >= 0 {
		t.Errorf("light text on a dark background should be negative, got %.4f", lightOnDark.Value)
	}
}

func TestAPCAChecksUseMagnitude(t *testing.T) {
	// Light text on a dark background passes on the absolute Lc even though
	// the raw value is negative.
	result := APCA{}.Calculate(mustParse(t, "#FFFFFF"), mustParse(t, "#000000"))
	for _, check := range result.Checks {
		if !check.Pass {
			t.Errorf("check %q failed for white on black", check.Name)
		}
	}
}

// TestAPCAPublishedTestVectors uses the vectors shipped with the apca-w3
// reference implementation. The first color is the text and the second is the
// background.
func TestAPCAPublishedTestVectors(t *testing.T) {
	cases := []struct {
		fg, bg string
		want   float64
	}{
		{"#888888", "#FFFFFF", 63.056},
		{"#FFFFFF", "#888888", -68.541},
		{"#112233", "#DDEEFF", 91.668},
	}
	for _, tc := range cases {
		got := APCA{}.Calculate(mustParse(t, tc.fg), mustParse(t, tc.bg))
		if math.Abs(got.Value-tc.want) > 0.01 {
			t.Errorf("APCA %s on %s = %.4f, want about %.3f", tc.fg, tc.bg, got.Value, tc.want)
		}
	}
}

// TestAPCABands checks the use case levels against the published Lc for
// #888888 on white. At 63.056 it clears the 60 band and everything below it,
// and misses the 75 and 90 bands.
func TestAPCABands(t *testing.T) {
	want := map[string]bool{
		"Lc 90 Body Preferred": false,
		"Lc 75 Body Minimum":   false,
		"Lc 60 Content Text":   true,
		"Lc 45 Large Text":     true,
		"Lc 30 Text Floor":     true,
		"Lc 15 Non-text":       true,
	}
	result := APCA{}.Calculate(mustParse(t, "#888888"), mustParse(t, "#FFFFFF"))
	if len(result.Checks) != len(want) {
		t.Fatalf("got %d checks, want %d", len(result.Checks), len(want))
	}
	for _, check := range result.Checks {
		expected, ok := want[check.Name]
		if !ok {
			t.Fatalf("unexpected check %q", check.Name)
		}
		if check.Pass != expected {
			t.Errorf("check %q = %v, want %v", check.Name, check.Pass, expected)
		}
	}
}

func TestRegistry(t *testing.T) {
	names := Names()
	if len(names) != 2 || names[0] != "apca" || names[1] != "wcag" {
		t.Fatalf("Names() = %v, want [apca wcag]", names)
	}

	for _, name := range names {
		algorithm, err := Get(name)
		if err != nil {
			t.Fatalf("Get(%q): %v", name, err)
		}
		if algorithm.Name() != name {
			t.Errorf("Get(%q).Name() = %q", name, algorithm.Name())
		}
	}

	if _, err := Get("nope"); err == nil {
		t.Error("Get should reject an unknown algorithm")
	} else if !strings.Contains(err.Error(), "apca") {
		t.Errorf("error should list the known algorithms, got %v", err)
	}

	if _, err := Get("WCAG"); err != nil {
		t.Errorf("Get should be case insensitive, got %v", err)
	}
}
