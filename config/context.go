package config

type HasConfig interface {
	GetConfig() *Config
}
