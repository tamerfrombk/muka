package main

import (
	"flag"
	"fmt"
)

func main() {

	directoryPtr := flag.String("dir", ".", "the directory to search")

	flag.Parse()

	fmt.Println("Hello ", *directoryPtr)
}
