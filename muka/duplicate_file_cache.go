package muka

// DuplicateFileCache holds a cache of possible duplicate muka
type DuplicateFileCache struct {
	fileHashByHash map[string]FileHash
	duplicates     []DuplicateFile
}

// NewCache duplicateFileCache constructor
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

	allDuplicates := cache.duplicates

	nonEmptyDuplicates := make([]DuplicateFile, 0, len(allDuplicates))
	for _, d := range allDuplicates {
		if len(d.Duplicates) > 0 && d.Original.Hash != "" {
			nonEmptyDuplicates = append(nonEmptyDuplicates, d)
		}
	}

	return nonEmptyDuplicates
}
