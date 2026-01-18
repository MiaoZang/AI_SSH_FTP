package ftp

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"ssh-ftp-proxy/internal/config"

	"github.com/jlaffaye/ftp"
)

type Service struct {
	config config.FTPConfig
}

func NewService(cfg config.FTPConfig) *Service {
	return &Service{
		config: cfg,
	}
}

func (s *Service) connect() (*ftp.ServerConn, error) {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to dial ftp: %w", err)
	}

	if err := c.Login(s.config.User, s.config.Password); err != nil {
		c.Quit()
		return nil, fmt.Errorf("failed to login ftp: %w", err)
	}

	return c, nil
}

type Entry struct {
	Name string `json:"name"`
	Type string `json:"type"` // "file" or "dir"
	Size uint64 `json:"size"`
	Time string `json:"time"`
}

func (s *Service) List(path string) ([]Entry, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	defer c.Quit()

	entries, err := c.List(path)
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	var result []Entry
	for _, e := range entries {
		entryType := "file"
		if e.Type == ftp.EntryTypeFolder {
			entryType = "dir"
		}
		result = append(result, Entry{
			Name: e.Name,
			Type: entryType,
			Size: e.Size,
			Time: e.Time.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (s *Service) Upload(path string, content []byte) error {
	c, err := s.connect()
	if err != nil {
		return err
	}
	defer c.Quit()

	reader := bytes.NewReader(content)
	if err := c.Stor(path, reader); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}

func (s *Service) Download(path string) ([]byte, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	defer c.Quit()

	r, err := c.Retr(path)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer r.Close()

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	return buf, nil
}
