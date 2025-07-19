package store

// cacheFilePath returns the path to the cache file.
func (fs FileStore[C, Cfg]) cacheFilePath() string {
	return fs.buildFilePath(fs.cacheDir, "cache")
}

// SaveCache serializes and saves the cache data to a file.
func (fs FileStore[C, Cfg]) SaveCache(data C) error {
	filePath := fs.cacheFilePath()
	return fs.saveData(data, filePath)
}

// LoadCache reads and deserializes the cache data from a file.
func (fs FileStore[C, Cfg]) LoadCache() (C, error) {
	var data C
	filePath := fs.cacheFilePath()
	err := fs.loadData(filePath, &data, true) // Allow missing cache files
	return data, err
}
