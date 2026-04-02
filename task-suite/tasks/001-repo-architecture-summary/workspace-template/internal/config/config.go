package config

type Config struct {
	Workspace string
}

func Load() Config {
	return Config{Workspace: "."}
}
