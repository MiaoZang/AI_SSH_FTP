package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ssh-ftp-proxy/internal/logger"
)

// Service handles file operations via SSH
type Service struct{}

// NewService creates a new file service
func NewService() *Service {
	return &Service{}
}

// SaveFile saves uploaded file to the specified path
func (s *Service) SaveFile(content io.Reader, destPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy content
	written, err := io.Copy(out, content)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	logger.Log.Info("File saved", "path", destPath, "size", written)
	return nil
}

// ExtractArchive extracts tar.gz or zip files to destination directory
func (s *Service) ExtractArchive(archivePath, destDir string) error {
	// Detect archive type
	ext := strings.ToLower(filepath.Ext(archivePath))

	switch {
	case strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz"):
		return s.extractTarGz(archivePath, destDir)
	case ext == ".zip":
		return s.extractZip(archivePath, destDir)
	case ext == ".tar":
		return s.extractTar(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

func (s *Service) extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	return s.extractTarReader(tar.NewReader(gzr), destDir)
}

func (s *Service) extractTar(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return s.extractTarReader(tar.NewReader(file), destDir)
}

func (s *Service) extractTarReader(tr *tar.Reader, destDir string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	logger.Log.Info("Archive extracted", "path", destDir)
	return nil
}

func (s *Service) extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		outFile, err := os.Create(target)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	logger.Log.Info("Zip extracted", "path", destDir)
	return nil
}

// DeleteFile removes a file
func (s *Service) DeleteFile(path string) error {
	return os.Remove(path)
}

// ListDir lists directory contents
func (s *Service) ListDir(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}
	return files, nil
}

// FileInfo represents file metadata
type FileInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
}
