package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/InkForge/internal/api"
	"github.com/insmtx/InkForge/internal/model"
	"github.com/ygpkg/yg-go/logs"
)

func StartServer() {
	// Browsers are expected to be pre-installed in the base image via the install command
	// If the cache directory is not present, this indicates a misconfiguration of the environment
	logs.Infof("Using pre-installed Playwright browsers from cache")

	config := GetServerConfig()

	// Init API service
	err := api.InitEngine(model.RenderingOptions{
		EnableKaTeX:              config.KaTeXEnabled,
		EnableMermaid:            config.MermaidEnabled,
		EnableSyntaxHighlighting: true,
		DisableExternalResources: false,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize rendering engine: %v", err))
	}

	eng := gin.Default()
	api.SetRouter(eng)

	server, err := api.NewServer(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create server: %v", err))
	}

	// Start the server
	err = server.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
