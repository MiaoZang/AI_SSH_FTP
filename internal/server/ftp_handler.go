package server

import (
	"fmt"
	"net/http"

	"ssh-ftp-proxy/internal/encoder"
	"ssh-ftp-proxy/internal/service/ftp"

	"github.com/gin-gonic/gin"
)

type FTPListRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

type FTPListResponse struct {
	Entries []ftp.Entry `json:"entries"`
	Error   string      `json:"error,omitempty"` // Base64 encoded
}

func (s *Server) handleFTPList(c *gin.Context) {
	var req FTPListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid request: %v", err)))
		return
	}

	path, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid base64 path: %v", err)))
		return
	}

	entries, err := s.ftpService.List(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, s.newFTPErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, FTPListResponse{Entries: entries})
}

type FTPUploadRequest struct {
	Path    string `json:"path" binding:"required"`    // Base64 encoded
	Content string `json:"content" binding:"required"` // Base64 encoded
}

func (s *Server) handleFTPUpload(c *gin.Context) {
	var req FTPUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid request: %v", err)))
		return
	}

	path, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid base64 path: %v", err)))
		return
	}

	// Manually decode content because it might be binary
	contentBytes, err := encoder.DecodeBytes(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid base64 content: %v", err)))
		return
	}

	if err := s.ftpService.Upload(path, contentBytes); err != nil {
		c.JSON(http.StatusInternalServerError, s.newFTPErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type FTPDownloadRequest struct {
	Path string `json:"path" binding:"required"` // Base64 encoded
}

type FTPDownloadResponse struct {
	Content string `json:"content"`         // Base64 encoded
	Error   string `json:"error,omitempty"` // Base64 encoded
}

func (s *Server) handleFTPDownload(c *gin.Context) {
	var req FTPDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid request: %v", err)))
		return
	}

	path, err := encoder.Decode(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, s.newFTPErrorResponse(fmt.Sprintf("Invalid base64 path: %v", err)))
		return
	}

	content, err := s.ftpService.Download(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, s.newFTPErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, FTPDownloadResponse{
		Content: encoder.EncodeBytes(content),
	})
}

func (s *Server) newFTPErrorResponse(msg string) gin.H {
	return gin.H{
		"error": encoder.Encode(msg),
	}
}
