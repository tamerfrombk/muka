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

// DuplicateFile holds original and duplicate FileHashes
type DuplicateFile struct {
	Original  FileHash
	Duplicate FileHash
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

func FindDuplicateFiles(fileHashes []FileHash) []DuplicateFile {

	hashes := make(map[string]FileHash)
	duplicateFiles := make([]DuplicateFile, 0, len(fileHashes))
	for _, fileHash := range fileHashes {
		if existingFileHash, exists := hashes[fileHash.Hash]; exists {
			dup := DuplicateFile{
				Original:  existingFileHash,
				Duplicate: fileHash,
			}
			duplicateFiles = append(duplicateFiles, dup)
		} else {
			hashes[fileHash.Hash] = fileHash
		}
	}

	return duplicateFiles
}

// This wrapper is here to possibly turn this into a goroutine
func deleteFile(path string) error {

	//	err := os.Remove(path)
	var err error = nil

	fmt.Printf("'%s' was removed.\n", path)

	return err
}

func promptToDelete(reader *bufio.Reader, dup DuplicateFile) {
	original := dup.Original.AbsolutePath
	duplicate := dup.Duplicate.AbsolutePath

	for done := false; !done; {
		fmt.Println("The following are duplicates:")
		fmt.Printf("1) %s\n", original)
		fmt.Printf("2) %s\n", duplicate)
		fmt.Print("Which file do you wish to remove? [1/2] > ")

		line, _ := reader.ReadString('\n')
		if line == "" {
			continue
		}

		answer := line[0]
		switch answer {
		case '1':
			deleteFile(original)
			done = true
			break
		case '2':
			deleteFile(duplicate)
			done = true
			break
		default:
			fmt.Fprintf(os.Stderr, "'%c' is not acceptable answer.\n", answer)
			break
		}
	}
}

func main() {

	directoryPtr := flag.String("dir", ".", "the directory to search")

	flag.Parse()

	fmt.Printf("Processing directory '%s'.\n", *directoryPtr)

	fileHashes, err := getFileHashes(*directoryPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find files in directory '%s' due to '%s'.\n", *directoryPtr, err.Error())
		os.Exit(1)
	}

	duplicates := FindDuplicateFiles(fileHashes)
	reader := bufio.NewReader(os.Stdin)
	for _, duplicate := range duplicates {
		promptToDelete(reader, duplicate)
	}

	fmt.Println("Done.")
}
