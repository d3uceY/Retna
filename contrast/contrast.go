// Package contrast holds the contrast algorithms Retna can run.
//
// Algorithms are looked up by name, so adding one is a file and a register
// call. The color package knows nothing about any of them.
package contrast

import (
	"fmt"
	"sort"
	"strings"

	"github.com/d3uceY/Retna/color"
)

// Check is a single rule an algorithm applies to a pair of colors.
type Check struct {
	Name string
	Min  float64
	Pass bool
}

// Result is what an algorithm reports for a pair of colors. Value is the
// algorithm's headline number: the contrast ratio for WCAG, Lc for APCA.
type Result struct {
	Algorithm string
	Value     float64
	Checks    []Check
}

// Algorithm measures the contrast between a foreground and a background.
type Algorithm interface {
	Name() string
	Calculate(foreground, background color.Color) Result
}

var registry = map[string]Algorithm{}

func register(a Algorithm) { registry[a.Name()] = a }

// Get returns the algorithm registered under name.
func Get(name string) (Algorithm, error) {
	a, ok := registry[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return nil, fmt.Errorf("unknown algorithm %q (want one of %s)", name, strings.Join(Names(), ", "))
	}
	return a, nil
}

// Names lists the registered algorithm names in alphabetical order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
