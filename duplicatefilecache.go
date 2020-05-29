package main

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

type duplicateFileCache struct {
	fileHashByHash map[string]FileHash
	duplicates     []DuplicateFile
}

// NewCache duplicateFileCache constructor
func NewCache() duplicateFileCache {

	return duplicateFileCache{
		fileHashByHash: make(map[string]FileHash),
		duplicates:     make([]DuplicateFile, 0),
	}
}

func (cache *duplicateFileCache) findDuplicateFile(hash FileHash) int {

	for i, dup := range cache.duplicates {
		if dup.Original.Hash == hash.Hash {
			return i
		}
	}

	return -1
}

// Add adds a FileHash to the cache accounting for possible duplicates
func (cache *duplicateFileCache) Add(hash FileHash) {
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
func (cache *duplicateFileCache) GetDuplicates() []DuplicateFile {

	return cache.duplicates
}
