package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ssh-ftp-proxy/internal/config"
	"ssh-ftp-proxy/internal/encoder"
	"ssh-ftp-proxy/internal/logger"
	"ssh-ftp-proxy/internal/service/file"
	"ssh-ftp-proxy/internal/service/ftp"
	"ssh-ftp-proxy/internal/service/ssh"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine      *gin.Engine
	sshService  *ssh.Service
	ftpService  *ftp.Service
	fileService *file.Service
}

func NewServer() *Server {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware())

	s := &Server{
		engine:      engine,
		sshService:  ssh.NewService(config.GlobalConfig.SSHServer),
		ftpService:  ftp.NewService(config.GlobalConfig.FTPServer),
		fileService: file.NewService(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.engine.GET("/api/health", s.handleHealth)

	sshGroup := s.engine.Group("/api/ssh")
	{
		sshGroup.POST("/exec", s.handleSSHExec)
	}

	ftpGroup := s.engine.Group("/api/ftp")
	{
		ftpGroup.POST("/list", s.handleFTPList)
		ftpGroup.POST("/upload", s.handleFTPUpload)
		ftpGroup.POST("/download", s.handleFTPDownload)
	}

	// New file API (HTTP multipart upload)
	fileGroup := s.engine.Group("/api/file")
	{
		fileGroup.POST("/upload", s.handleFileUpload)
		fileGroup.POST("/list", s.handleFileList)
		fileGroup.POST("/download", s.handleFileDownload)
		fileGroup.POST("/delete", s.handleFileDelete)
		// New file operations
		fileGroup.POST("/mkdir", s.handleFileMkdir)
		fileGroup.POST("/rename", s.handleFileRename)
		fileGroup.POST("/copy", s.handleFileCopy)
		fileGroup.POST("/info", s.handleFileInfo)
		fileGroup.POST("/batch/delete", s.handleFileBatchDelete)
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", config.GlobalConfig.Server.BindIP, config.GlobalConfig.Server.HTTPPort)
	logger.Log.Info("Starting HTTP Server", "addr", addr)
	return s.engine.Run(addr)
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		c.Next()
		logger.Log.Info("Request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
		)
	}
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type SSHExecRequest struct {
	Command string `json:"command" binding:"required"` // Base64 encoded
}

type SSHExecResponse struct {
	Stdout   string `json:"stdout"` // Base64 encoded
	Stderr   string `json:"stderr"` // Base64 encoded
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"` // Base64 encoded
}

func (s *Server) handleSSHExec(c *gin.Context) {
	var req SSHExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, s.newErrorResponse(fmt.Sprintf("Invalid request: %v", err)))
		return
	}

	// 1. Decode Command
	cmd, err := encoder.Decode(req.Command)
	if err != nil {
		c.JSON(http.StatusBadRequest, s.newErrorResponse(fmt.Sprintf("Invalid base64 command: %v", err)))
		return
	}

	logger.Log.Debug("Executing SSH command", "command", cmd)

	// 2. Execute
	stdout, stderr, exitCode, execErr := s.sshService.Exec(cmd)

	// 3. Encode Response
	resp := SSHExecResponse{
		Stdout:   encoder.Encode(stdout),
		Stderr:   encoder.Encode(stderr),
		ExitCode: exitCode,
	}

	if execErr != nil {
		resp.Error = encoder.Encode(execErr.Error())
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) newErrorResponse(msg string) SSHExecResponse {
	return SSHExecResponse{
		Error:    encoder.Encode(msg),
		ExitCode: 1, // Indicate error
	}
}

// ============ File API Handlers ============

// FileUploadResponse represents the response for file upload
type FileUploadResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Error   string `json:"error,omitempty"`
}

