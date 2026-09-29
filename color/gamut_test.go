package color

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
)

// The D65 and D50 white points as XYZ, which every matrix has to reproduce.
var (
	d65White = [3]float64{0.9504559270516716, 1, 1.0890577507598784}
	d50White = [3]float64{0.9642956764295677, 1, 0.8251046025104601}
)

// TestGamutMatricesAreInverses catches a mistyped entry in either direction of
// any matrix pair.
func TestGamutMatricesAreInverses(t *testing.T) {
	probes := [][3]float64{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1},
		{0.3, 0.6, 0.9}, {0.11, 0.97, 0.42},
	}
	for name, space := range gamuts {
		for _, p := range probes {
			x, y, z := mat3(space.toXYZ, p[0], p[1], p[2])
			r, g, b := mat3(space.fromXYZ, x, y, z)
			if math.Abs(r-p[0]) > 1e-9 || math.Abs(g-p[1]) > 1e-9 || math.Abs(b-p[2]) > 1e-9 {
				t.Errorf("%s: %v survived a round trip through XYZ as %v", name, p, [3]float64{r, g, b})
			}
		}
	}
}

// TestGamutMatricesReachTheirWhitePoint checks that each row of toXYZ sums to
// the space's white point, which is the property that makes white map to white.
func TestGamutMatricesReachTheirWhitePoint(t *testing.T) {
	for name, space := range gamuts {
		want := d65White
		if space.d50 {
			want = d50White
		}
		x, y, z := mat3(space.toXYZ, 1, 1, 1)
		if math.Abs(x-want[0]) > 1e-9 || math.Abs(y-want[1]) > 1e-9 || math.Abs(z-want[2]) > 1e-9 {
			t.Errorf("%s white is %v, want %v", name, [3]float64{x, y, z}, want)
		}
	}
}

func TestColorSpaceWhiteAndBlack(t *testing.T) {
	for _, name := range wideGamutNames {
		white, err := Parse("color(" + name + " 1 1 1)")
		if err != nil {
			t.Fatalf("color(%s 1 1 1): %v", name, err)
		}
		if white.Hex() != "#FFFFFF" {
			t.Errorf("color(%s 1 1 1) = %s, want #FFFFFF", name, white.Hex())
		}

		black, err := Parse("color(" + name + " 0 0 0)")
		if err != nil {
			t.Fatalf("color(%s 0 0 0): %v", name, err)
		}
		if black.Hex() != "#000000" {
			t.Errorf("color(%s 0 0 0) = %s, want #000000", name, black.Hex())
		}
	}
}

// TestColorSpaceSRGBMatchesRGB pins down that color(srgb ...) is the same
// gamma-encoded sRGB triple that rgb() already accepts.
func TestColorSpaceSRGBMatchesRGB(t *testing.T) {
	pairs := [][2]string{
		{"color(srgb 0.5 0.5 0.5)", "rgb(128, 128, 128)"},
		{"color(srgb 1 0 0.4)", "rgb(255, 0, 102)"},
	}
	for _, pair := range pairs {
		a, err := Parse(pair[0])
		if err != nil {
			t.Fatalf("Parse(%q): %v", pair[0], err)
		}
		b, err := Parse(pair[1])
		if err != nil {
			t.Fatalf("Parse(%q): %v", pair[1], err)
		}
		if a.Hex() != b.Hex() {
			t.Errorf("%s = %s but %s = %s", pair[0], a.Hex(), pair[1], b.Hex())
		}
	}
}

// TestDisplayP3GrayMatchesSRGB works because display-p3 shares both the sRGB
// transfer curve and the D65 white point, so an even gray is the same in both.
func TestDisplayP3GrayMatchesSRGB(t *testing.T) {
	for _, step := range []float64{0, 0.25, 0.5, 1} {
		in := "color(display-p3 " + num(step, 4) + " " + num(step, 4) + " " + num(step, 4) + ")"
		got, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		want := byteChannel(step)
		r, g, b := got.Channels()
		if r != want || g != want || b != want {
			t.Errorf("%s = %s, want rgb(%d, %d, %d)", in, got.Hex(), want, want, want)
		}
	}
}

