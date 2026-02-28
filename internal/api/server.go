package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/insmtx/InkForge/internal/model"
	"github.com/ygpkg/yg-go/logs"
)

// Server wraps the Gin engine with additional configuration
type Server struct {
	Engine *gin.Engine
	Config *model.ServerConfig
}

// NewServer creates a new API server instance
func NewServer(config *model.ServerConfig) (*Server, error) {
	// Set Gin mode based on config or default
	if config == nil {
		config = &model.ServerConfig{
			Port: 8080,
			Host: "localhost",
		}
	}

	gin.SetMode(gin.ReleaseMode) // Change to gin.DebugMode for debugging
	engine := gin.New()

	// Add logging and recovery middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// Configure CORS
	configCORS := cors.DefaultConfig()
	configCORS.AllowAllOrigins = true
	configCORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	configCORS.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	engine.Use(cors.New(configCORS))

	// Initialize the API service
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize API service: %w", err)
	}

	// Set up routes
	SetRouter(engine)

	server := &Server{
		Engine: engine,
		Config: config,
	}

	return server, nil
}

// Start begins serving HTTP requests
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.Engine,
		ReadTimeout:  time.Duration(s.Config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.IdleTimeout) * time.Second,
	}

	logs.Infof("Starting InkForge server on %s", addr)
	logs.Infof("Max content length: %d bytes", s.Config.MaxContentLength)
	logs.Infof("Max image dimensions: %dx%d", s.Config.MaxImageWidth, s.Config.MaxImageHeight)

	return server.ListenAndServe()
}

// StartTLS begins serving HTTPS requests
func (s *Server) StartTLS(certFile, keyFile string) error {
	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.Engine,
		ReadTimeout:  time.Duration(s.Config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.IdleTimeout) * time.Second,
	}

	logs.Infof("Starting HTTPS InkForge server on %s", addr)

	return server.ListenAndServeTLS(certFile, keyFile)
}
