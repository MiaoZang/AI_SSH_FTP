package server

import (
	"fmt"
	"net/http"

	"ssh-ftp-proxy/internal/config"
	"ssh-ftp-proxy/internal/encoder"
	"ssh-ftp-proxy/internal/logger"
	"ssh-ftp-proxy/internal/service/ftp"
	"ssh-ftp-proxy/internal/service/ssh"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine     *gin.Engine
	sshService *ssh.Service
	ftpService *ftp.Service
}

func NewServer() *Server {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware())

	s := &Server{
		engine:     engine,
		sshService: ssh.NewService(config.GlobalConfig.SSHServer),
		ftpService: ftp.NewService(config.GlobalConfig.FTPServer),
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
