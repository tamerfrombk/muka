package main

import (
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

func getFilePaths(dir string) ([]string, error) {
	var files []string
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		absolutePath, _ := filepath.Abs(file)
		files = append(files, filepath.ToSlash(absolutePath))

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

func findDuplicateFiles(directory string) ([]DuplicateFile, error) {
	filePaths, err := getFilePaths(directory)

	duplicateFiles := make([]DuplicateFile, 0, len(filePaths))
	if err != nil {
		return duplicateFiles, err
	}

	hashes := make(map[string]FileHash)
	for _, filePath := range filePaths {
		fileHash, err := hashFile(filePath)
		if err != nil {
			fmt.Printf("Could not hash '%s'.\n", filePath)
			continue
		}

		if existingFileHash, exists := hashes[fileHash.Hash]; exists {
			fmt.Printf("'%s' is a duplicate of '%s'!\n", filePath, existingFileHash.AbsolutePath)
			dup := DuplicateFile{
				Original:  existingFileHash,
				Duplicate: fileHash,
			}
			duplicateFiles = append(duplicateFiles, dup)
		} else {
			hashes[fileHash.Hash] = fileHash
		}
	}

	return duplicateFiles, nil
}

func main() {

	directoryPtr := flag.String("dir", ".", "the directory to search")

	flag.Parse()

	fmt.Printf("Processing directory '%s'.\n", *directoryPtr)

	duplicates, err := findDuplicateFiles(*directoryPtr)
	if err != nil {
		fmt.Println(err)
	}

	for _, duplicate := range duplicates {
		fmt.Printf("'%s' is a duplicate of '%s'.\n", duplicate.Duplicate.AbsolutePath, duplicate.Original.AbsolutePath)
	}

	fmt.Println("Done.")
}
