package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func createDefaultConfig() error {
	defaultConfig := `
SSH_PORT=
SSH_USER=
SSH_PASSWORD=
SSH_TIMEOUT=
SFTP_ENABLE=
`
	return os.WriteFile("/.ssh_config", []byte(defaultConfig), 0644)
}

func getConfigValue(key string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	_, err := os.Stat("/.ssh_config")
	if os.IsNotExist(err) {
		color.Yellow("/.ssh_config not found. Creating with default values.")
		if err := createDefaultConfig(); err != nil {
			color.Red("Error creating /.ssh_config: %v", err)
			return ""
		}
	} else if err != nil {
		color.Red("Error accessing /.ssh_config: %v", err)
		return ""
	}

	content, err := os.ReadFile("/.ssh_config")
	if err != nil {
		color.Red("Error reading /.ssh_config: %v", err)
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
		if len(parts) == 2 && parts[0] == key {
			return strings.TrimSpace(parts[1])
		}
	}

	return ""
}

func getTimeoutValue() time.Duration {
	valueStr := getConfigValue("SSH_TIMEOUT")
	if valueStr == "" {
		return 0
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		color.Red("Error parsing SSH_TIMEOUT: %v", err)
		return 0
	}

	return time.Duration(value) * time.Second
}

func sftpHandler(sess ssh.Session) {
	debugStream := io.Discard
	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}
	server, err := sftp.NewServer(sess, serverOptions...)
	if err != nil {
		color.Red("sftp server init error: %s", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
		color.Green("sftp client exited session.")
	} else if err != nil {
		color.Red("sftp server completed with error: %s", err)
	}
}

func logLoginAttempt(ip, user string, success bool, method string) {
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("%s - IP: %s, User: %s, Method: %s, Success: %v", timestamp, ip, user, method, success)

	if success {
		color.Green(logEntry)
	} else {
		color.Red(logEntry)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error getting home directory: %v", err)
		return
	}

	logFile := filepath.Join(homeDir, "ssh.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		color.Red("Error opening log file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry + "\n"); err != nil {
		color.Red("Error writing to log file: %v", err)
	}
}

func handleSession(s ssh.Session) {
	cmd := exec.Command("sh")
	ptyReq, winCh, isPty := s.Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			color.Red("Error starting pty: %v", err)
			io.WriteString(s, fmt.Sprintf("Error starting pty: %v\n", err))
			s.Exit(1)
			return
		}
		go func() {
			for win := range winCh {
				setWinsize(f, win.Width, win.Height)
			}
		}()
		go func() {
			io.Copy(f, s)
		}()
		io.Copy(s, f)
		cmd.Wait()
	} else {
		io.WriteString(s, "No PTY requested.\n")
		s.Exit(1)
	}
}

func main() {
	port := getConfigValue("SSH_PORT")
	if port == "" {
		port = "2222"
	}

	sshUser := getConfigValue("SSH_USER")
	sshPassword := getConfigValue("SSH_PASSWORD")
	sshTimeout := getTimeoutValue()
	sftpEnabled := getConfigValue("SFTP_ENABLE") == "true"

	server := &ssh.Server{
		Addr: ":" + port,
		PasswordHandler: func(ctx ssh.Context, pass string) bool {
			success := sshPassword != "" && ctx.User() == sshUser && pass == sshPassword
			logLoginAttempt(ctx.RemoteAddr().String(), ctx.User(), success, "password")
			return success
		},
	}

	if sftpEnabled {
		server.SubsystemHandlers = map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
		}
	}

	if sshPassword == "" {
		server.PasswordHandler = nil
	}

	server.Handle(handleSession)

	if sshTimeout > 0 {
		server.MaxTimeout = sshTimeout
		server.IdleTimeout = sshTimeout
		color.Yellow("SSH server configured with timeouts:")
		color.Yellow("  - Maximum connection duration: %s", sshTimeout)
		color.Yellow("  - Idle timeout: %s", sshTimeout)
	}

	color.Blue("Starting SSH server on port %s...", port)

	color.Yellow("  - Press 'q' to exit.")

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