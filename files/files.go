package files

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileHash defines the file hash
type FileHash struct {
	AbsolutePath string
	Hash         string
}

func (hash FileHash) String() string {

	return hash.AbsolutePath
}

// DuplicateFile holds original and duplicate FileHashes
type DuplicateFile struct {
	Original   FileHash
	Duplicates []FileHash
}

func (duplicate DuplicateFile) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Original: %s\n", duplicate.Original)
	dupLength := len(duplicate.Duplicates)
	if dupLength == 0 {
		return b.String()
	}

	b.WriteString("Duplicates: [ ")
	for i := 0; i < dupLength-1; i++ {
		b.WriteString(duplicate.Duplicates[i].String())
		b.WriteString(", ")
	}

	b.WriteString(duplicate.Duplicates[dupLength-1].String())
	b.WriteString(" ]\n")

	return b.String()
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

// PromptToDelete this function interactively prompts the user to delete the duplicate
func PromptToDelete(writer io.Writer, reader io.Reader, deleter Deleter, dup DuplicateFile) {

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

// PrintDuplicates print the duplicate to stdout
func PrintDuplicates(duplicates []DuplicateFile) {
	for _, duplicate := range duplicates {
		fmt.Println(duplicate)
	}
}

// ForceDelete deletes the duplicates without asking for user intervention
func ForceDelete(duplicates []DuplicateFile, deleter Deleter) {
	for _, dup := range duplicates {
		for _, f := range dup.Duplicates {
			deleter.Delete(f.AbsolutePath)
		}
	}
}