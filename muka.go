package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Args holds the parsed program arguments
type Args struct {
	OriginalDirectory string
	DirectoryToSearch string
	IsInteractive     bool
	IsForce           bool
	IsDryRun          bool
}

// CollectFiles Recursively walks the provided directory and creates FileHash for each encountered file
func CollectFiles(dir string) ([]FileHash, error) {
	var files []FileHash
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		absolutePath, _ := filepath.Abs(file)
		cleanPath := filepath.ToSlash(absolutePath)
		hash, err := hashFile(cleanPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "'%s' could not be hashed because '%s'.\n", cleanPath, err.Error())
			return nil
		}

		files = append(files, hash)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func hashFile(filePath string) (FileHash, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileHash{}, err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return FileHash{}, err
	}

	fileHash := FileHash{
		AbsolutePath: filePath,
		Hash:         hex.EncodeToString(hasher.Sum(nil)),
	}

	return fileHash, nil
}

// FindDuplicateFiles does as it suggests
func FindDuplicateFiles(fileHashes []FileHash) []DuplicateFile {

	cache := NewCache()
	for _, fileHash := range fileHashes {
		cache.Add(fileHash)
	}

	return cache.GetDuplicates()
}

func promptToDelete(writer io.Writer, reader io.Reader, deleter Deleter, dup DuplicateFile) {
	if len(dup.Duplicates) == 0 {
		return
	}

	bufReader := bufio.NewReader(reader)
	for done := false; !done; {
		fmt.Fprintf(writer, "%s\n", dup)
		fmt.Fprintf(writer, "Which file(s) do you wish to remove? [o/d/s] > ")

		line, _ := bufReader.ReadString('\n')
		if line == "\n" {
			continue
		}

		switch answer := line[0]; answer {
		case 'd':
			for _, d := range dup.Duplicates {
				deleter.Delete(d.AbsolutePath)
			}
			done = true
			break
		case 'o':
			deleter.Delete(dup.Original.AbsolutePath)
			done = true
			break
		case 's':
			done = true
			break
		default:
			fmt.Fprintf(os.Stderr, "'%c' is not acceptable answer.\n", answer)
			break
		}
	}
}

func printDuplicates(duplicates []DuplicateFile) {
	for _, duplicate := range duplicates {
		if len(duplicate.Duplicates) > 0 {
			fmt.Println(duplicate)
		}
	}
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

func forceDelete(duplicates []DuplicateFile, deleter Deleter) {
	for _, dup := range duplicates {
		if len(dup.Duplicates) == 0 {
			continue
		}

		for _, f := range dup.Duplicates {
			deleter.Delete(f.AbsolutePath)
		}
	}
}

func main() {

	args := parseArgs()

	fileHashes, err := CollectFiles(args.DirectoryToSearch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find files in directory '%s' due to '%s'.\n", args.OriginalDirectory, err.Error())
		os.Exit(1)
	}

	deleter := MakeDeleter(args.IsDryRun)
	if duplicates := FindDuplicateFiles(fileHashes); args.IsForce {
		forceDelete(duplicates, deleter)
	} else if args.IsInteractive {
		for _, duplicate := range duplicates {
			promptToDelete(os.Stdout, os.Stdin, deleter, duplicate)
		}
	} else {
		printDuplicates(duplicates)
	}
}
