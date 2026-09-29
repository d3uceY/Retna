package color

import (
	"math"
	"testing"
)

// eps is the slack allowed for a conversion that goes out to a space and back.
// Lab crosses two white points through Bradford matrices, so it is not
// bit-exact; 1e-4 is still about a fortieth of an 8-bit step (1/255).
const eps = 1e-4

func TestLuminance(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"#000000", 0},
		{"#FFFFFF", 1},
		{"#3498DB", 0.28301},
	}
	for _, tc := range cases {
		c, err := Parse(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if got := c.Luminance(); math.Abs(got-tc.want) > 1e-5 {
			t.Errorf("Luminance(%s) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestReferenceConversions(t *testing.T) {
	cases := []struct {
		hex  string
		lab  [3]float64
		oklm [3]float64
	}{
		{"#000000", [3]float64{0, 0, 0}, [3]float64{0, 0, 0}},
		{"#FFFFFF", [3]float64{100, 0, 0}, [3]float64{1, 0, 0}},
		{"#FF0000", [3]float64{54.2905, 80.8049, 69.8910}, [3]float64{0.6279554, 0.2248632, 0.1258463}},
		{"#00FF00", [3]float64{87.8185, -79.2711, 80.9946}, [3]float64{0.8664396, -0.2338874, 0.1794985}},
		{"#0000FF", [3]float64{29.5683, 68.2874, -112.0297}, [3]float64{0.4520137, -0.0324570, -0.3115281}},
	}
	for _, tc := range cases {
		c, err := Parse(tc.hex)
		if err != nil {
			t.Fatal(err)
		}

		l, a, b := ToLab(c)
		if math.Abs(l-tc.lab[0]) > 1e-3 || math.Abs(a-tc.lab[1]) > 1e-3 || math.Abs(b-tc.lab[2]) > 1e-3 {
			t.Errorf("ToLab(%s) = %v, want %v", tc.hex, [3]float64{l, a, b}, tc.lab)
		}

		ol, oa, ob := ToOKLab(c)
		if math.Abs(ol-tc.oklm[0]) > 1e-5 || math.Abs(oa-tc.oklm[1]) > 1e-5 || math.Abs(ob-tc.oklm[2]) > 1e-5 {
			t.Errorf("ToOKLab(%s) = %v, want %v", tc.hex, [3]float64{ol, oa, ob}, tc.oklm)
		}
	}
}

func TestReferenceHSLAndHSV(t *testing.T) {
	cases := []struct {
		hex string
		hsl [3]float64
		hsv [3]float64
		hwb [3]float64
	}{
		{"#FF0000", [3]float64{0, 1, 0.5}, [3]float64{0, 1, 1}, [3]float64{0, 0, 0}},
		{"#00FF00", [3]float64{120, 1, 0.5}, [3]float64{120, 1, 1}, [3]float64{120, 0, 0}},
		{"#0000FF", [3]float64{240, 1, 0.5}, [3]float64{240, 1, 1}, [3]float64{240, 0, 0}},
		{"#FFFFFF", [3]float64{0, 0, 1}, [3]float64{0, 0, 1}, [3]float64{0, 1, 0}},
		{"#000000", [3]float64{0, 0, 0}, [3]float64{0, 0, 0}, [3]float64{0, 0, 1}},
	}
	for _, tc := range cases {
		c, err := Parse(tc.hex)
		if err != nil {
			t.Fatal(err)
		}

		h, s, l := ToHSL(c)
		checkTriple(t, "ToHSL", tc.hex, [3]float64{h, s, l}, tc.hsl)

		vh, vs, vv := ToHSV(c)
		checkTriple(t, "ToHSV", tc.hex, [3]float64{vh, vs, vv}, tc.hsv)

		wh, ww, wb := ToHWB(c)
		checkTriple(t, "ToHWB", tc.hex, [3]float64{wh, ww, wb}, tc.hwb)
	}
}

func checkTriple(t *testing.T, name, hex string, got, want [3]float64) {
	t.Helper()
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-6 {
			t.Errorf("%s(%s) = %v, want %v", name, hex, got, want)
			return
		}
	}
}

// TestSpaceRoundTrip feeds each in-gamut color through a space and back. The
// math has to be reversible, so the tolerance here is tight.
func TestSpaceRoundTrip(t *testing.T) {
	inputs := []string{
		"#000000", "#FFFFFF", "#FF0000", "#00FF00", "#0000FF",
		"#777777", "#1A73E8", "#3498DB", "#663399", "#808080", "#F5F5F5",
	}
	for _, in := range inputs {
		c, err := Parse(in)
		if err != nil {
			t.Fatal(err)
		}

		h, s, l := ToHSL(c)
		near(t, in, "hsl", c, HSL(h, s, l, 1))

		vh, vs, vv := ToHSV(c)
		near(t, in, "hsv", c, HSV(vh, vs, vv, 1))

		wh, ww, wb := ToHWB(c)
		near(t, in, "hwb", c, HWB(wh, ww, wb, 1))

		ll, la, lb := ToLab(c)
		near(t, in, "lab", c, Lab(ll, la, lb, 1))

		lh, lc, lhue := ToLCH(c)
		near(t, in, "lch", c, LCH(lh, lc, lhue, 1))

		ol, oa, ob := ToOKLab(c)
		near(t, in, "oklab", c, OKLab(ol, oa, ob, 1))

		kl, kc, kh := ToOKLCH(c)
		near(t, in, "oklch", c, OKLCH(kl, kc, kh, 1))
	}
}

func near(t *testing.T, in, space string, want, got Color) {
	t.Helper()
	if math.Abs(want.R-got.R) > eps || math.Abs(want.G-got.G) > eps || math.Abs(want.B-got.B) > eps {
		t.Errorf("%s round trip of %s: got %s, want %s", space, in, got.Hex(), want.Hex())
	}
}

// TestFormatReparse checks the text form of each space survives a reparse. The
// printed form rounds to a few decimals, so a byte of slack is expected.
func TestFormatReparse(t *testing.T) {
	inputs := []string{
		"#000000", "#FFFFFF", "#FF0000", "#00FF00", "#0000FF",
		"#777777", "#1A73E8", "#3498DB", "#663399", "#F5F5F5",
	}
	for _, in := range inputs {
		want, err := Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		for _, space := range Spaces {
			text, err := Format(want, space)
			if err != nil {
				t.Fatalf("Format(%s, %s): %v", in, space, err)
			}
			got, err := Parse(text)
			if err != nil {
				t.Errorf("Parse(%q) from %s of %s: %v", text, space, in, err)
				continue
			}
			gr, gg, gb := got.Channels()
			wr, wg, wb := want.Channels()
			if diff(gr, wr) > 1 || diff(gg, wg) > 1 || diff(gb, wb) > 1 {
				t.Errorf("%s round trip of %s: %q gave %s, want %s", space, in, text, got.Hex(), want.Hex())
			}
		}
	}
}

func diff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func TestFormatNames(t *testing.T) {
	c, err := Parse("#1A73E8")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"hex":   "#1A73E8",
		"rgb":   "rgb(26, 115, 232)",
		"rgba":  "rgba(26, 115, 232, 1)",
		"hsl":   "hsl(214.08, 81.75%, 50.59%)",
		"hsv":   "hsv(214.08, 88.79%, 90.98%)",
		"hwb":   "hwb(214.08 10.2% 9.02%)",
		"lab":   "lab(48.77, 9.69, -67.47)",
		"oklab": "oklab(0.5737, -0.0409, -0.1902)",
	}
	for space, want := range cases {
		got, err := Format(c, space)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("Format(%s) = %q, want %q", space, got, want)
		}
	}
	if _, err := Format(c, "cmyk"); err == nil {
		t.Error("Format should reject an unknown space")
	}
}

