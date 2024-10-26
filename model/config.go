package model

type Config struct {
	AcceptCustomSelection bool
}

// var (
// 	activeConfig     *Config
// 	onceActiveConfig sync.Once
// )

// func GetActiveConfig() *Config {
// 	onceActiveConfig.Do(func() {
// 		activeConfig = DefaultConfig()
// 	})
// 	return activeConfig
// }

func DefaultConfig() Config {
	return Config{
		AcceptCustomSelection: true,
	}
}
