package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func parseFlags() *configs.AgentConfig {
	fs := flag.NewFlagSet("agent", flag.ExitOnError)

	options := []configs.AgentOption{
		withAddress(fs),
		withPollInterval(fs),
		withReportInterval(fs),
		withLogLevel(fs),
		withHashKey(fs),
		withHashHeader(fs),
	}

	fs.Parse(os.Args[1:])

	return configs.NewAgentConfig(options...)
}

func withAddress(fs *flag.FlagSet) configs.AgentOption {
	var addr string
	fs.StringVar(&addr, "a", "http://localhost:8080", "address of HTTP server")

	return func(cfg *configs.AgentConfig) {
		if env := os.Getenv("ADDRESS"); env != "" {
			cfg.Address = env
		} else {
			cfg.Address = addr
		}
	}
}

func withPollInterval(fs *flag.FlagSet) configs.AgentOption {
	var poll int
	fs.IntVar(&poll, "p", 2, "poll interval in seconds")

	return func(cfg *configs.AgentConfig) {
		if env := os.Getenv("POLL_INTERVAL"); env != "" {
			if val, err := strconv.Atoi(env); err == nil {
				cfg.PollInterval = val
				return
			}
		}
		cfg.PollInterval = poll
	}
}

func withReportInterval(fs *flag.FlagSet) configs.AgentOption {
	var report int
	fs.IntVar(&report, "r", 10, "report interval in seconds")

	return func(cfg *configs.AgentConfig) {
		if env := os.Getenv("REPORT_INTERVAL"); env != "" {
			if val, err := strconv.Atoi(env); err == nil {
				cfg.ReportInterval = val
				return
			}
		}
		cfg.ReportInterval = report
	}
}

func withLogLevel(fs *flag.FlagSet) configs.AgentOption {
	var level string
	fs.StringVar(&level, "l", "info", "log level")

	return func(cfg *configs.AgentConfig) {
		if env := os.Getenv("LOG_LEVEL"); env != "" {
			cfg.LogLevel = env
		} else {
			cfg.LogLevel = level
		}
	}
}

func withHashKey(fs *flag.FlagSet) configs.AgentOption {
	var key string
	fs.StringVar(&key, "k", "", "hash key for HMAC signing")

	return func(cfg *configs.AgentConfig) {
		if env := os.Getenv("KEY"); env != "" {
			cfg.HashKey = env
		} else {
			cfg.HashKey = key
		}
	}
}

func withHashHeader(fs *flag.FlagSet) configs.AgentOption {
	var header string
	fs.StringVar(&header, "hh", "HashSHA256", "HTTP header to store HMAC signature")

	return func(cfg *configs.AgentConfig) {
		if env := os.Getenv("HASH_HEADER"); env != "" {
			cfg.HashHeader = env
		} else {
			cfg.HashHeader = header
		}
	}
}
