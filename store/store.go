package store

import "time"

type Cache struct {
	UsageCount       map[string]int `json:"usageCount"`
	NotFoundAccepted []string       `json:"notFoundAccepted"`
	// LastEntry is the last entry that was selected by the user.
	LastEntry     string `json:"lastEntry"`
	LastEntryTime int64  `json:"lastEntryTime"`
	// LastInput is the last input that was entered by the user.
	LastInput string `json:"lastInput"`
}

func (c *Cache) SetLastEntry(entry string) {
	c.LastEntry = entry
	c.LastEntryTime = time.Now().Unix()
}

// SetLastInput sets the last input to the cache.
func (c *Cache) SetLastInput(input string) {
	c.LastInput = input
}

type Config struct {
	AppTitle      string `json:"appTitle"`
	DefaultPrompt string `json:"defaultPrompt"`
	DefaultLimit  int    `json:"defaultLimit"`
	WindowSize    struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"windowSize"`
	SearchMethod string `json:"searchMethod"`
}

type Store interface {
	SaveCache(data Cache) error
	LoadCache() (Cache, error)
	SaveConfig(config Config) error
	LoadConfig() (Config, error)
}
