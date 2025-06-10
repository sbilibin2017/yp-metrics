package configs

type AgentConfig struct {
	LogLevel         string
	ServerRunAddress string
	ServerEndpoint   string
	PollInterval     int
	ReportInterval   int
}

type AgentOption func(*AgentConfig)

func NewAgentConfig(opts ...AgentOption) *AgentConfig {
	cfg := &AgentConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithAgentLogLevel(level string) AgentOption {
	return func(cfg *AgentConfig) {
		cfg.LogLevel = level
	}
}

func WithAgentServerRunAddress(address string) AgentOption {
	return func(cfg *AgentConfig) {
		cfg.ServerRunAddress = address
	}
}

func WithAgentPollInterval(interval int) AgentOption {
	return func(cfg *AgentConfig) {
		cfg.PollInterval = interval
	}
}

func WithAgentReportInterval(interval int) AgentOption {
	return func(cfg *AgentConfig) {
		cfg.ReportInterval = interval
	}
}

func WithAgentServerEndpoint(endpoint string) AgentOption {
	return func(cfg *AgentConfig) {
		cfg.ServerEndpoint = endpoint
	}
}
