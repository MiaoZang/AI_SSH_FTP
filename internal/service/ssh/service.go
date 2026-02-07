package ssh

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"ssh-ftp-proxy/internal/config"
	"ssh-ftp-proxy/internal/logger"

	"golang.org/x/crypto/ssh"
)

type Service struct {
	mu     sync.Mutex
	client *ssh.Client
	config config.SSHConfig
}

func NewService(cfg config.SSHConfig) *Service {
	return &Service{
		config: cfg,
	}
}

func (s *Service) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		// Verify connection
		_, _, err := s.client.Conn.SendRequest("keepalive@openssh.com", true, nil)
		if err == nil {
			return nil
		}
		s.client.Close()
	}

	auth := []ssh.AuthMethod{}
	if s.config.Password != "" {
		auth = append(auth, ssh.Password(s.config.Password))
	}
	if s.config.KeyFile != "" {
		key, err := os.ReadFile(s.config.KeyFile)
		if err != nil {
			logger.Log.Warn("Failed to read key file", "error", err)
		} else {
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				logger.Log.Warn("Failed to parse private key", "error", err)
			} else {
				auth = append(auth, ssh.PublicKeys(signer))
			}
		}
	}

	clientConfig := &ssh.ClientConfig{
		User: s.config.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // Insecure: accept any host key
		},
		Timeout: 5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return err
	}

	s.client = client
	return nil
}

func (s *Service) Exec(cmd string) (string, string, int, error) {
	if err := s.connect(); err != nil {
		return "", "", -1, fmt.Errorf("connection failed: %w", err)
	}

	session, err := s.client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(cmd)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			// Other errors (connection lost, etc.)
			// We might want to return the error, but if it ran and failed, exitCode is usually non-zero.
			// If it failed to run (e.g. command not found), exitCode might be 127 but that usually comes as ExitError too.
			exitCode = -1
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode, err
}
