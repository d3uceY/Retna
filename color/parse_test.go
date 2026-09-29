package color

import (
	"errors"
	"testing"
)

func TestParseHex(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"#fff", "#FFFFFF"},
		{"#FFF", "#FFFFFF"},
		{"fff", "#FFFFFF"},
		{"#ffffff", "#FFFFFF"},
		{"#777", "#777777"},
		{"#0f0", "#00FF00"},
		{"#1a73e8", "#1A73E8"},
		{"  #1A73E8  ", "#1A73E8"},
		{"#ffffffff", "#FFFFFF"},
		{"#00000000", "#00000000"},
		{"#ff000080", "#FF000080"},
		{"#1a73e8cc", "#1A73E8CC"},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc.in, err)
			continue
		}
		if got.Hex() != tc.want {
			t.Errorf("Parse(%q) = %s, want %s", tc.in, got.Hex(), tc.want)
		}
	}
}

func TestParseNamed(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"red", "#FF0000"},
		{"RED", "#FF0000"},
		{" white ", "#FFFFFF"},
		{"rebeccapurple", "#663399"},
		{"CadetBlue", "#5F9EA0"},
		{"transparent", "#00000000"},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc.in, err)
			continue
		}
		if got.Hex() != tc.want {
			t.Errorf("Parse(%q) = %s, want %s", tc.in, got.Hex(), tc.want)
		}
	}
}

func TestParseFunctions(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"rgb(26, 115, 232)", "#1A73E8"},
		{"rgb(26 115 232)", "#1A73E8"},
		{"rgb(26,115,232)", "#1A73E8"},
		{"rgba(26, 115, 232, 0.5)", "#1A73E880"},
		{"rgba(26 115 232 / 0.5)", "#1A73E880"},
		{"rgb(100%, 100%, 100%)", "#FFFFFF"},
		{"rgb(0%, 0%, 0%)", "#000000"},
		{"hsl(0, 100%, 50%)", "#FF0000"},
		{"hsl(120, 100%, 50%)", "#00FF00"},
		{"hsl(240 100% 50%)", "#0000FF"},
		{"hsl(0.5turn 100% 50%)", "#00FFFF"},
		{"hsl(180deg 100% 50%)", "#00FFFF"},
		{"hsl(200grad 100% 50%)", "#00FFFF"},
		{"hsl(0, 0%, 0%)", "#000000"},
		{"hsl(0, 0%, 100%)", "#FFFFFF"},
		{"hsl(214.08, 81.75%, 50.59%)", "#1A73E8"},
		{"hsv(0, 100%, 100%)", "#FF0000"},
		{"hsb(120, 100%, 100%)", "#00FF00"},
		{"hwb(0 0% 0%)", "#FF0000"},
		{"hwb(0 100% 0%)", "#FFFFFF"},
		{"hwb(0 0% 100%)", "#000000"},
		{"lab(0 0 0)", "#000000"},
		{"lab(100 0 0)", "#FFFFFF"},
		{"oklab(0 0 0)", "#000000"},
		{"oklab(1 0 0)", "#FFFFFF"},
		{"oklch(0 0 0)", "#000000"},
		{"oklch(1 0 0)", "#FFFFFF"},
		{"oklch(0.62796 0.25768 29.23)", "#FF0000"},
		{"oklch(0.86644 0.29483 142.5)", "#00FF00"},
		{"oklch(0.45201 0.31321 264.05)", "#0000FF"},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc.in, err)
			continue
		}
		if got.Hex() != tc.want {
			t.Errorf("Parse(%q) = %s, want %s", tc.in, got.Hex(), tc.want)
		}
	}
}

func TestParseRejects(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"notacolor",
		"#12345",
		"#gggggg",
		"rgb(1, 2)",
		"rgb(1, 2, 3, 4, 5)",
		"hsl(0, 100%)",
		"rgb(1, x, 3)",
		"hsl(0, 100%, 50%",
		"hwb(0)",
		"oklch(0.5 0.1)",
	}
	for _, in := range cases {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) should have failed", in)
		}
	}
}

func TestParseUnknownFunctionIsErrUnknownFormat(t *testing.T) {
	_, err := Parse("device-cmyk(0 0 0 1)")
	if !errors.Is(err, ErrUnknownFormat) {
		t.Fatalf("got %v, want ErrUnknownFormat", err)
	}
}

