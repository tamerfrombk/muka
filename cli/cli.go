package cli

import (
	"flag"
	"log"
	"os"

	"github.com/tamerfrombk/muka/muka"
)

type args struct {
	OriginalDirectory string
	DirectoryToSearch string
	IsInteractive     bool
	IsForce           bool
	IsDryRun          bool
}

func parseArgs(mainArgs []string) args {
	mukaFlags := flag.NewFlagSet("muka", flag.ExitOnError)

	directoryPtr := mukaFlags.String("d", ".", "the directory to search")
	interactivePtr := mukaFlags.Bool("i", false, "enable interactive mode to remove duplicates")
	forcePtr := mukaFlags.Bool("f", false, "remove duplicates without prompting")
	dryRunPtr := mukaFlags.Bool("dryrun", false, "do not actually remove any files")

	mukaFlags.Parse(mainArgs)

	var directoryToSearch string
	if *directoryPtr == "." {
		directoryToSearch, _ = os.Getwd()
	} else {
		directoryToSearch = *directoryPtr
	}

	return args{
		OriginalDirectory: *directoryPtr,
		DirectoryToSearch: directoryToSearch,
		IsInteractive:     *interactivePtr,
		IsForce:           *forcePtr,
		IsDryRun:          *dryRunPtr,
	}
}

func setupLogger() {
	// Prevent displaying any additional data to log messages
	log.SetFlags(0)
}

// Run main entry point
func Run(mainArgs []string) int {

	setupLogger()

	args := parseArgs(mainArgs)

	fileHashes, err := muka.CollectFiles(args.DirectoryToSearch)
	if err != nil {
		log.Printf("unable to find files in %q: %v", args.OriginalDirectory, err)
		return 1
	}

	deleter := muka.MakeDeleter(args.IsDryRun)
	if duplicates := muka.FindDuplicateFiles(fileHashes); args.IsForce {
		muka.ForceDelete(duplicates, deleter)
	} else if args.IsInteractive {
		for _, duplicate := range duplicates {
			muka.PromptToDelete(os.Stdout, os.Stdin, deleter, duplicate)
		}
	} else {
		muka.PrintDuplicates(duplicates)
	}

	return 0
}