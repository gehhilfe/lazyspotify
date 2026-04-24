package utils

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const SpotifyClientIDHelpURL = "https://github.com/dubeyKartikay/lazyspotify?tab=readme-ov-file#set-up-your-spotify-client-id"
const NavidromeConfigHelpURL = "https://github.com/dubeyKartikay/lazyspotify?tab=readme-ov-file#navidrome-backend"

const (
	BackendSpotify   = "spotify"
	BackendNavidrome = "navidrome"
)

const (
	appConfigFileName           = "config.yml"
	spotifyClientIDPlaceholder  = "your_spotify_app_client_id"
	defaultAppConfigFileContent = "backend: spotify\nauth:\n  client_id: your_spotify_app_client_id\n"
)

var (
	config        AppConfig
	configLoadErr error
)

func GetConfig() AppConfig {
	return config
}

func init() {
	config, configLoadErr = LoadConfig()
	if configLoadErr != nil {
		config = getDefaultAppConfig()
	}
}

type AppConfig struct {
	LogLevel string `mapstructure:"log_level"`
	Backend  string `mapstructure:"backend"`
	Auth     struct {
		ClientID         string `mapstructure:"client_id"`
		Host             string `mapstructure:"host"`
		Port             int    `mapstructure:"port"`
		RedirectEndpoint string `mapstructure:"redirect-endpoint"`
		Timeout          int    `mapstructure:"timeout"`
		Keyring          struct {
			Service string `mapstructure:"service"`
			Key     string `mapstructure:"key"`
		} `mapstructure:"keyring"`
	} `mapstructure:"auth"`
	Librespot struct {
		Host       string `mapstructure:"host"`
		Port       int    `mapstructure:"port"`
		Timeout    int    `mapstructure:"timeout"`
		RetryDelay int    `mapstructure:"retry-delay"`
		MaxRetries int    `mapstructure:"max-retries"`
		SeekStepMs int    `mapstructure:"seek-step-ms"`
		VolumeStep int    `mapstructure:"volume-step"`
		Daemon     struct {
			Cmd             []string `mapstructure:"cmd"`
			LogLevel        string   `mapstructure:"log_level"`
			ZeroconfEnabled bool     `mapstructure:"zeroconf_enabled"`
		} `mapstructure:"daemon"`
	} `mapstructure:"librespot"`
	Navidrome struct {
		ServerURL          string `mapstructure:"server_url"`
		Username           string `mapstructure:"username"`
		AuthMethod         string `mapstructure:"auth_method"`
		StreamFormat       string `mapstructure:"stream_format"`
		InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
		Keyring            struct {
			Service string `mapstructure:"service"`
			Key     string `mapstructure:"key"`
		} `mapstructure:"keyring"`
	} `mapstructure:"navidrome"`
	Player struct {
		SeekStepMs int `mapstructure:"seek_step_ms"`
		VolumeStep int `mapstructure:"volume_step"`
		Mpv        struct {
			Cmd []string `mapstructure:"cmd"`
		} `mapstructure:"mpv"`
	} `mapstructure:"player"`
}

func (c AppConfig) SpotifyClientID() string {
	return strings.TrimSpace(c.Auth.ClientID)
}

func getDefaultAppConfig() AppConfig {
	cfg := AppConfig{}
	cfg.LogLevel = "ERROR"
	cfg.Backend = BackendSpotify
	cfg.Auth.Host = "127.0.0.1"
	cfg.Auth.Port = 8287
	cfg.Auth.RedirectEndpoint = "/callback"
	cfg.Auth.Timeout = 30
	cfg.Auth.Keyring.Service = "spotify"
	cfg.Auth.Keyring.Key = "token-v2"
	cfg.Librespot.Host = "127.0.0.1"
	cfg.Librespot.Port = 4040
	cfg.Librespot.Timeout = 180
	cfg.Librespot.RetryDelay = 100
	cfg.Librespot.MaxRetries = 3
	cfg.Librespot.SeekStepMs = 5000
	cfg.Librespot.VolumeStep = 65535 / 20
	cfg.Librespot.Daemon.LogLevel = "ERROR"
	cfg.Navidrome.AuthMethod = "password"
	cfg.Navidrome.StreamFormat = "raw"
	cfg.Navidrome.Keyring.Service = "lazyspotify"
	cfg.Navidrome.Keyring.Key = "navidrome-password"
	cfg.Player.SeekStepMs = 5000
	cfg.Player.VolumeStep = 5
	cfg.Player.Mpv.Cmd = []string{"mpv"}
	return cfg
}

