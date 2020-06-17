package muka

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileHash defines the file hash
type FileHash struct {
	AbsolutePath string
	Hash         string
	SizeInBytes  int64
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

// FileCollectionOptions options used by CollectFiles
type FileCollectionOptions struct {
	DirectoryToSearch string
	ExcludeDirs       []*regexp.Regexp
	ExcludeFiles      []*regexp.Regexp
}

// Report reports on program performance
type Report struct {
	CollectedFileCount    int
	CollectedFileSizeInKB float64
	DuplicateFileCount    int
	DuplicateFileSizeInKB float64
	DuplicatePercentage   float64
	DeletedFileCount      int
	DeletedFileSizeInKB   float64
}

func (r Report) String() string {

	var b strings.Builder

	fmt.Fprintf(&b, "Files Scanned: %d (%.2f KB)\n", r.CollectedFileCount, r.CollectedFileSizeInKB)
	fmt.Fprintf(&b, "Duplicates Found: %d (%.2f KB), %.2f%% of scanned files\n",
		r.DuplicateFileCount, r.DuplicateFileSizeInKB, r.DuplicatePercentage)
	fmt.Fprintf(&b, "%d files were deleted saving %.2f KB\n", r.DeletedFileCount, r.DeletedFileSizeInKB)

	return b.String()
}

// CollectFiles Recursively walks the provided directory and creates FileHash for each encountered file
func CollectFiles(options FileCollectionOptions) ([]FileHash, error) {
	var files []FileHash
	sizeCache := make(map[int64]int)
	err := filepath.Walk(filepath.Clean(options.DirectoryToSearch), func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, excludeDirPattern := range options.ExcludeDirs {
				if excludeDirPattern.MatchString(info.Name()) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		for _, excludeFilePattern := range options.ExcludeFiles {
			if excludeFilePattern.MatchString(info.Name()) {
				return nil
			}
		}

		absolutePath, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		if n, exists := sizeCache[info.Size()]; exists {
			sizeCache[info.Size()] = n + 1
		} else {
			sizeCache[info.Size()] = 1
		}

		files = append(files, FileHash{
			AbsolutePath: absolutePath,
			SizeInBytes:  info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(files); i++ {
		f := &files[i]
		// If the file has a unique size, there is no way it could be a duplicate
		// so we avoid having to hash it for performance reasons
		if sizeCache[f.SizeInBytes] >= 2 {
			hash, err := hashFile(f.AbsolutePath)
			if err == nil {
				f.Hash = hash
			} else {
				log.Printf("hashing %q : %v", f.AbsolutePath, err)
			}
		}
	}

	return files, nil
}

func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum([]byte{})), nil
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
// and returns any files the user has deleted
func PromptToDelete(writer io.Writer, reader io.Reader, deleter Deleter, dup DuplicateFile) ([]FileHash, error) {

	bufReader := bufio.NewReader(reader)
	for {
		fmt.Fprintln(writer, dup)
		fmt.Fprintf(writer, "Which file(s) do you wish to remove? [o/d/s] > ")

		line, err := bufReader.ReadString('\n')
		if err != nil {
			return []FileHash{}, err
		}

		if line == "\n" {
			continue
		}

		switch answer := line[0]; answer {
		case 'd':
			deletedFiles := make([]FileHash, 0, len(dup.Duplicates))
			for _, d := range dup.Duplicates {
				if err := deleter.Delete(d.AbsolutePath); err == nil {
					deletedFiles = append(deletedFiles, d)
				} else {
					log.Printf("unable to delete %q: %v", d.AbsolutePath, err)
				}
			}
			return deletedFiles, nil
		case 'o':
			if err := deleter.Delete(dup.Original.AbsolutePath); err == nil {
				return []FileHash{dup.Original}, err
			} else {
				log.Printf("unable to delete %q: %v", dup.Original.AbsolutePath, err)
				return []FileHash{}, err
			}
		case 's':
			return []FileHash{}, nil
		default:
			log.Printf("%q is not acceptable answer", answer)
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

// ForceDelete deletes the duplicates without asking for user interventionand returns
// all of the deleted files
func ForceDelete(duplicates []DuplicateFile, deleter Deleter) []FileHash {
	var deletedFiles []FileHash
	for _, dup := range duplicates {
		for _, f := range dup.Duplicates {
			if err := deleter.Delete(f.AbsolutePath); err == nil {
				deletedFiles = append(deletedFiles, f)
			} else {
				log.Printf("unable to delete %q: %v", f.AbsolutePath, err)
			}
		}
	}

	return deletedFiles
}

// CompileSpaceSeparatedPatterns takes a string of space separated regexes
// and compiles them into regex state machines.
// If the input is empty, an empty array is returned with no errors
// If any error occurs while compiling one of the patterns, that error
// is returned along with whatever patterns where previously compiled.
func CompileSpaceSeparatedPatterns(s string) ([]*regexp.Regexp, error) {

	var patterns []*regexp.Regexp
	if len(s) == 0 {
		return patterns, nil
	}

	tokens := strings.Split(s, " ")
	for _, token := range tokens {
		pattern, err := regexp.Compile(token)
		if err != nil {
			return patterns, err
		}
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// CalculateReport Generates a report detailing basic program behavior
func CalculateReport(fileHashes []FileHash, duplicates []DuplicateFile, deletedFiles []FileHash) Report {

	sum := func(hashes []FileHash) int64 {
		sum := int64(0)
		for _, f := range hashes {
			sum += f.SizeInBytes
		}
		return sum
	}

	sumOfFileSizes := sum(fileHashes)

	sumOfDuplicateSizes := int64(0)
	sumOfDuplicateCount := 0
	for _, f := range duplicates {
		sumOfDuplicateSizes += sum(f.Duplicates)
		sumOfDuplicateCount += len(f.Duplicates)
	}

	sumOfDeletedFileSizes := sum(deletedFiles)

	return Report{
		CollectedFileCount:    len(fileHashes),
		CollectedFileSizeInKB: float64(sumOfFileSizes) / 1000.0,
		DuplicateFileCount:    sumOfDuplicateCount,
		DuplicateFileSizeInKB: float64(sumOfDuplicateSizes) / 1000.0,
		DuplicatePercentage:   (float64(sumOfDuplicateCount) / float64(len(fileHashes))) * 100,
		DeletedFileCount:      len(deletedFiles),
		DeletedFileSizeInKB:   float64(sumOfDeletedFileSizes) / 1000.0,
	}
}
