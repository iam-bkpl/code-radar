package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Environment struct {
	Name    string   `yaml:"name" json:"name"`
	Pattern []string `yaml:"pattern" json:"pattern"`
	Color   string   `yaml:"color" json:"color"`
}

type Config struct {
	Environments []Environment `yaml:"environments" json:"environments"`
}

func defaultConfig() Config {
	return Config{
		Environments: []Environment{
			{Name: "DEV", Pattern: []string{"develop", "dev"}, Color: "#6272A4"},
			{Name: "QA", Pattern: []string{"qa", "test"}, Color: "#F1FA8C"},
			{Name: "UAT", Pattern: []string{"uat"}, Color: "#FFB86C"},
			{Name: "STAGING", Pattern: []string{"staging", "stg"}, Color: "#8BE9FD"},
			{Name: "MASTER", Pattern: []string{"master", "main"}, Color: "#BD93F9"},
			{Name: "PROD", Pattern: []string{"prod", "production"}, Color: "#FF5555"},
		},
	}
}

func loadConfig() Config {
	if data, err := os.ReadFile(".code-radar.yaml"); err == nil {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err == nil && len(cfg.Environments) > 0 {
			return cfg
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".config", "code-radar", "config.yaml")
		if data, err := os.ReadFile(globalPath); err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err == nil && len(cfg.Environments) > 0 {
				return cfg
			}
		}
	}

	return defaultConfig()
}

func initConfig() error {
	path := ".code-radar.yaml"
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	cfg := defaultConfig()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func matchEnvironment(branch string, cfg Config) string {
	lower := strings.ToLower(branch)
	for _, env := range cfg.Environments {
		for _, pattern := range env.Pattern {
			p := strings.ToLower(pattern)
			if p == lower || strings.Contains(lower, p) {
				return env.Name
			}
		}
	}
	return "OTHER"
}

func categorizeBranches(branches []string, cfg Config) map[string][]string {
	result := make(map[string][]string)
	for _, b := range branches {
		env := matchEnvironment(b, cfg)
		result[env] = append(result[env], b)
	}
	return result
}
