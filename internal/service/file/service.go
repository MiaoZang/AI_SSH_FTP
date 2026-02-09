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

// DetailedFileInfo represents detailed file metadata
type DetailedFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	Mode    string `json:"mode"`
}

// Mkdir creates a directory (with parents if needed)
func (s *Service) Mkdir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	logger.Log.Info("Directory created", "path", path)
	return nil
}

// Rename moves/renames a file or directory
func (s *Service) Rename(src, dst string) error {
	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	logger.Log.Info("File renamed", "src", src, "dst", dst)
	return nil
}

// Copy copies a file or directory
func (s *Service) Copy(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	if srcInfo.IsDir() {
		return s.copyDir(src, dst)
	}
	return s.copyFile(src, dst)
}

func (s *Service) copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy file permissions
	srcInfo, _ := os.Stat(src)
	os.Chmod(dst, srcInfo.Mode())

	logger.Log.Info("File copied", "src", src, "dst", dst)
	return nil
}

func (s *Service) copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	logger.Log.Info("Directory copied", "src", src, "dst", dst)
	return nil
}

// GetInfo returns detailed information about a file or directory
func (s *Service) GetInfo(path string) (*DetailedFileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &DetailedFileInfo{
		Name:    info.Name(),
		Path:    path,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime().Unix(),
		Mode:    info.Mode().String(),
	}, nil
}

// DeleteDir removes a directory and its contents
func (s *Service) DeleteDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}
	logger.Log.Info("Directory deleted", "path", path)
	return nil
}

// BatchDeleteResult represents the result of batch delete
type BatchDeleteResult struct {
	Success []string           `json:"success"`
	Failed  []BatchDeleteError `json:"failed"`
}

// BatchDeleteError represents a single delete error
type BatchDeleteError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// BatchDelete deletes multiple files/directories
func (s *Service) BatchDelete(paths []string) *BatchDeleteResult {
	result := &BatchDeleteResult{
		Success: []string{},
		Failed:  []BatchDeleteError{},
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			result.Failed = append(result.Failed, BatchDeleteError{Path: path, Error: err.Error()})
			continue
		}

		if info.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil {
			result.Failed = append(result.Failed, BatchDeleteError{Path: path, Error: err.Error()})
		} else {
			result.Success = append(result.Success, path)
		}
	}

	logger.Log.Info("Batch delete completed", "success", len(result.Success), "failed", len(result.Failed))
	return result
}
