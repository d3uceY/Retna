// Command retna converts colors between color spaces and measures how
// readable one color is on another.
package main

import (
	"os"

	"github.com/d3uceY/Retna/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
