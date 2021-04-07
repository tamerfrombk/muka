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

	"github.com/fatih/color"
)

// FileSizeCache is a mapping between the size of a file and the count
type FileSizeCache map[int64]int

// FileData holds basic metadata about a file
type FileData struct {
	AbsolutePath string
	SizeInBytes  int64
}

// FileHash defines the file hash
type FileHash struct {
	FileData
	Hash string
}

func (hash FileHash) String() string {

	return hash.AbsolutePath
}

// Directory holds the collected information about a directory
type Directory struct {
	EncounteredFiles []FileData
	HashedFiles      []FileHash
}

// DuplicateFile holds original and duplicate FileHashes
type DuplicateFile struct {
	Original   FileHash
	Duplicates []FileHash
}

func (duplicate DuplicateFile) String() string {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	bold := color.New(color.Bold)

	var b strings.Builder

	bold.Fprint(&b, "Original: ")

	green.Fprintf(&b, "%s\n", duplicate.Original)
	dupLength := len(duplicate.Duplicates)
	if dupLength == 0 {
		return b.String()
	}

	bold.Fprint(&b, "Duplicates: [ ")
	for i := 0; i < dupLength-1; i++ {
		red.Fprint(&b, duplicate.Duplicates[i].String())
		bold.Fprint(&b, ", ")
	}

	red.Fprint(&b, duplicate.Duplicates[dupLength-1].String())
	bold.Fprint(&b, " ]\n")

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

	unitSizeFn := func(sizeInKb float64) (string, float64) {
		unit := "KB"
		if sizeInKb/1_000_000 > 1 {
			unit = "GB"
			sizeInKb /= 1_000_000
		} else if sizeInKb/1_000 > 1 {
			unit = "MB"
			sizeInKb /= 1_000
		}

		return unit, sizeInKb
	}

	bold := color.New(color.Bold)
	var b strings.Builder

	scannedUnit, scannedSize := unitSizeFn(r.CollectedFileSizeInKB)
	bold.Fprintf(&b, "Files Scanned: %d (%.2f %s)\n", r.CollectedFileCount, scannedSize, scannedUnit)

	dupUnit, dupSize := unitSizeFn(r.DuplicateFileSizeInKB)
	bold.Fprintf(&b, "Duplicates Found: %d (%.2f %s), %.2f%% of scanned files\n",
		r.DuplicateFileCount, dupSize, dupUnit, r.DuplicatePercentage)

	delUnit, delSize := unitSizeFn(r.DeletedFileSizeInKB)
	bold.Fprintf(&b, "%d files were deleted saving %.2f %s\n", r.DeletedFileCount, delSize, delUnit)

	return b.String()
}

// CollectFiles Recursively walks the provided directory and processes each file and directory it encounters
// This function will return, in order, a list of all the files it encountered, a list of files that were hashed,
// and an error if one is encountered.
func CollectFiles(options FileCollectionOptions) (Directory, error) {
	var fileData []FileData
	sizeCache := make(FileSizeCache)
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

		fileData = append(fileData, FileData{
			AbsolutePath: absolutePath,
			SizeInBytes:  info.Size(),
		})

		return nil
	})

	if err != nil {
		return Directory{}, err
	}

	return Directory{
		EncounteredFiles: fileData,
		HashedFiles:      hashFiles(fileData, sizeCache),
	}, nil
}

func hashFiles(fileData []FileData, sizeCache FileSizeCache) []FileHash {

	fileHashes := make([]FileHash, 0, len(fileData))
	for _, fd := range fileData {
		// If the file has a unique size, there is no way it could be a duplicate
		// so we avoid having to hash it for performance reasons
		if sizeCache[fd.SizeInBytes] > 1 {
			h, err := hashFile(fd.AbsolutePath)
			if err == nil {
				fileHashes = append(fileHashes, FileHash{
					FileData: fd,
					Hash:     h,
				})
			} else {
				log.Printf("hashing %q : %v", fd.AbsolutePath, err)
			}
		}
	}

	return fileHashes
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
func FindDuplicateFiles(directory Directory) []DuplicateFile {

	cache := NewCache()
	for _, fileHash := range directory.HashedFiles {
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
func CalculateReport(directory Directory, duplicates []DuplicateFile, deletedFiles []FileHash) Report {

	sum := func(hashes []FileHash) int64 {
		sum := int64(0)
		for _, f := range hashes {
			sum += f.SizeInBytes
		}
		return sum
	}

	sumOfFileSizes := int64(0)
	for _, fd := range directory.EncounteredFiles {
		sumOfFileSizes += fd.SizeInBytes
	}

	sumOfDuplicateSizes := int64(0)
	sumOfDuplicateCount := 0
	for _, f := range duplicates {
		sumOfDuplicateSizes += sum(f.Duplicates)
		sumOfDuplicateCount += len(f.Duplicates)
	}

	sumOfDeletedFileSizes := sum(deletedFiles)

	return Report{
		CollectedFileCount:    len(directory.EncounteredFiles),
		CollectedFileSizeInKB: float64(sumOfFileSizes) / 1000.0,
		DuplicateFileCount:    sumOfDuplicateCount,
		DuplicateFileSizeInKB: float64(sumOfDuplicateSizes) / 1000.0,
		DuplicatePercentage:   (float64(sumOfDuplicateCount) / float64(len(directory.EncounteredFiles))) * 100,
		DeletedFileCount:      len(deletedFiles),
		DeletedFileSizeInKB:   float64(sumOfDeletedFileSizes) / 1000.0,
	}
}
