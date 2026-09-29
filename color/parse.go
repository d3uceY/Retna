package color

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ErrUnknownFormat is returned when a string is not a color Retna recognizes.
var ErrUnknownFormat = errors.New("unrecognized color format")

// Parse turns a CSS color string into an sRGB color. It accepts hex, the
// rgb, hsl, hsv, hwb, lab, lch, oklab and oklch functions, and the CSS named
// colors.
func Parse(s string) (Color, error) {
	in := strings.TrimSpace(s)
	if in == "" {
		return Color{}, errors.New("empty color")
	}

	lower := strings.ToLower(in)
	if lower == "transparent" {
		return Color{0, 0, 0, 0}, nil
	}
	if hex, ok := named[lower]; ok {
		return parseHex(hex)
	}
	if i := strings.IndexByte(in, '('); i >= 0 {
		return parseFunc(lower[:i], in[i:])
	}
	return parseHex(in)
}

func parseFunc(name, call string) (Color, error) {
	args, err := callArgs(call)
	if err != nil {
		return Color{}, err
	}
	switch strings.TrimSpace(name) {
	case "rgb", "rgba":
		return parseRGB(args)
	case "hsl", "hsla":
		return parseHSL(args)
	case "hsv", "hsb", "hsva", "hsba":
		return parseHSV(args)
	case "hwb", "hwba":
		return parseHWB(args)
	case "lab":
		return parseLab(args)
	case "lch":
		return parseLCH(args)
	case "oklab":
		return parseOKLab(args)
	case "oklch":
		return parseOKLCH(args)
	}
	return Color{}, fmt.Errorf("%w: %q", ErrUnknownFormat, name)
}

// callArgs splits "name(a b c / d)" into its channel strings. Commas and
// whitespace both separate channels, and a slash introduces alpha.
func callArgs(call string) ([]string, error) {
	if !strings.HasSuffix(call, ")") {
		return nil, fmt.Errorf("%w: missing closing parenthesis in %q", ErrUnknownFormat, call)
	}
	body := call[1 : len(call)-1]
	channels, alpha := body, ""
	if i := strings.IndexByte(body, '/'); i >= 0 {
		channels, alpha = body[:i], body[i+1:]
	}
	args := strings.Fields(strings.ReplaceAll(channels, ",", " "))
	if strings.TrimSpace(alpha) != "" {
		args = append(args, strings.TrimSpace(alpha))
	}
	return args, nil
}

func parseRGB(args []string) (Color, error) {
	if err := wantArgs("rgb", args, 3, 4); err != nil {
		return Color{}, err
	}
	var out [3]float64
	for i := range out {
		v, err := number(args[i], 255)
		if err != nil {
			return Color{}, fmt.Errorf("rgb channel %d: %w", i+1, err)
		}
		out[i] = clamp01(v / 255)
	}
	a, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return Color{out[0], out[1], out[2], a}, nil
}

func parseHSL(args []string) (Color, error) {
	if err := wantArgs("hsl", args, 3, 4); err != nil {
		return Color{}, err
	}
	h, err := angle(args[0])
	if err != nil {
		return Color{}, fmt.Errorf("hsl hue: %w", err)
	}
	s, err := number(args[1], 1)
	if err != nil {
		return Color{}, fmt.Errorf("hsl saturation: %w", err)
	}
	l, err := number(args[2], 1)
	if err != nil {
		return Color{}, fmt.Errorf("hsl lightness: %w", err)
	}
	a, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return HSL(h, clamp01(s), clamp01(l), a), nil
}

func parseHSV(args []string) (Color, error) {
	if err := wantArgs("hsv", args, 3, 4); err != nil {
		return Color{}, err
	}
	h, err := angle(args[0])
	if err != nil {
		return Color{}, fmt.Errorf("hsv hue: %w", err)
	}
	s, err := number(args[1], 1)
	if err != nil {
		return Color{}, fmt.Errorf("hsv saturation: %w", err)
	}
	v, err := number(args[2], 1)
	if err != nil {
		return Color{}, fmt.Errorf("hsv value: %w", err)
	}
	a, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return HSV(h, clamp01(s), clamp01(v), a), nil
}

func parseHWB(args []string) (Color, error) {
	if err := wantArgs("hwb", args, 3, 4); err != nil {
		return Color{}, err
	}
	h, err := angle(args[0])
	if err != nil {
		return Color{}, fmt.Errorf("hwb hue: %w", err)
	}
	w, err := number(args[1], 1)
	if err != nil {
		return Color{}, fmt.Errorf("hwb whiteness: %w", err)
	}
	bl, err := number(args[2], 1)
	if err != nil {
		return Color{}, fmt.Errorf("hwb blackness: %w", err)
	}
	a, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return HWB(h, clamp01(w), clamp01(bl), a), nil
}

