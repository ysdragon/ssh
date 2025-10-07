package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration structure
type Config struct {
	SSH struct {
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"ssh"`
	SFTP struct {
		Enable bool `yaml:"enable"`
	} `yaml:"sftp"`
}

var configPath string

func checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".write_test")
	err := os.WriteFile(testFile, []byte(""), 0600)
	if err != nil {
		return err
	}
	os.Remove(testFile)
	return nil
}

func init() {
	// Try to use root directory first
	configPath = filepath.Join("/", "ssh_config.yml")

	// Check if we have permission to write to root directory
	if err := checkWritePermission("/"); err != nil {
		// Fallback to home directory if no permission in root
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory cannot be determined
			homeDir = "/"
		}
		configPath = filepath.Join(homeDir, "ssh_config.yml")
	}
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig() error {
	defaultConfig := Config{}
	defaultConfig.SSH.Port = "2222"
	defaultConfig.SSH.User = "root"
	defaultConfig.SSH.Password = "password"
	defaultConfig.SSH.Timeout = 300
	defaultConfig.SFTP.Enable = true

	yamlData, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, yamlData, 0644)
}

// LoadConfig loads the configuration from the YAML file
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Check if config file exists, create if not
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		color.Yellow("Configuration file not found. Creating default config at %s", configPath)
		if err := CreateDefaultConfig(); err != nil {
			color.Red("Error creating default config: %v", err)
			return nil, err
		}
	} else if err != nil {
		color.Red("Error checking config file: %v", err)
		return nil, err
	}

	// Read the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		color.Red("Error reading config file: %v", err)
		return nil, err
	}

	// Parse YAML into config struct
	if err := yaml.Unmarshal(content, cfg); err != nil {
		color.Red("Error parsing config: %v", err)
		return nil, err
	}

	return cfg, nil
}

// IsBcryptHash detects if a string is a bcrypt hash
func IsBcryptHash(str string) bool {
	return len(str) > 0 && (strings.HasPrefix(str, "$2a$") ||
		strings.HasPrefix(str, "$2b$") ||
		strings.HasPrefix(str, "$2y$"))
}

// CheckPassword checks password - handles both bcrypt and plaintext
func CheckPassword(storedPassword, inputPassword string) bool {
	// If it looks like a bcrypt hash, use bcrypt comparison
	if IsBcryptHash(storedPassword) {
		err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(inputPassword))
		return err == nil
	}

	// Otherwise, use plain text comparison
	return storedPassword == inputPassword
}
