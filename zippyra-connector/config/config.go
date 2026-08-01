package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ConnectionID        string `yaml:"connection_id"`
	AgentAPIKey         string `yaml:"agent_api_key"`
	WebhookSecret       string `yaml:"webhook_secret"`
	ZippyraAPIBaseURL   string `yaml:"zippyra_api_base_url"`
	ErpType             string `yaml:"erp_type"` // TALLY or BUSY
	ErpLocalEndpoint    string `yaml:"erp_local_endpoint"`
	PollIntervalSeconds int    `yaml:"poll_interval_seconds"`
	StatusServerPort    int    `yaml:"status_server_port"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	cfg.ErpType = strings.ToUpper(strings.TrimSpace(cfg.ErpType))
	if cfg.ErpType != "TALLY" && cfg.ErpType != "BUSY" {
		return nil, fmt.Errorf("invalid erp_type %q; must be TALLY or BUSY", cfg.ErpType)
	}

	if cfg.ConnectionID == "" {
		return nil, fmt.Errorf("missing connection_id in config")
	}
	if cfg.AgentAPIKey == "" {
		return nil, fmt.Errorf("missing agent_api_key in config")
	}
	if cfg.ZippyraAPIBaseURL == "" {
		return nil, fmt.Errorf("missing zippyra_api_base_url in config")
	}

	if cfg.PollIntervalSeconds <= 0 {
		cfg.PollIntervalSeconds = 60
	}
	if cfg.StatusServerPort <= 0 {
		cfg.StatusServerPort = 8085
	}
	if cfg.ErpLocalEndpoint == "" {
		if cfg.ErpType == "TALLY" {
			cfg.ErpLocalEndpoint = "http://127.0.0.1:9000"
		} else {
			cfg.ErpLocalEndpoint = "http://127.0.0.1:8080/api"
		}
	}

	return &cfg, nil
}
