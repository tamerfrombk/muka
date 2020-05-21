package main

import (
	"crypto/sha256"
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

		rel := file
		if dir != "." {
			rel = file[len(dir)+1:]
		}

		absolutePath, _ := filepath.Abs(rel)
		files = append(files, filepath.ToSlash(absolutePath))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func hashFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func processFiles(directory string) (map[string]FileHash, error) {
	filePaths, err := getFilePaths(directory)

	hashes := make(map[string]FileHash)
	if err != nil {
		return hashes, err
	}

	for i := range filePaths {
		filePath := filePaths[i]
		hash, err := hashFile(filePath)
		if err != nil {
			fmt.Printf("Could not hash '%s'.\n", filePath)
			continue
		}

		if existingFile, exists := hashes[hash]; exists {
			fmt.Printf("'%s' is a duplicate of '%s'!\n", filePath, existingFile.AbsolutePath)
		} else {
			hashes[hash] = FileHash{
				AbsolutePath: filePath,
				Hash:         hash,
			}
		}
	}

	return hashes, nil
}

func main() {

	directoryPtr := flag.String("dir", ".", "the directory to search")

	flag.Parse()

	fmt.Printf("Processing directory '%s'.\n", *directoryPtr)

	hashes, err := processFiles(*directoryPtr)
	if err != nil {
		fmt.Println(err)
	}

	for p, h := range hashes {
		fmt.Printf("'%s => '%s'\n", p, h)
	}

	fmt.Println("Done.")
}
