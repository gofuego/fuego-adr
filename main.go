package main

import (
	"fmt"
	"os"

	"github.com/gofuego/fuego-adr/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "fuego-adr: %v\n", err)
		os.Exit(1)
	}
}
