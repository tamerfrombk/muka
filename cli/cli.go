package cli

import (
	"flag"
	"log"
	"os"

	"github.com/tamerfrombk/muka/muka"
)

type args struct {
	OriginalDirectory  string
	IsInteractive      bool
	IsForce            bool
	IsDryRun           bool
	FileCollectOptions muka.FileCollectionOptions
}

func parseArgs(mainArgs []string) (args, error) {
	mukaFlags := flag.NewFlagSet("muka", flag.ExitOnError)

	directoryPtr := mukaFlags.String("d", ".", "the directory to search")
	interactivePtr := mukaFlags.Bool("i", false, "enable interactive mode to remove duplicates")
	forcePtr := mukaFlags.Bool("f", false, "remove duplicates without prompting")
	dryRunPtr := mukaFlags.Bool("dryrun", false, "do not actually remove any files")
	excludeDirsPtr := mukaFlags.String("X", "", "exclude the provided directories from searching (regex supported)")

	mukaFlags.Parse(mainArgs)

	var directoryToSearch string
	var err error
	if *directoryPtr == "." {
		directoryToSearch, err = os.Getwd()
		if err != nil {
			return args{}, err
		}
	} else {
		directoryToSearch = *directoryPtr
	}

	excludeDirs, err := muka.CompileSpaceSeparatedPatterns(*excludeDirsPtr)
	if err != nil {
		return args{}, nil
	}

	return args{
		OriginalDirectory: *directoryPtr,
		IsInteractive:     *interactivePtr,
		IsForce:           *forcePtr,
		IsDryRun:          *dryRunPtr,
		FileCollectOptions: muka.FileCollectionOptions{
			DirectoryToSearch: directoryToSearch,
			ExcludeDirs:       excludeDirs,
		},
	}, nil
}

func setupLogger() {
	// Prevent displaying any additional data to log messages
	log.SetFlags(0)
}

// Run main entry point
func Run(mainArgs []string) int {

	setupLogger()

	args, err := parseArgs(mainArgs)
	if err != nil {
		log.Printf("unable to parse arguments: %v", err)
		return 1
	}

	fileHashes, err := muka.CollectFiles(args.FileCollectOptions)
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
