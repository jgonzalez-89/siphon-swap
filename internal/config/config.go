package config

import (
	"strconv"
)

type Config struct {
	Server    Server    `yaml:"server"`
	Logger    Logger    `yaml:"logger"`
	Database  Database  `yaml:"database"`
	Exchanges Exchanges `yaml:"exchanges"`
	Daemon    Daemon    `yaml:"daemon"`
	Messaging RabbitMQ  `yaml:"messaging"`
}

type Daemon struct {
	Enabled string `yaml:"enabled"`
}

func (c *Config) IsDaemonEnabled() bool {
	return c.Daemon.Enabled == "true"
}

type Server struct {
	Port string `yaml:"port"`
}

type RabbitMQ struct {
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Exchange       string `yaml:"exchange"`
	PrefetchCount  string `yaml:"prefetch_count"`
	MaxReconnects  string `yaml:"max_reconnects"`
	ReconnectDelay string `yaml:"reconnect_delay"`
}

func (r *RabbitMQ) GetPrefetchCount() int {
	if r.PrefetchCount == "" {
		return 1
	}
	return parseInt(r.PrefetchCount)
}
func (r *RabbitMQ) GetMaxReconnects() int {
	if r.MaxReconnects == "" {
		return 5
	}
	return parseInt(r.MaxReconnects)
}

func (r *RabbitMQ) GetReconnectDelay() int {
	if r.ReconnectDelay == "" {
		return 5
	}
	return parseInt(r.ReconnectDelay)
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

type Logger struct {
	Level string `yaml:"level"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
}

type Exchanges struct {
	ChangeNow Exchange `yaml:"change_now"`
	StealthEx Exchange `yaml:"stealthex"`
	CoinGecko Exchange `yaml:"coingecko"`
}

type Exchange struct {
	ApiKey         string `yaml:"api_key"`
	AuthHeader     string `yaml:"auth_header"`
	AuthScheme     string `yaml:"auth_scheme"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	BaseURL        string `yaml:"base_url"`
}
