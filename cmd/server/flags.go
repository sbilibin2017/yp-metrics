package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func parseFlags() *configs.ServerConfig {
	fs := flag.NewFlagSet("server", flag.ExitOnError)

	options := []configs.ServerOption{
		withAddr(fs),
		withStoreInterval(fs),
		withFileStoragePath(fs),
		withRestore(fs),
		withDatabaseDSN(fs),
		withLogLevel(fs),
		withHashKey(fs),
	}

	fs.Parse(os.Args[1:])

	return configs.NewServerConfig(options...)
}

func withAddr(fs *flag.FlagSet) configs.ServerOption {
	var addr string
	fs.StringVar(&addr, "a", ":8080", "address and port to run server")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("ADDRESS"); env != "" {
			cfg.Addr = env
		} else {
			cfg.Addr = addr
		}
	}
}

func withStoreInterval(fs *flag.FlagSet) configs.ServerOption {
	var interval int
	fs.IntVar(&interval, "i", 300, "store interval in seconds (0 = sync write)")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("STORE_INTERVAL"); env != "" {
			if val, err := strconv.Atoi(env); err == nil {
				cfg.StoreInterval = val
				return
			}
		}
		cfg.StoreInterval = interval
	}
}

func withFileStoragePath(fs *flag.FlagSet) configs.ServerOption {
	var path string
	fs.StringVar(&path, "f", "./data/metrics.json", "file storage path")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("FILE_STORAGE_PATH"); env != "" {
			cfg.FileStoragePath = env
		} else {
			cfg.FileStoragePath = path
		}
	}
}

func withRestore(fs *flag.FlagSet) configs.ServerOption {
	var restore bool
	fs.BoolVar(&restore, "r", true, "restore metrics from file at startup")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("RESTORE"); env != "" {
			if val, err := strconv.ParseBool(env); err == nil {
				cfg.Restore = val
				return
			}
		}
		cfg.Restore = restore
	}
}

func withDatabaseDSN(fs *flag.FlagSet) configs.ServerOption {
	var dsn string
	fs.StringVar(&dsn, "d", "", "PostgreSQL DSN")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("DATABASE_DSN"); env != "" {
			cfg.DatabaseDSN = env
		} else {
			cfg.DatabaseDSN = dsn
		}
	}
}

func withLogLevel(fs *flag.FlagSet) configs.ServerOption {
	var level string
	fs.StringVar(&level, "l", "info", "log level")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("LOG_LEVEL"); env != "" {
			cfg.LogLevel = env
		} else {
			cfg.LogLevel = level
		}
	}
}

func withHashKey(fs *flag.FlagSet) configs.ServerOption {
	var key string
	fs.StringVar(&key, "k", "", "hash key")

	return func(cfg *configs.ServerConfig) {
		if env := os.Getenv("KEY"); env != "" {
			cfg.HashKey = env
		} else {
			cfg.HashKey = key
		}
	}
}