// TestDisplayP3ReferenceValues uses the published conversion of the sRGB
// primaries into display-p3, as computed by colorjs.io.
func TestDisplayP3ReferenceValues(t *testing.T) {
	cases := []struct {
		in   string
		want [3]float64
	}{
		{"red", [3]float64{0.917509, 0.200286, 0.138561}},
		{"white", [3]float64{1, 1, 1}},
		{"black", [3]float64{0, 0, 0}},
	}
	for _, tc := range cases {
		c, err := Parse(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		r, g, b, err := wideComponents(c, "display-p3")
		if err != nil {
			t.Fatal(err)
		}
		got := [3]float64{r, g, b}
		for i := range got {
			// 1e-4 is loose enough for the rounding in the published figures
			// and far tighter than the error a wrong matrix would produce.
			if math.Abs(got[i]-tc.want[i]) > 1e-4 {
				t.Errorf("display-p3 of %s = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}
}

// TestWideGamutHoldsAllOfSRGB is the invariant that catches a bad matrix:
// display-p3, a98-rgb, prophoto-rgb and rec2020 all contain sRGB, so no sRGB
// color can land outside 0..1 in any of them.
func TestWideGamutHoldsAllOfSRGB(t *testing.T) {
	inputs := []string{
		"#000000", "#FFFFFF", "#FF0000", "#00FF00", "#0000FF",
		"#FFFF00", "#00FFFF", "#FF00FF", "#777777", "#3498DB", "#1A73E8",
	}
	for _, in := range inputs {
		c, err := Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"display-p3", "a98-rgb", "prophoto-rgb", "rec2020"} {
			r, g, b, err := wideComponents(c, name)
			if err != nil {
				t.Fatal(err)
			}
			for i, v := range [3]float64{r, g, b} {
				// The matrices are inverses to about 1e-8, so allow that much
				// slop rather than demanding an exact zero.
				if v < -1e-6 || v > 1+1e-6 {
					t.Errorf("%s component %d in %s is %v, outside the gamut", in, i, name, v)
				}
			}
		}
	}
}

func TestColorSpaceAlpha(t *testing.T) {
	got, err := Parse("color(display-p3 1 0 0 / 0.5)")
	if err != nil {
		t.Fatal(err)
	}
	if got.A != 0.5 {
		t.Errorf("alpha = %v, want 0.5", got.A)
	}
	if got.Hex() != "#FF000080" {
		t.Errorf("hex = %s, want #FF000080", got.Hex())
	}

	withComma, err := Parse("color(display-p3 1, 0, 0, 0.25)")
	if err != nil {
		t.Fatal(err)
	}
	if withComma.A != 0.25 {
		t.Errorf("alpha = %v, want 0.25", withComma.A)
	}
}

func TestColorSpacePercentages(t *testing.T) {
	percent, err := Parse("color(rec2020 100% 0% 0%)")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Parse("color(rec2020 1 0 0)")
	if err != nil {
		t.Fatal(err)
	}
	if percent.Hex() != plain.Hex() {
		t.Errorf("percentages gave %s, plain values gave %s", percent.Hex(), plain.Hex())
	}
}

func TestColorSpaceIsCaseInsensitive(t *testing.T) {
	lower, err := Parse("color(display-p3 1 0 0)")
	if err != nil {
		t.Fatal(err)
	}
	upper, err := Parse("color(DisPlay-P3 1 0 0)")
	if err != nil {
		t.Fatal(err)
	}
	if lower.Hex() != upper.Hex() {
		t.Errorf("%s and %s differ", lower.Hex(), upper.Hex())
	}
}

// TestColorSpaceOutOfRangeValuesAreAccepted covers the CSS rule that color()
// components may sit outside 0..1. Display has to clip, but the parse must
// succeed rather than reject the color.
func TestColorSpaceOutOfRangeValuesAreAccepted(t *testing.T) {
	cases := map[string]string{
		"color(display-p3 1.2 0 0)": "#FF0000",
		// A negative component is not garbage, it names a real color. This one
		// is #00180D in sRGB, which is what makes sign preserving transfer
		// functions worth the trouble.
		"color(display-p3 -0.5 0 0)": "#00180D",
		"color(rec2020 1 0 0)":       "#FF0000",
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", in, err)
			continue
		}
		if got.Hex() != want {
			t.Errorf("Parse(%q) = %s, want %s", in, got.Hex(), want)
		}
	}
}

func TestColorSpaceRejects(t *testing.T) {
	cases := []string{
		"color()",
		"color(display-p3)",
		"color(display-p3 1)",
		"color(display-p3 1 0)",
		"color(display-p3 1 0 0 0 0)",
		"color(nope 1 0 0)",
		"color(xyz-d65 0.2 0.3 0.4)",
		"color(display-p3 one 0 0)",
	}
	for _, in := range cases {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) should have failed", in)
		}
	}

	// An unknown space is a format problem, not a syntax problem.
	_, err := Parse("color(nope 1 0 0)")
	if !errors.Is(err, ErrUnknownFormat) {
		t.Errorf("got %v, want ErrUnknownFormat", err)
	}
	if err != nil && !strings.Contains(err.Error(), "display-p3") {
		t.Errorf("the error should list the known spaces, got %v", err)
	}
}

func TestColorSpaceRoundTrip(t *testing.T) {
	inputs := []string{
		"#000000", "#FFFFFF", "#FF0000", "#00FF00", "#0000FF",
		"#777777", "#3498DB", "#1A73E8", "#F5F5F5", "#663399",
	}
	for _, in := range inputs {
		want, err := Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range wideGamutNames {
			text, err := Format(want, name)
			if err != nil {
				t.Fatalf("Format(%s, %s): %v", in, name, err)
			}
			got, err := Parse(text)
			if err != nil {
				t.Errorf("Parse(%q) from %s of %s: %v", text, name, in, err)
				continue
			}
			gr, gg, gb := got.Channels()
			wr, wg, wb := want.Channels()
			if diff(gr, wr) > 1 || diff(gg, wg) > 1 || diff(gb, wb) > 1 {
				t.Errorf("%s round trip of %s: %q gave %s, want %s", name, in, text, got.Hex(), want.Hex())
			}
		}
	}
}

func TestFormatWideGamutShape(t *testing.T) {
	c, err := Parse("red")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Format(c, "display-p3")
	if err != nil {
		t.Fatal(err)
	}
	if got != "color(display-p3 0.9175 0.2003 0.1386)" {
		t.Errorf("Format = %q", got)
	}

	if _, err := Format(c, "xyz"); err == nil {
		t.Error("Format should reject an unknown wide gamut name")
	}
	if _, err := WideGamut(c, "xyz"); err == nil {
		t.Error("WideGamut should reject an unknown name")
	}
}

// TestSpacesListsEveryGamut keeps the display list and the color() table from
// drifting apart.
func TestSpacesListsEveryGamut(t *testing.T) {
	for _, name := range wideGamutNames {
		if !slices.Contains(Spaces, name) {
			t.Errorf("Spaces is missing %q", name)
		}
	}
	if len(gamuts) != len(wideGamutNames) {
		t.Errorf("gamuts has %d entries but wideGamutNames has %d", len(gamuts), len(wideGamutNames))
	}
}

// TestRec2020UsesGamma24 pins the rec2020 transfer function to the plain gamma
// 2.4 curve CSS specifies (the BT.1886 reference EOTF), not the scene referred
// BT.2020 OETF with its 1.0993 / 0.01805 constants and 4.5 slope. The two
// disagree by a factor of two near black, and the difference is silent.
func TestRec2020UsesGamma24(t *testing.T) {
	for _, v := range []float64{-1, -0.5, -0.01805, 0, 0.01805, 0.1, 0.5, 1} {
		want := math.Copysign(math.Pow(math.Abs(v), 2.4), v)
		if got := rec2020ToLinear(v); got != want {
			t.Errorf("rec2020ToLinear(%v) = %v, want %v", v, got, want)
		}
		wantBack := math.Copysign(math.Pow(math.Abs(v), 1/2.4), v)
		if got := rec2020FromLinear(v); got != wantBack {
			t.Errorf("rec2020FromLinear(%v) = %v, want %v", v, got, wantBack)
		}
	}

	// The old OETF would decode these two values to 0.004011 and 0.100568.
	if got := rec2020ToLinear(0.01805); math.Abs(got-0.004011) < 0.001 {
		t.Errorf("rec2020ToLinear(0.01805) = %v, which looks like the old BT.2020 OETF", got)
	}
}

// TestBradfordMatricesAreExactInverses pins the D65 to D50 pair to the one CSS
// publishes. That pair was recomputed from the cone response matrix rather than
// from rounded intermediates, so it inverts to machine precision and it carries
// the D65 white point onto the chromaticity derived D50 exactly. The earlier
// pair was off by about 1e-7, which is small but shows up as drift.
func TestBradfordMatricesAreExactInverses(t *testing.T) {
	probes := [][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, {0.3, 0.6, 0.9}}
	for _, p := range probes {
		x, y, z := d65ToD50(p[0], p[1], p[2])
		r, g, b := d50ToD65(x, y, z)
		for i, got := range [3]float64{r, g, b} {
			if math.Abs(got-p[i]) > 1e-15 {
				t.Errorf("D65 %v round tripped through D50 as %v", p, [3]float64{r, g, b})
				break
			}
		}
	}

	x, y, z := d65ToD50(d65White[0], d65White[1], d65White[2])
	for i, got := range [3]float64{x, y, z} {
		if math.Abs(got-d50White[i]) > 1e-15 {
			t.Errorf("D65 white adapted to D50 is %v, want %v", [3]float64{x, y, z}, d50White)
			break
		}
	}
}
