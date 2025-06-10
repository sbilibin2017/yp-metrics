package configs

type ServerConfig struct {
	LogLevel   string
	RunAddress string
}

type ServerOption func(*ServerConfig)

func NewServerConfig(opts ...ServerOption) *ServerConfig {
	cfg := &ServerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithServerLogLevel() ServerOption {
	return func(cfg *ServerConfig) {
		cfg.LogLevel = "info"
	}
}

func WithServerRunAddress(address string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RunAddress = address
	}
}
