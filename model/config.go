package model

type Config struct {
	AcceptCustomSelection bool
	CliArgs
}

// CliArgs is a struct to hold the root CLI arguments.
type CliArgs struct {
	// Title string
	Title string
	// Menu Prompt string
	Prompt string
	// Menu ID
	MenuID string
	// Search method
	SearchMethod string
	// Preserve the order of the input items.
	PreserveOrder bool
	// initial query
	InitialQuery string
	// TODO: Allow custom output.
	// allowCustomOutput bool
	// AutoAccept indicates whether to auto accept the only item if there's only one match
	AutoAccept bool
	// TerminalMode indicates whether to run in terminal-only mode without GUI
	TerminalMode bool
	// DisableNumericSelection allows the user to select an item by its 1-based index.
	DisableNumericSelection bool
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
