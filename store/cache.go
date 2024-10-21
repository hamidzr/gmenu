package store

import "os"

// cacheFilePath returns the path to the cache file.
func (fs FileStore[C, Cfg]) cacheFilePath() string {
	return fs.cacheDir + "/cache." + fs.format
}

// SaveCache serializes and saves the cache data to a file.
func (fs FileStore[C, Cfg]) SaveCache(data C) error {
	serialized, err := fs.Marshal(data)
	if err != nil {
		return err
	}
	filePath := fs.cacheFilePath()
	// fmt.Println("Saving cache to", filePath)
	return os.WriteFile(filePath, serialized, 0o644)
}

// LoadCache reads and deserializes the cache data from a file.
func (fs FileStore[C, Cfg]) LoadCache() (C, error) {
	var data C
	filePath := fs.cacheFilePath()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return data, nil
	}
	serialized, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}
	err = fs.Unmarshal(serialized, &data)
	return data, err
}
