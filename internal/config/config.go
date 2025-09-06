package config

type Config struct {
	Server    Server
	Logger    Logger
	Database  Database
	Exchanges Exchanges
}

type Server struct {
	Port string `yaml:"port"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
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
