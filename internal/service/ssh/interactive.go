package ssh

import (
	"context"
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
	return s.StartInteractiveWithContext(context.Background(), ws)
}

func (s *Service) StartInteractiveWithContext(ctx context.Context, ws *websocket.Conn) error {
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

	// Create cancellable context for goroutines
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Message loops with proper cleanup
	var wg sync.WaitGroup
	wg.Add(3)

	// 1. SSH Stdout -> WS
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			select {
			case <-childCtx.Done():
				return
			default:
				n, err := stdout.Read(buf)
				if n > 0 {
					data := encoder.EncodeBytes(buf[:n])
					msg := WSMessage{Type: "output", Payload: data}
					ws.WriteJSON(msg)
				}
				if err != nil {
					cancel() // Signal other goroutines to stop
					return
				}
			}
		}
	}()

	// 2. SSH Stderr -> WS
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			select {
			case <-childCtx.Done():
				return
			default:
				n, err := stderr.Read(buf)
				if n > 0 {
					data := encoder.EncodeBytes(buf[:n])
					msg := WSMessage{Type: "output", Payload: data}
					ws.WriteJSON(msg)
				}
				if err != nil {
					cancel() // Signal other goroutines to stop
					return
				}
			}
		}
	}()

	// 3. WS Input -> SSH Stdin
	go func() {
		defer wg.Done()
		for {
			select {
			case <-childCtx.Done():
				return
			default:
				var msg WSMessage
				if err := ws.ReadJSON(&msg); err != nil {
					cancel() // Signal other goroutines to stop
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
		}
	}()

	// Wait for session to end
	sessionDone := make(chan error, 1)
	go func() {
		sessionDone <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		cancel()
		session.Close()
	case err := <-sessionDone:
		cancel()
		if err != nil {
			logger.Log.Debug("Session ended", "error", err)
		}
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished
	case <-ctx.Done():
		// Context cancelled, force exit
	}

	return nil
}
