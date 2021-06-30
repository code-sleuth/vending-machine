package config

import (
	"github.com/code-sleuth/vending-machine/helpers"
)

// Config structure
type Config struct {
	DB *DBConfig
}

// DBConfig structure
type DBConfig struct {
	Dialect    string
	Host       string
	Port       string
	Username   string
	Password   string
	DBName     string
	TestDBName string
	SSLMode    string
}

// GetConfig function
func GetConfig() *Config {
	return &Config{
		DB: &DBConfig{
			Dialect:    helpers.GetEnv("DB_DIALECT", ""),
			Host:       helpers.GetEnv("DB_HOST", ""),
			Port:       helpers.GetEnv("DB_PORT", ""),
			Username:   helpers.GetEnv("DB_USERNAME", ""),
			Password:   helpers.GetEnv("DB_PASSWORD", ""),
			DBName:     helpers.GetEnv("DB_NAME", ""),
			TestDBName: helpers.GetEnv("TEST_DB_NAME", ""),
			SSLMode:    helpers.GetEnv("DB_SSL_MODE", ""),
		},
	}
}
