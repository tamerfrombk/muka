package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tamerfrombk/muka/muka"
)

type args struct {
	OriginalDirectory  string
	IsInteractive      bool
	IsForce            bool
	IsDryRun           bool
	IsReport           bool
	FileCollectOptions muka.FileCollectionOptions
}

func parseArgs(mainArgs []string) (args, error) {
	mukaFlags := flag.NewFlagSet("muka", flag.ExitOnError)

	directoryPtr := mukaFlags.String("d", ".", "the directory to search")
	interactivePtr := mukaFlags.Bool("i", false, "enable interactive mode to remove duplicates")
	forcePtr := mukaFlags.Bool("f", false, "remove duplicates without prompting")
	dryRunPtr := mukaFlags.Bool("dryrun", false, "do not actually remove any files")
	reportPtr := mukaFlags.Bool("report", false, "generates a report displaying basic program performance")
	excludeDirsPtr := mukaFlags.String("X", "", "exclude the provided directories from consideration (regex supported)")
	excludeFilesPtr := mukaFlags.String("x", "", "exclude the provided files from consideration (regex supported)")

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

	excludeFiles, err := muka.CompileSpaceSeparatedPatterns(*excludeFilesPtr)
	if err != nil {
		return args{}, nil
	}

	return args{
		OriginalDirectory: *directoryPtr,
		IsInteractive:     *interactivePtr,
		IsForce:           *forcePtr,
		IsDryRun:          *dryRunPtr,
		IsReport:          *reportPtr,
		FileCollectOptions: muka.FileCollectionOptions{
			DirectoryToSearch: directoryToSearch,
			ExcludeDirs:       excludeDirs,
			ExcludeFiles:      excludeFiles,
		},
	}, nil
}

func setupLogger() {
	// Prevent displaying any additional data to log messages
	log.SetFlags(0)
}

func onInteractive(deleter muka.Deleter, duplicates []muka.DuplicateFile) []muka.FileHash {
	var deletedFiles []muka.FileHash
	for _, duplicate := range duplicates {
		files, err := muka.PromptToDelete(os.Stdout, os.Stdin, deleter, duplicate)
		if err == nil {
			for _, f := range files {
				deletedFiles = append(deletedFiles, f)
			}
		}
	}

	return deletedFiles
}

// Run main entry point
func Run(mainArgs []string) int {

	setupLogger()

	args, err := parseArgs(mainArgs)
	if err != nil {
		log.Printf("unable to parse arguments: %v", err)
		return 1
	}

	directory, err := muka.CollectFiles(args.FileCollectOptions)
	if err != nil {
		log.Printf("unable to find files in %q: %v", args.OriginalDirectory, err)
		return 1
	}

	deleter := muka.MakeDeleter(args.IsDryRun)
	duplicates := muka.FindDuplicateFiles(directory)

	var deletedFiles []muka.FileHash
	if args.IsForce {
		deletedFiles = muka.ForceDelete(duplicates, deleter)
	} else if args.IsInteractive {
		deletedFiles = onInteractive(deleter, duplicates)
	} else {
		muka.PrintDuplicates(duplicates)
	}

	if args.IsReport {
		report := muka.CalculateReport(directory, duplicates, deletedFiles)
		fmt.Println(report)
	}

	return 0
}
