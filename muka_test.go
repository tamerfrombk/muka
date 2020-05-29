package main

import (
	"testing"
)

func TestDuplicateFileWithZeroDuplicates(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 1 {
		t.Fatalf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles))
	}

	if len(duplicateFiles[0].Duplicates) != 0 {
		t.Fatalf("Expected '%d' entry but found '%d' instead.\n", 0, len(duplicateFiles[0].Duplicates))
	}
}

func TestDuplicateFileWithOneDuplicate(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
		{
			AbsolutePath: "file2.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 1 {
		t.Fatalf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles))
	}

	if len(duplicateFiles[0].Duplicates) != 1 {
		t.Fatalf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles[0].Duplicates))
	}
}

func TestDuplicateFileWithMoreThanOneDuplicate(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
		{
			AbsolutePath: "file2.txt",
			Hash:         "abcdefg",
		},
		{
			AbsolutePath: "file3.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 1 {
		t.Fatalf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles))
	}

	if len(duplicateFiles[0].Duplicates) != 2 {
		t.Fatalf("Expected '%d' entry but found '%d' instead.\n", 2, len(duplicateFiles[0].Duplicates))
	}
}
