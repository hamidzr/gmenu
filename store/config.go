package store

// configFilePath returns the path to the config file.
func (fs FileStore[C, Cfg]) configFilePath() string {
	return fs.buildFilePath(fs.configDir, "config")
}

// SaveConfig serializes and saves the config data to a file.
func (fs FileStore[C, Cfg]) SaveConfig(config Cfg) error {
	filePath := fs.configFilePath()
	return fs.saveData(config, filePath)
}

// LoadConfig reads and deserializes the config data from a file.
func (fs FileStore[C, Cfg]) LoadConfig() (Cfg, error) {
	var config Cfg
	filePath := fs.configFilePath()
	err := fs.loadData(filePath, &config, false) // Don't allow missing config files
	return config, err
}
