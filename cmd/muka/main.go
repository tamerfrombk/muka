package main

import (
	"os"

	"github.com/tamerfrombk/muka/pkg/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