// TestPercentageReferenceRanges pins the CSS percent reference ranges for the
// axes that are not 1: 100% is 125 in Lab, 150 chroma in LCH, 0.4 in Oklab and
// 0.4 chroma in OkLCh. A percentage and the equivalent number must therefore
// parse to the identical color.
func TestPercentageReferenceRanges(t *testing.T) {
	cases := [][2]string{
		{"lab(50 100% 0)", "lab(50 125 0)"},
		{"lab(50 -100% 0)", "lab(50 -125 0)"},
		{"lab(50 0 50%)", "lab(50 0 62.5)"},
		{"lab(100% 0 0)", "lab(100 0 0)"},
		{"lch(50 100% 40)", "lch(50 150 40)"},
		{"lch(50% 50% 40)", "lch(50 75 40)"},
		{"oklab(0.5 100% 0)", "oklab(0.5 0.4 0)"},
		{"oklab(0.5 0 -100%)", "oklab(0.5 0 -0.4)"},
		{"oklch(50% 100% 40)", "oklch(0.5 0.4 40)"},
	}
	for _, tc := range cases {
		percent, err := Parse(tc[0])
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc[0], err)
			continue
		}
		plain, err := Parse(tc[1])
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc[1], err)
			continue
		}
		if percent != plain {
			t.Errorf("%s and %s should be the same color, got %v and %v", tc[0], tc[1], percent, plain)
		}
	}
}

// TestParseTimeClamping covers the clamping CSS specifies at parsed-value time:
// Lab lightness to 0..100, negative chroma to 0, and negative HSL saturation to
// 0. Everything else stays unbounded so an out of gamut color can still be
// named and round tripped.
func TestParseTimeClamping(t *testing.T) {
	clamped := [][2]string{
		{"lab(150 0 0)", "lab(100 0 0)"},
		{"lab(-20 0 0)", "lab(0 0 0)"},
		{"lch(150 30 40)", "lch(100 30 40)"},
		{"lch(50 -30 40)", "lch(50 0 40)"},
		{"oklab(1.5 0 0)", "oklab(1 0 0)"},
		{"oklab(-0.5 0 0)", "oklab(0 0 0)"},
		{"oklch(0.5 -0.2 40)", "oklch(0.5 0 40)"},
		{"hsl(0 -50% 50%)", "hsl(0 0% 50%)"},
		// Lightness clamps before it can overflow, so this is white, not an error.
		{"lab(1e300 0 0)", "lab(100 0 0)"},
	}
	for _, tc := range clamped {
		got, err := Parse(tc[0])
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc[0], err)
			continue
		}
		want, err := Parse(tc[1])
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc[1], err)
			continue
		}
		if got != want {
			t.Errorf("%s should clamp to %s, got %v and %v", tc[0], tc[1], got, want)
		}
	}

	// The far ends of every axis stay unbounded, so an out of gamut color keeps
	// its real components instead of being squeezed onto the gamut boundary.
	// Comparing against the constructors keeps clipping out of the picture.
	unclamped := []struct {
		in   string
		want Color
	}{
		{"lab(50 200 0)", Lab(50, 200, 0, 1)},
		{"lch(50 300 40)", LCH(50, 300, 40, 1)},
		{"oklab(0.5 0.9 0)", OKLab(0.5, 0.9, 0, 1)},
		{"oklch(0.5 0.8 40)", OKLCH(0.5, 0.8, 40, 1)},
		{"hsl(0 150% 50%)", HSL(0, 1.5, 0.5, 1)},
		{"hsl(0 100% -20%)", HSL(0, 1, -0.2, 1)},
		{"hsl(0 100% 120%)", HSL(0, 1, 1.2, 1)},
	}
	for _, tc := range unclamped {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Parse(%q) = %v, want the unclamped %v", tc.in, got, tc.want)
		}
	}
}

func TestParseShortHexShorthand(t *testing.T) {
	got, err := Parse("#abc")
	if err != nil {
		t.Fatal(err)
	}
	if got.Hex() != "#AABBCC" {
		t.Fatalf("got %s, want #AABBCC", got.Hex())
	}

	withAlpha, err := Parse("#abcd")
	if err != nil {
		t.Fatal(err)
	}
	if withAlpha.Hex() != "#AABBCCDD" {
		t.Fatalf("got %s, want #AABBCCDD", withAlpha.Hex())
	}
}