// handleFileUpload handles multipart file upload
// Supports optional auto-extract for tar.gz/zip files
func (s *Server) handleFileUpload(c *gin.Context) {
	// Get destination path (base64 encoded)
	pathB64 := c.PostForm("path")
	if pathB64 == "" {
		c.JSON(http.StatusBadRequest, FileUploadResponse{Error: "path is required"})
		return
	}

	destPath, err := encoder.Decode(pathB64)
	if err != nil {
		c.JSON(http.StatusBadRequest, FileUploadResponse{Error: "invalid base64 path"})
		return
	}

	logger.Log.Debug("File upload request", "destPath", destPath)

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Log.Error("Failed to get file from form", "error", err)
		c.JSON(http.StatusBadRequest, FileUploadResponse{Error: "file is required: " + err.Error()})
		return
	}

	logger.Log.Debug("Received file", "filename", fileHeader.Filename, "size", fileHeader.Size)

	// Open uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		logger.Log.Error("Failed to open uploaded file", "error", err)
		c.JSON(http.StatusInternalServerError, FileUploadResponse{Error: err.Error()})
		return
	}
	defer src.Close()

	// Determine full destination path
	// If destPath ends with / or is an existing directory, append filename
	fullPath := destPath
	if strings.HasSuffix(destPath, "/") || strings.HasSuffix(destPath, "\\") {
		// Path ends with separator, treat as directory
		fullPath = filepath.Join(destPath, fileHeader.Filename)
	} else if info, err := os.Stat(destPath); err == nil && info.IsDir() {
		// Path is an existing directory
		fullPath = filepath.Join(destPath, fileHeader.Filename)
	}

	logger.Log.Debug("Saving file", "fullPath", fullPath)

	// Save file
	if err := s.fileService.SaveFile(src, fullPath); err != nil {
		logger.Log.Error("Failed to save file", "error", err, "path", fullPath)
		c.JSON(http.StatusInternalServerError, FileUploadResponse{Error: err.Error()})
		return
	}

	logger.Log.Info("File saved", "path", fullPath, "size", fileHeader.Size)

	// Check if auto-extract is requested
	extract := c.PostForm("extract")
	if extract == "true" {
		// Extract to same directory as archive
		extractDir := filepath.Dir(fullPath)
		logger.Log.Debug("Extracting archive", "archive", fullPath, "destDir", extractDir)
		if err := s.fileService.ExtractArchive(fullPath, extractDir); err != nil {
			logger.Log.Error("Failed to extract archive", "error", err)
			c.JSON(http.StatusOK, FileUploadResponse{
				Success: true,
				Path:    fullPath,
				Size:    fileHeader.Size,
				Error:   "upload success but extract failed: " + err.Error(),
			})
			return
		}
		// Delete archive after successful extraction
		os.Remove(fullPath)
		logger.Log.Info("File uploaded and extracted", "path", extractDir)
	}

	c.JSON(http.StatusOK, FileUploadResponse{
		Success: true,
		Path:    fullPath,
		Size:    fileHeader.Size,
	})
}

// FileListRequest represents the request for file listing
type FileListRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

// handleFileList lists directory contents
func (s *Server) handleFileList(c *gin.Context) {
	var req FileListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	dirPath, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 path"})
		return
	}

	files, err := s.fileService.ListDir(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// FileDownloadRequest represents the request for file download
type FileDownloadRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

// handleFileDownload downloads a file
func (s *Server) handleFileDownload(c *gin.Context) {
	var req FileDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	filePath, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 path"})
		return
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot download directory"})
		return
	}

	// Open and read file
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": encoder.EncodeBytes(content),
		"name":    filepath.Base(filePath),
		"size":    info.Size(),
	})
}

// FileDeleteRequest represents the request for file deletion
type FileDeleteRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

// handleFileDelete deletes a file or directory
func (s *Server) handleFileDelete(c *gin.Context) {
	var req FileDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	filePath, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 path"})
		return
	}

	// Remove file or directory
	if err := os.RemoveAll(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// FileMkdirRequest represents the request for mkdir
type FileMkdirRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

// handleFileMkdir creates a directory
func (s *Server) handleFileMkdir(c *gin.Context) {
	var req FileMkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	dirPath, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 path"})
		return
	}

	if err := s.fileService.Mkdir(dirPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "path": dirPath})
}

// FileRenameRequest represents the request for rename/move
type FileRenameRequest struct {
	Src string `json:"src" binding:"required"` // Base64 encoded
	Dst string `json:"dst" binding:"required"` // Base64 encoded
}

// handleFileRename moves/renames a file or directory
func (s *Server) handleFileRename(c *gin.Context) {
	var req FileRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "src and dst are required"})
		return
	}

	srcPath, err := encoder.Decode(req.Src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 src"})
		return
	}

	dstPath, err := encoder.Decode(req.Dst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 dst"})
		return
	}

	if err := s.fileService.Rename(srcPath, dstPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "src": srcPath, "dst": dstPath})
}

// FileCopyRequest represents the request for copy
type FileCopyRequest struct {
	Src string `json:"src" binding:"required"` // Base64 encoded
	Dst string `json:"dst" binding:"required"` // Base64 encoded
}

// handleFileCopy copies a file or directory
func (s *Server) handleFileCopy(c *gin.Context) {
	var req FileCopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "src and dst are required"})
		return
	}

	srcPath, err := encoder.Decode(req.Src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 src"})
		return
	}

	dstPath, err := encoder.Decode(req.Dst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 dst"})
		return
	}

	if err := s.fileService.Copy(srcPath, dstPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "src": srcPath, "dst": dstPath})
}

// FileInfoRequest represents the request for file info
type FileInfoRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

// handleFileInfo returns detailed file information
func (s *Server) handleFileInfo(c *gin.Context) {
	var req FileInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	filePath, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 path"})
		return
	}

	info, err := s.fileService.GetInfo(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// FileBatchDeleteRequest represents the request for batch delete
type FileBatchDeleteRequest struct {
	Paths []string `json:"paths" binding:"required"` // Array of Base64 encoded paths
}

// handleFileBatchDelete deletes multiple files/directories
func (s *Server) handleFileBatchDelete(c *gin.Context) {
	var req FileBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paths array is required"})
		return
	}

	// Decode all paths
	decodedPaths := make([]string, 0, len(req.Paths))
	for _, p := range req.Paths {
		decoded, err := encoder.Decode(p)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid base64 path: %s", p)})
			return
		}
		decodedPaths = append(decodedPaths, decoded)
	}

	result := s.fileService.BatchDelete(decodedPaths)
	c.JSON(http.StatusOK, result)
}
