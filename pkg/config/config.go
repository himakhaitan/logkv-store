package config

import "time"

type Config struct {
	DataDir       string
	MergeInterval time.Duration
}

func Load() (*Config, error) {
	// TODO: Load configuration from yaml files or environment variables
	return &Config{
		DataDir:       "data",
		MergeInterval: 30 * time.Minute,
	}, nil
}
