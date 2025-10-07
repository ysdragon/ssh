package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"

	"ssh/internal/config"
	"ssh/internal/handlers"
)

func checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".write_test")
	err := os.WriteFile(testFile, []byte(""), 0600)
	if err != nil {
		return err
	}
	os.Remove(testFile)
	return nil
}

func loadOrCreateHostKey() (gossh.Signer, error) {
	// Try to use root directory first
	hostKeyPath := filepath.Join("/", "ssh_host_rsa_key")

	// Check if we have permission to write to root directory
	if err := checkWritePermission("/"); err != nil {
		// Fallback to home directory if no permission in root
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory cannot be determined
			homeDir = "/"
		}
		hostKeyPath = filepath.Join(homeDir, "ssh_host_rsa_key")
	}

	// Check if host key exists
	if _, err := os.Stat(hostKeyPath); os.IsNotExist(err) {
		// Generate a new RSA key
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		// Encode the private key
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		})

		// Write the private key to file
		err = os.WriteFile(hostKeyPath, privateKeyPEM, 0600)
		if err != nil {
			return nil, err
		}

		// Create signer from the new key
		signer, err := gossh.NewSignerFromKey(privateKey)
		if err != nil {
			return nil, err
		}

		color.Yellow("Created new host key at %s", hostKeyPath)
		return signer, nil
	} else if err != nil {
		return nil, err
	}

	// Load existing host key
	privateKeyBytes, err := os.ReadFile(hostKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := gossh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func main() {
	// Load configuration from YAML file
	cfg, err := config.LoadConfig()
	if err != nil {
		color.Red("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Load or create persistent host key
	hostKey, err := loadOrCreateHostKey()
	if err != nil {
		color.Red("Failed to load/create host key: %v", err)
		os.Exit(1)
	}

	// Set default port if not configured
	if cfg.SSH.Port == "" {
		cfg.SSH.Port = "2222"
	}

	// Parse timeout to duration
	var sshTimeout time.Duration
	if cfg.SSH.Timeout > 0 {
		sshTimeout = time.Duration(cfg.SSH.Timeout) * time.Second
	}

	// Check for hashed passwords (bcrypt or argon2)

	server := &ssh.Server{
		Addr: ":" + cfg.SSH.Port,
		PasswordHandler: func(ctx ssh.Context, pass string) bool {
			// Make sure username matches and check password
			success := cfg.SSH.User == ctx.User() && config.CheckPassword(cfg.SSH.Password, pass)
			handlers.LogLoginAttempt(ctx.RemoteAddr().String(), ctx.User(), success, "password")
			return success
		},
	}

	// Add the host key to the server
	server.AddHostKey(hostKey)

	if cfg.SFTP.Enable {
		server.SubsystemHandlers = map[string]ssh.SubsystemHandler{
			"sftp": handlers.SFTPHandler,
		}
	}

	if cfg.SSH.Password == "" {
		server.PasswordHandler = nil
	}

	server.Handle(handlers.SessionHandler)

	if sshTimeout > 0 {
		server.MaxTimeout = sshTimeout
		server.IdleTimeout = sshTimeout
		color.Yellow("SSH server configured with timeouts:")
		color.Yellow("  - Maximum connection duration: %s", sshTimeout)
		color.Yellow("  - Idle timeout: %s", sshTimeout)
	}

	color.Yellow("  - User: %s", cfg.SSH.User)
	if config.IsBcryptHash(cfg.SSH.Password) {
		color.Yellow("  - Using bcrypt hashed password")
	} else if config.IsArgon2Hash(cfg.SSH.Password) {
		color.Yellow("  - Using argon2 hashed password")
	}
	color.Yellow("  - SFTP enabled: %v", cfg.SFTP.Enable)
	color.Blue("Starting SSH server on port %s...", cfg.SSH.Port)
	color.Yellow("  - Type 'q' to exit.")

	// Start the SSH server in a separate goroutine
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	// Scanner to detect 'q' input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "q" {
			color.Yellow("Exit command detected. Closing the SSH server.")
			os.Exit(0)
		}
	}
}
