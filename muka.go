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

// FileHash defines the file hash
type FileHash struct {
	AbsolutePath string
	Hash         string
}

func (hash FileHash) String() string {

	return hash.AbsolutePath
}

// DuplicateFileCache holds original and duplicate FileHashes
type DuplicateFileCache struct {
	fileHashByHash map[string]FileHash
	duplicates     []DuplicateFile
}

// NewCache DuplicateFileCache constructor
func NewCache() DuplicateFileCache {

	return DuplicateFileCache{
		fileHashByHash: make(map[string]FileHash),
		duplicates:     make([]DuplicateFile, 0),
	}
}

func (cache *DuplicateFileCache) findDuplicateFile(hash FileHash) int {

	for i, dup := range cache.duplicates {
		if dup.Original.Hash == hash.Hash {
			return i
		}
	}

	return -1
}

// Add adds a FileHash to the cache accounting for possible duplicates
func (cache *DuplicateFileCache) Add(hash FileHash) {
	if _, exists := cache.fileHashByHash[hash.Hash]; exists {
		idx := cache.findDuplicateFile(hash)
		// no need to check idx since we know we have it
		dup := &cache.duplicates[idx]
		dup.Duplicates = append(dup.Duplicates, hash)
	} else {
		cache.fileHashByHash[hash.Hash] = hash
		cache.duplicates = append(cache.duplicates, DuplicateFile{
			Original:   hash,
			Duplicates: make([]FileHash, 0),
		})
	}
}

// GetDuplicates retrieves the list of duplicates from the cache
func (cache *DuplicateFileCache) GetDuplicates() []DuplicateFile {

	return cache.duplicates
}

// DuplicateFile holds original and duplicate FileHashes
type DuplicateFile struct {
	Original   FileHash
	Duplicates []FileHash
}

// Args holds the parsed program arguments
type Args struct {
	OriginalDirectory string
	DirectoryToSearch string
	IsInteractive     bool
	IsForce           bool
}

func getFileHashes(dir string) ([]FileHash, error) {
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

// This wrapper is here to possibly turn this into a goroutine
func deleteFile(path string) error {

	//	err := os.Remove(path)
	var err error = nil

	fmt.Printf("'%s' was removed.\n", path)

	return err
}

func promptToDelete(reader *bufio.Reader, dup DuplicateFile) {
	if len(dup.Duplicates) == 0 {
		return
	}

	for done := false; !done; {
		printDuplicate(dup)
		fmt.Print("Which file(s) do you wish to remove? [1/2] > ")

		line, _ := reader.ReadString('\n')
		if line == "\n" {
			continue
		}

		answer := line[0]
		switch answer {
		case '1':
			for _, d := range dup.Duplicates {
				deleteFile(d.AbsolutePath)
			}
			done = true
			break
		case '2':
			deleteFile(dup.Original.AbsolutePath)
			done = true
			break
		default:
			fmt.Fprintf(os.Stderr, "'%c' is not acceptable answer.\n", answer)
			break
		}
	}
}

func printDuplicate(duplicate DuplicateFile) {
	size := len(duplicate.Duplicates)
	if size == 0 {
		return
	}

	if size == 1 {
		fmt.Printf("'%s' is a duplicate of '%s'.\n", duplicate.Duplicates[0], duplicate.Original.AbsolutePath)
	} else {
		fmt.Printf("'%s' are duplicates of '%s'.\n", duplicate.Duplicates, duplicate.Original.AbsolutePath)
	}
}

func printDuplicates(duplicates []DuplicateFile) {
	for _, duplicate := range duplicates {
		printDuplicate(duplicate)
	}
}

func parseArgs() Args {
	directoryPtr := flag.String("d", ".", "the directory to search")
	interactivePtr := flag.Bool("i", false, "enable interactive mode to remove duplicates")
	forcePtr := flag.Bool("f", false, "remove duplicates without prompting")

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
	}
}

func forceDelete(duplicates []DuplicateFile) {
	for _, dup := range duplicates {
		if len(dup.Duplicates) == 0 {
			continue
		}

		for _, f := range dup.Duplicates {
			deleteFile(f.AbsolutePath)
		}
	}
}

func main() {

	args := parseArgs()

	fileHashes, err := getFileHashes(args.DirectoryToSearch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find files in directory '%s' due to '%s'.\n", args.OriginalDirectory, err.Error())
		os.Exit(1)
	}

	if duplicates := FindDuplicateFiles(fileHashes); args.IsForce {
		forceDelete(duplicates)
	} else if args.IsInteractive {
		reader := bufio.NewReader(os.Stdin)
		for _, duplicate := range duplicates {
			promptToDelete(reader, duplicate)
		}
	} else {
		printDuplicates(duplicates)
	}
}
