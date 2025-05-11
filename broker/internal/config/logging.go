package config

type Log struct {
	// Level with possible values: `DEBUG|ERROR|FATAL|INFO|PANIC|WARN` E.g INFO.
	Level string `default:"DEBUG" envconfig:"LOG_LEVEL"`

	// StdoutEnabled is a toggle whether to log in stdout or not.
	StdoutEnabled bool `default:"true" envconfig:"LOG_STDOUT_ENABLED"`

	// ExcludePath is a list of paths to exclude from the logger.
	ExcludePaths []string `default:"/live" envconfig:"LOG_EXCLUDE_PATHS"`

	// FilePath is the path to the log file.
	FilePath string `default:"../logs/broker.log" envconfig:"LOG_FILE_PATH"`
}