func LoadConfig() (AppConfig, error) {
	configDir, err := ensureAppConfigFile()
	if err != nil {
		return AppConfig{}, err
	}

	v := viper.New()
	applyConfigDefaults(v)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)
	err = v.ReadInConfig()
	var configErr viper.ConfigFileNotFoundError
	if err != nil && !errors.As(err, &configErr) {
		return AppConfig{}, err
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	var config AppConfig
	err = v.Unmarshal(&config)
	if err != nil {
		return AppConfig{}, err
	}
	return config, nil
}

func ValidateStartupConfig() error {
	if configLoadErr != nil {
		return fmt.Errorf("failed to load config: %w", configLoadErr)
	}
	return validateStartupConfig(config)
}

func validateStartupConfig(cfg AppConfig) error {
	backend := cfg.Backend
	if backend == "" {
		backend = BackendSpotify
	}
	switch backend {
	case BackendSpotify:
		if clientID := cfg.SpotifyClientID(); clientID == "" || clientID == spotifyClientIDPlaceholder {
			return fmt.Errorf("missing required config value `auth.client_id`; see %s", SpotifyClientIDHelpURL)
		}
	case BackendNavidrome:
		if strings.TrimSpace(cfg.Navidrome.ServerURL) == "" {
			return fmt.Errorf("missing required config value `navidrome.server_url`; see %s", NavidromeConfigHelpURL)
		}
		if strings.TrimSpace(cfg.Navidrome.Username) == "" {
			return fmt.Errorf("missing required config value `navidrome.username`; see %s", NavidromeConfigHelpURL)
		}
	default:
		return fmt.Errorf("unknown backend %q; expected %q or %q", backend, BackendSpotify, BackendNavidrome)
	}
	return nil
}

func ensureAppConfigFile() (string, error) {
	configDir := getConfigDir()
	if configDir == "" {
		return "", fmt.Errorf("failed to resolve user config directory")
	}
	if err := EnsureExists(configDir); err != nil {
		return "", err
	}

	configPath := filepath.Join(configDir, appConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return configDir, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if err := os.WriteFile(configPath, []byte(defaultAppConfigFileContent), 0644); err != nil {
		return "", err
	}
	return configDir, nil
}

func getConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	configDir := filepath.Join(dir, "lazyspotify")
	return configDir
}

func SafeGetConfigDir() string {
	configDir := getConfigDir()
	EnsureExists(configDir)
	return configDir
}

func applyConfigDefaults(v *viper.Viper) {
	defaults := getDefaultAppConfig()
	v.SetDefault("log_level", defaults.LogLevel)
	v.SetDefault("backend", defaults.Backend)
	v.SetDefault("auth.host", defaults.Auth.Host)
	v.SetDefault("auth.port", defaults.Auth.Port)
	v.SetDefault("auth.redirect-endpoint", defaults.Auth.RedirectEndpoint)
	v.SetDefault("auth.timeout", defaults.Auth.Timeout)
	v.SetDefault("auth.keyring.service", defaults.Auth.Keyring.Service)
	v.SetDefault("auth.keyring.key", defaults.Auth.Keyring.Key)
	v.SetDefault("librespot.host", defaults.Librespot.Host)
	v.SetDefault("librespot.port", defaults.Librespot.Port)
	v.SetDefault("librespot.timeout", defaults.Librespot.Timeout)
	v.SetDefault("librespot.retry-delay", defaults.Librespot.RetryDelay)
	v.SetDefault("librespot.max-retries", defaults.Librespot.MaxRetries)
	v.SetDefault("librespot.seek-step-ms", defaults.Librespot.SeekStepMs)
	v.SetDefault("librespot.volume-step", defaults.Librespot.VolumeStep)
	v.SetDefault("librespot.daemon.log_level", defaults.Librespot.Daemon.LogLevel)
	v.SetDefault("librespot.daemon.zeroconf_enabled", defaults.Librespot.Daemon.ZeroconfEnabled)
	v.SetDefault("navidrome.server_url", defaults.Navidrome.ServerURL)
	v.SetDefault("navidrome.username", defaults.Navidrome.Username)
	v.SetDefault("navidrome.auth_method", defaults.Navidrome.AuthMethod)
	v.SetDefault("navidrome.stream_format", defaults.Navidrome.StreamFormat)
	v.SetDefault("navidrome.insecure_skip_verify", defaults.Navidrome.InsecureSkipVerify)
	v.SetDefault("navidrome.keyring.service", defaults.Navidrome.Keyring.Service)
	v.SetDefault("navidrome.keyring.key", defaults.Navidrome.Keyring.Key)
	v.SetDefault("player.seek_step_ms", defaults.Player.SeekStepMs)
	v.SetDefault("player.volume_step", defaults.Player.VolumeStep)
	v.SetDefault("player.mpv.cmd", defaults.Player.Mpv.Cmd)
}
