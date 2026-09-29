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
