package configs

type ServerConfig struct {
	Addr            string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	LogLevel        string
	HashKey         string
}

type ServerOption func(*ServerConfig)

func NewServerConfig(opts ...ServerOption) *ServerConfig {
	cfg := &ServerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
