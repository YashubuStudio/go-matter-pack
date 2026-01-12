package app

// Config holds application configuration values.
type Config struct {
	StateDir string
}

// DefaultConfig returns a Config with default values.
func DefaultConfig(appName string) Config {
	return Config{StateDir: StateDir(appName)}
}
