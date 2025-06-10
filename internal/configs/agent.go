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

func WithAgentLogLevel() AgentOption {
	return func(cfg *AgentConfig) {
		cfg.LogLevel = "info"
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

func WithAgentServerEndpoint() AgentOption {
	return func(cfg *AgentConfig) {
		cfg.ServerEndpoint = "/update"
	}
}
