package main

import (
	"os"

	"github.com/tamerfrombk/muka/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
