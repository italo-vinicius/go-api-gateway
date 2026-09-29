package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration: %w", err)
	}
	var c Config
	if err = yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	c.ApplyDefaults()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}
