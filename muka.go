package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func processFiles(directory string) error {

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absolutePath, err2 := filepath.Abs(path)
		if err2 != nil {
			return err2
		}

		fmt.Println(absolutePath)

		return nil
	}

	return filepath.Walk(directory, walkFunc)
}

func main() {

	directoryPtr := flag.String("dir", ".", "the directory to search")

	flag.Parse()

	fmt.Printf("Processing directory '%s'.\n", *directoryPtr)

	processFiles(*directoryPtr)

	fmt.Println("Done.")
}
