package color

import (
	"errors"
	"testing"
)

// TestParseRejectsNonFinite covers a hole Go opens by accepting "NaN" and
// "Inf" in ParseFloat. Those used to be taken as components and silently
// rounded into a plausible looking black.
func TestParseRejectsNonFinite(t *testing.T) {
	cases := []string{
		"rgb(NaN, 0, 0)",
		"rgb(0, Inf, 0)",
		"rgb(0, 0, -Inf)",
		"rgb(NaN%, 0%, 0%)",
		"rgba(0, 0, 0, NaN)",
		"hsl(NaN, 100%, 50%)",
		"hsl(0, Inf, 50%)",
		"hsl(Infdeg 100% 50%)",
		"hsv(0, NaN, 50%)",
		"hwb(NaN 0% 0%)",
		"lab(NaN 0 0)",
		"lch(50 NaN 0)",
		"oklab(0 NaN 0)",
		"oklch(0.5 0.1 NaN)",
		"oklch(Infinity 0.1 0)",
		"color(display-p3 NaN 0 0)",
		"color(display-p3 0 0 0 / NaN)",
		"color(rec2020 Inf 0 0)",
	}
	for _, in := range cases {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) should have rejected a non-finite value", in)
		}
	}
}

// TestParseRejectsOverflow covers values that are finite going in but overflow
// inside the conversion. The result was NaN, which rendered as #000000 with a
// zero exit code.
//
// Lab lightness is not in this list any more: CSS clamps it to 0..100 at
// parsed-value time, so lab(1e300 0 0) is white rather than an error. The
// opponent axes are still unbounded targets for an overflow.
func TestParseRejectsOverflow(t *testing.T) {
	cases := []string{
		"lab(50 1e300 0)",
		"lch(50 1e300 0)",
		"oklab(0 1e300 0)",
		"oklch(0.5 1e300 0)",
		"color(rec2020 1e300 0 0)",
		"color(prophoto-rgb 1e300 0 0)",
	}
	for _, in := range cases {
		_, err := Parse(in)
		if err == nil {
			t.Errorf("Parse(%q) should have been rejected as out of range", in)
			continue
		}
		if !errors.Is(err, ErrOutOfRange) {
			t.Errorf("Parse(%q) = %v, want ErrOutOfRange", in, err)
		}
	}
}

// TestParseClampsRatherThanOverflowing is the other half of the pair above: a
// huge but representable coordinate just lands outside the gamut and gets
// clipped, which is not an error.
func TestParseClampsRatherThanOverflowing(t *testing.T) {
	got, err := Parse("lab(50 0 1e300)")
	if err != nil {
		t.Fatalf("an out of gamut but representable color should clip: %v", err)
	}
	if got.Hex() == "" {
		t.Error("a clipped color should still render")
	}
}

// TestParseKeepsLegitimateOutOfRangeWideGamut guards the fix above against
// overreach: a negative display-p3 component names a real color and has to
// keep working.
func TestParseKeepsLegitimateOutOfRangeWideGamut(t *testing.T) {
	got, err := Parse("color(display-p3 -0.5 0 0)")
	if err != nil {
		t.Fatalf("a negative wide gamut component should parse: %v", err)
	}
	if got.Hex() != "#00180D" {
		t.Errorf("got %s, want #00180D", got.Hex())
	}
}

// TestHueBoundaryQuantization pins down the channels that land exactly on a
// .5 boundary. They used to quantize one step low, because 0.75 + 1/3 is not
// exactly representable in binary.
func TestHueBoundaryQuantization(t *testing.T) {
	cases := map[string]string{
		"hsl(0, 100%, 50%)":   "#FF0000",
		"hsl(60, 100%, 50%)":  "#FFFF00",
		"hsl(90, 100%, 50%)":  "#80FF00",
		"hsl(150, 100%, 50%)": "#00FF80",
		"hsl(210, 100%, 50%)": "#0080FF",
		"hsl(270, 100%, 50%)": "#8000FF",
		"hsl(330, 100%, 50%)": "#FF0080",
		"hsl(240, 100%, 50%)": "#0000FF",
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q): %v", in, err)
			continue
		}
		if got.Hex() != want {
			t.Errorf("Parse(%q) = %s, want %s", in, got.Hex(), want)
		}
	}
}

func TestNumberRejectsNonFinite(t *testing.T) {
	for _, in := range []string{"NaN", "nan", "Inf", "-Inf", "+Inf", "infinity"} {
		if _, err := number(in, 255); err == nil {
			t.Errorf("number(%q) should have been rejected", in)
		}
	}
	// The range error from ParseFloat already covers overflow.
	if _, err := number("1e400", 255); err == nil {
		t.Error("number(1e400) should have been rejected")
	}
}

func TestAngleRejectsNonFinite(t *testing.T) {
	for _, in := range []string{"NaN", "Inf", "NaNDeg", "Infturn"} {
		if _, err := angle(in); err == nil {
			t.Errorf("angle(%q) should have been rejected", in)
		}
	}
}