func TestOver(t *testing.T) {
	white := RGB(1, 1, 1)
	half := Color{0, 0, 0, 0.5}
	if got := half.Over(white).Hex(); got != "#808080" {
		t.Errorf("half black over white = %s, want #808080", got)
	}
}

func TestClampKeepsChannels(t *testing.T) {
	out := Color{2, -1, 0.5, 1}.Clamp()
	if out.R != 1 || out.G != 0 || out.B != 0.5 {
		t.Fatalf("Clamp = %v", out)
	}
}

// TestAchromaticHueIsZero keeps a gray from printing an arbitrary angle. The
// conversion leaves a small residue on a neutral color (about 8e-6 in Lab and
// 4e-8 in OKLab) and the residue points in a different direction in each space,
// so both report 0.
func TestAchromaticHueIsZero(t *testing.T) {
	for _, in := range []string{"#000000", "#808080", "#FFFFFF", "#777777", "#F5F5F5"} {
		c, err := Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		if _, chroma, hue := ToLCH(c); hue != 0 {
			t.Errorf("ToLCH(%s) hue = %v at chroma %v, want 0", in, hue, chroma)
		}
		if _, chroma, hue := ToOKLCH(c); hue != 0 {
			t.Errorf("ToOKLCH(%s) hue = %v at chroma %v, want 0", in, hue, chroma)
		}
	}
}

// TestChromaticHueSurvives guards the other direction, so the threshold above
// cannot swallow a real hue.
func TestChromaticHueSurvives(t *testing.T) {
	c, err := Parse("red")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, hue := ToLCH(c); math.Abs(hue-40.86) > 0.05 {
		t.Errorf("ToLCH hue = %v, want about 40.86", hue)
	}
	if _, _, hue := ToOKLCH(c); math.Abs(hue-29.23) > 0.05 {
		t.Errorf("ToOKLCH hue = %v, want about 29.23", hue)
	}
}
