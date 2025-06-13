package configs

type AgentConfig struct {
	Address        string
	PollInterval   int
	ReportInterval int
	LogLevel       string
	HashKey        string
	HashHeader     string
}

type AgentOption func(cfg *AgentConfig)

func NewAgentConfig(opts ...AgentOption) *AgentConfig {
	cfg := &AgentConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
