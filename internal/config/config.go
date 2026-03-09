package config

import (
	"fmt"
	"os"
	"strconv"
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
	// Set defaults
	viper.SetDefault("server.http_port", 48891)
	viper.SetDefault("server.ws_port", 48892)
	viper.SetDefault("server.bind_ip", "0.0.0.0")
	viper.SetDefault("ssh_server.host", "127.0.0.1")
	viper.SetDefault("ssh_server.port", 22)
	viper.SetDefault("ssh_server.user", "root")
	viper.SetDefault("ssh_server.password", "")
	viper.SetDefault("ssh_server.key_file", "")
	viper.SetDefault("ftp_server.host", "127.0.0.1")
	viper.SetDefault("ftp_server.port", 21)
	viper.SetDefault("ftp_server.user", "root")
	viper.SetDefault("ftp_server.password", "")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "logs/server.log")

	// Viper env binding: APP_SERVER_HTTP_PORT -> server.http_port
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to load config file (optional - not fatal if missing)
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Config file not found (%s), using defaults + env vars\n", path)
	}

	// Apply SFTP_ prefixed env vars (more intuitive for AI)
	applyEnvOverrides()

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// FTP defaults to SSH values if not explicitly set
	if GlobalConfig.FTPServer.Host == "127.0.0.1" && os.Getenv("SFTP_FTP_HOST") == "" && os.Getenv("APP_FTP_SERVER_HOST") == "" {
		GlobalConfig.FTPServer.Host = GlobalConfig.SSHServer.Host
	}
	if GlobalConfig.FTPServer.User == "root" && os.Getenv("SFTP_FTP_USER") == "" && os.Getenv("APP_FTP_SERVER_USER") == "" {
		GlobalConfig.FTPServer.User = GlobalConfig.SSHServer.User
	}
	if GlobalConfig.FTPServer.Password == "" && os.Getenv("SFTP_FTP_PASS") == "" && os.Getenv("APP_FTP_SERVER_PASSWORD") == "" {
		GlobalConfig.FTPServer.Password = GlobalConfig.SSHServer.Password
	}

	return nil
}

// applyEnvOverrides reads SFTP_* environment variables for simple AI usage
// Example: SFTP_SSH_HOST=127.0.0.1 SFTP_SSH_PORT=22 SFTP_SSH_USER=root SFTP_SSH_PASS=xxx ./ssh-ftp-proxy
func applyEnvOverrides() {
	envMap := map[string]string{
		"SFTP_HTTP_PORT": "server.http_port",
		"SFTP_WS_PORT":   "server.ws_port",
		"SFTP_BIND_IP":   "server.bind_ip",
		"SFTP_SSH_HOST":  "ssh_server.host",
		"SFTP_SSH_PORT":  "ssh_server.port",
		"SFTP_SSH_USER":  "ssh_server.user",
		"SFTP_SSH_PASS":  "ssh_server.password",
		"SFTP_SSH_KEY":   "ssh_server.key_file",
		"SFTP_FTP_HOST":  "ftp_server.host",
		"SFTP_FTP_PORT":  "ftp_server.port",
		"SFTP_FTP_USER":  "ftp_server.user",
		"SFTP_FTP_PASS":  "ftp_server.password",
		"SFTP_LOG_LEVEL": "log.level",
		"SFTP_LOG_FILE":  "log.file",
	}

	for envKey, viperKey := range envMap {
		if val := os.Getenv(envKey); val != "" {
			// Try int conversion for port fields
			if strings.HasSuffix(viperKey, "_port") || viperKey == "ssh_server.port" {
				if intVal, err := strconv.Atoi(val); err == nil {
					viper.Set(viperKey, intVal)
					continue
				}
			}
			viper.Set(viperKey, val)
		}
	}
}
