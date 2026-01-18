package ssh

import (
	"fmt"
	"sync"

	"ssh-ftp-proxy/internal/encoder"
	"ssh-ftp-proxy/internal/logger"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// WSMessage represents the JSON message format for WebSocket communication.
type WSMessage struct {
	Type    string `json:"type"`    // "input", "output", "resize", "error"
	Payload string `json:"payload"` // Base64 encoded content
}

func (s *Service) StartInteractive(ws *websocket.Conn) error {
	if err := s.connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Request PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // Enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		return fmt.Errorf("request pty failed: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe failed: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe failed: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe failed: %w", err)
	}

	// Start shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("start shell failed: %w", err)
	}

	// Message loops
	var wg sync.WaitGroup
	wg.Add(3)

	// 1. SSH Stdout -> WS
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				data := encoder.EncodeBytes(buf[:n])
				msg := WSMessage{Type: "output", Payload: data}
				ws.WriteJSON(msg)
			}
			if err != nil {
				return
			}
		}
	}()

	// 2. SSH Stderr -> WS
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				data := encoder.EncodeBytes(buf[:n])
				msg := WSMessage{Type: "output", Payload: data} // Treat stderr as output for simplicity or separate type
				ws.WriteJSON(msg)
			}
			if err != nil {
				return
			}
		}
	}()

	// 3. WS Input -> SSH Stdin
	go func() {
		defer wg.Done()
		for {
			var msg WSMessage
			if err := ws.ReadJSON(&msg); err != nil {
				return
			}

			if msg.Type == "input" {
				data, err := encoder.DecodeBytes(msg.Payload)
				if err != nil {
					logger.Log.Warn("Invalid base64 input", "error", err)
					continue
				}
				stdin.Write(data)
			} else if msg.Type == "resize" {
				// Handle resize payload: "rows,cols"
				// Not implemented in MVP
			}
		}
	}()

	// Wait for session to end
	session.Wait()
	return nil
}
