package store

import "os"

// confiFilePath returns the path to the config file.
func (fs FileStore[C, Cfg]) configFilePath() string {
	return fs.configDir + "/config." + fs.format
}

// SaveConfig serializes and saves the config data to a file.
func (fs FileStore[C, Cfg]) SaveConfig(config Cfg) error {
	serialized, err := fs.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.configFilePath(), serialized, 0o644)
}

// LoadConfig reads and deserializes the config data from a file.
func (fs FileStore[C, Cfg]) LoadConfig() (Cfg, error) {
	var config Cfg
	serialized, err := os.ReadFile(fs.configFilePath())
	if err != nil {
		return config, err
	}
	err = fs.Unmarshal(serialized, &config)
	return config, err
}
