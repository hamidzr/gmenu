package store

type Cache struct {
	UsageCount       map[string]int `json:"usageCount"`
	NotFoundAccepted []string       `json:"notFoundAccepted"`
	LastEntry        string         `json:"lastEntry"`
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
	SaveCache(menuID string, data Cache) error
	LoadCache(menuID string) (Cache, error)
	SaveConfig(menuID string, config Config) error
	LoadConfig(menuID string) (Config, error)
}
