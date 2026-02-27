package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/InkForge/internal/api"
	"github.com/insmtx/InkForge/internal/model"
	"github.com/playwright-community/playwright-go"
)

var (
	port string
	host string
)

func StartServer() {
	err := playwright.Install()
	if err != nil {
		panic(err)
	}

	config := GetServerConfig()

	// Init API service
	err = api.InitEngine(model.RenderingOptions{
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
