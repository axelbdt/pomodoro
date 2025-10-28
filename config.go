package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds timer configuration
type Config struct {
	WorkMinutes           int    `toml:"work_minutes"`
	ShortBreakMinutes     int    `toml:"short_break_minutes"`
	LongBreakMinutes      int    `toml:"long_break_minutes"`
	WorkSessionsPerCycle  int    `toml:"work_sessions_per_cycle"`
	SoundWorkStart        string `toml:"sound_work_start"`
	SoundBreakStart       string `toml:"sound_break_start"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		WorkMinutes:          25,
		ShortBreakMinutes:    5,
		LongBreakMinutes:     20,
		WorkSessionsPerCycle: 3,
		SoundWorkStart:       "",
		SoundBreakStart:      "",
	}
}

// LoadConfig loads configuration from file, returns defaults on error
func LoadConfig() *Config {
	cfg := DefaultConfig()

	configPath := filepath.Join(os.Getenv("HOME"), ".config", "pomodoro", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found at %s, using defaults", configPath)
		return cfg
	}

	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		log.Printf("Failed to parse config file: %v, using defaults", err)
		return DefaultConfig()
	}

	// Validate config values
	if cfg.WorkMinutes <= 0 {
		log.Printf("Invalid work_minutes: %d, using default 25", cfg.WorkMinutes)
		cfg.WorkMinutes = 25
	}
	if cfg.ShortBreakMinutes <= 0 {
		log.Printf("Invalid short_break_minutes: %d, using default 5", cfg.ShortBreakMinutes)
		cfg.ShortBreakMinutes = 5
	}
	if cfg.LongBreakMinutes <= 0 {
		log.Printf("Invalid long_break_minutes: %d, using default 20", cfg.LongBreakMinutes)
		cfg.LongBreakMinutes = 20
	}
	if cfg.WorkSessionsPerCycle <= 0 {
		log.Printf("Invalid work_sessions_per_cycle: %d, using default 3", cfg.WorkSessionsPerCycle)
		cfg.WorkSessionsPerCycle = 3
	}

	return cfg
}
