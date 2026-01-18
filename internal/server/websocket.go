package server

import (
	"fmt"
	"net/http"

	"ssh-ftp-proxy/internal/config"
	"ssh-ftp-proxy/internal/logger"
	"ssh-ftp-proxy/internal/service/ssh"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	engine     *gin.Engine
	sshService *ssh.Service
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func NewWSServer() *WSServer {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware())

	s := &WSServer{
		engine:     engine,
		sshService: ssh.NewService(config.GlobalConfig.SSHServer),
	}

	s.setupRoutes()
	return s
}

func (s *WSServer) setupRoutes() {
	s.engine.GET("/ws/ssh", s.handleSSHInteractive)
}

func (s *WSServer) Run() error {
	addr := fmt.Sprintf("%s:%d", config.GlobalConfig.Server.BindIP, config.GlobalConfig.Server.WSPort)
	logger.Log.Info("Starting WebSocket Server", "addr", addr)
	return s.engine.Run(addr)
}

func (s *WSServer) handleSSHInteractive(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade websocket", "error", err)
		return
	}
	defer conn.Close()

	if err := s.sshService.StartInteractive(conn); err != nil {
		logger.Log.Error("SSH Interactive session failed", "error", err)
		conn.WriteJSON(gin.H{"type": "error", "payload": err.Error()})
	}
}
