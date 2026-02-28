package main

import (
	"fmt"
	"log"
	"os"

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
	// Only attempt to install Playwright if needed (e.g., in development)
	// In production environment (container), browsers should already be pre-installed
	if _, err := os.Stat("/root/.cache/ms-playwright-go"); os.IsNotExist(err) {
		// Cache directory doesn't exist, attempt installation
		err := playwright.Install()
		if err != nil {
			log.Printf("Warning: Could not install Playwright: %v. Browsers may need to be pre-installed in the container.", err)
		}
	} else {
		// Browsers are expected to be pre-installed in the base image
		fmt.Println("Using pre-installed Playwright browsers from cache")
	}

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
