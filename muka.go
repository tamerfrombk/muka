package main

import (
	"flag"
	"log"
	"muka/files"
	"os"
)

// Args holds the parsed program arguments
type Args struct {
	OriginalDirectory string
	DirectoryToSearch string
	IsInteractive     bool
	IsForce           bool
	IsDryRun          bool
}

func parseArgs() Args {
	directoryPtr := flag.String("d", ".", "the directory to search")
	interactivePtr := flag.Bool("i", false, "enable interactive mode to remove duplicates")
	forcePtr := flag.Bool("f", false, "remove duplicates without prompting")
	dryRunPtr := flag.Bool("dryrun", false, "does not actually remove any files")

	flag.Parse()

	var directoryToSearch string
	if *directoryPtr == "." {
		directoryToSearch, _ = os.Getwd()
	} else {
		directoryToSearch = *directoryPtr
	}

	return Args{
		OriginalDirectory: *directoryPtr,
		DirectoryToSearch: directoryToSearch,
		IsInteractive:     *interactivePtr,
		IsForce:           *forcePtr,
		IsDryRun:          *dryRunPtr,
	}
}

func setupLogger() {
	log.SetFlags(0)
}

func main() {

	setupLogger()

	args := parseArgs()

	fileHashes, err := files.CollectFiles(args.DirectoryToSearch)
	if err != nil {
		log.Printf("unable to find files in %q: %v", args.OriginalDirectory, err)
		os.Exit(1)
	}

	deleter := files.MakeDeleter(args.IsDryRun)
	if duplicates := files.FindDuplicateFiles(fileHashes); args.IsForce {
		files.ForceDelete(duplicates, deleter)
	} else if args.IsInteractive {
		for _, duplicate := range duplicates {
			files.PromptToDelete(os.Stdout, os.Stdin, deleter, duplicate)
		}
	} else {
		files.PrintDuplicates(duplicates)
	}
}
