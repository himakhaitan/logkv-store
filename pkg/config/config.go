package config

type Config struct {
	DataDir string
}

func Load() (*Config, error) {
	// TODO: Load configuration from yaml files or environment variables
	return &Config{
		DataDir: "data",
	}, nil
}