func parseLab(args []string) (Color, error) {
	if err := wantArgs("lab", args, 3, 4); err != nil {
		return Color{}, err
	}
	l, err := number(args[0], 100)
	if err != nil {
		return Color{}, fmt.Errorf("lab lightness: %w", err)
	}
	a, err := number(args[1], 1)
	if err != nil {
		return Color{}, fmt.Errorf("lab a: %w", err)
	}
	b, err := number(args[2], 1)
	if err != nil {
		return Color{}, fmt.Errorf("lab b: %w", err)
	}
	alpha, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return Lab(l, a, b, alpha), nil
}

func parseLCH(args []string) (Color, error) {
	if err := wantArgs("lch", args, 3, 4); err != nil {
		return Color{}, err
	}
	l, err := number(args[0], 100)
	if err != nil {
		return Color{}, fmt.Errorf("lch lightness: %w", err)
	}
	chroma, err := number(args[1], 1.5)
	if err != nil {
		return Color{}, fmt.Errorf("lch chroma: %w", err)
	}
	h, err := angle(args[2])
	if err != nil {
		return Color{}, fmt.Errorf("lch hue: %w", err)
	}
	alpha, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return LCH(l, chroma, h, alpha), nil
}

func parseOKLab(args []string) (Color, error) {
	if err := wantArgs("oklab", args, 3, 4); err != nil {
		return Color{}, err
	}
	l, err := number(args[0], 1)
	if err != nil {
		return Color{}, fmt.Errorf("oklab lightness: %w", err)
	}
	a, err := number(args[1], 1)
	if err != nil {
		return Color{}, fmt.Errorf("oklab a: %w", err)
	}
	b, err := number(args[2], 1)
	if err != nil {
		return Color{}, fmt.Errorf("oklab b: %w", err)
	}
	alpha, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return OKLab(clamp01(l), a, b, alpha), nil
}

func parseOKLCH(args []string) (Color, error) {
	if err := wantArgs("oklch", args, 3, 4); err != nil {
		return Color{}, err
	}
	l, err := number(args[0], 1)
	if err != nil {
		return Color{}, fmt.Errorf("oklch lightness: %w", err)
	}
	chroma, err := number(args[1], 0.4)
	if err != nil {
		return Color{}, fmt.Errorf("oklch chroma: %w", err)
	}
	h, err := angle(args[2])
	if err != nil {
		return Color{}, fmt.Errorf("oklch hue: %w", err)
	}
	alpha, err := alphaArg(args, 3)
	if err != nil {
		return Color{}, err
	}
	return OKLCH(clamp01(l), chroma, h, alpha), nil
}

func parseHex(s string) (Color, error) {
	h := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(s)), "#")
	if !isHex(h) {
		return Color{}, fmt.Errorf("%w: %q", ErrUnknownFormat, s)
	}
	if len(h) == 3 || len(h) == 4 {
		var b strings.Builder
		for _, r := range h {
			b.WriteRune(r)
			b.WriteRune(r)
		}
		h = b.String()
	}
	if len(h) != 6 && len(h) != 8 {
		return Color{}, fmt.Errorf("%w: %q", ErrUnknownFormat, s)
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return Color{}, fmt.Errorf("%w: %q", ErrUnknownFormat, s)
	}
	if len(h) == 6 {
		return RGB8(uint8(n>>16), uint8(n>>8), uint8(n)), nil
	}
	return RGBA8(uint8(n>>24), uint8(n>>16), uint8(n>>8), float64(uint8(n))/255), nil
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// number parses a plain number or a percentage. scale is what 100% maps to,
// so rgb channels pass 255 while saturation passes 1.
func number(s string, scale float64) (float64, error) {
	s = strings.TrimSpace(s)
	raw := s
	if p, ok := strings.CutSuffix(s, "%"); ok {
		s = strings.TrimSpace(p)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", raw)
	}
	if raw != s {
		return v / 100 * scale, nil
	}
	return v, nil
}

func angle(s string) (float64, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(t, "grad"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(t, "grad"), 64)
		return wrapHue(v * 0.9), err
	case strings.HasSuffix(t, "deg"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(t, "deg"), 64)
		return wrapHue(v), err
	case strings.HasSuffix(t, "rad"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(t, "rad"), 64)
		return wrapHue(v * 180 / math.Pi), err
	case strings.HasSuffix(t, "turn"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(t, "turn"), 64)
		return wrapHue(v * 360), err
	}
	v, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid angle %q", s)
	}
	return wrapHue(v), nil
}

func alphaArg(args []string, idx int) (float64, error) {
	if len(args) <= idx {
		return 1, nil
	}
	a, err := number(args[idx], 1)
	if err != nil {
		return 0, fmt.Errorf("alpha: %w", err)
	}
	return clamp01(a), nil
}

func wantArgs(name string, args []string, min, max int) error {
	if len(args) < min || len(args) > max {
		return fmt.Errorf("%s takes %d or %d values, got %d", name, min, max, len(args))
	}
	return nil
}
