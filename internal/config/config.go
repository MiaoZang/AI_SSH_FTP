package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig `mapstructure:"server"`
	SSHServer SSHConfig    `mapstructure:"ssh_server"`
	FTPServer FTPConfig    `mapstructure:"ftp_server"`
	Log       LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	HTTPPort int    `mapstructure:"http_port"`
	WSPort   int    `mapstructure:"ws_port"`
	BindIP   string `mapstructure:"bind_ip"`
}

type SSHConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	KeyFile  string `mapstructure:"key_file"`
}

type FTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

var GlobalConfig Config

func LoadConfig(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
