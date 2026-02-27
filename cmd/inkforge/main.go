package main

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/InkForge/internal/model"
)

var (
	portVar string
	hostVar string

	rootCmd = &cobra.Command{
		Use:   "inkforge",
		Short: "InkForge is a high-performance Markdown rendering engine.",
		Long: `InkForge is a high-performance Markdown rendering engine designed for developers and AI systems. 
It converts Markdown content into high-quality image outputs with full support for standard Markdown 
syntax, LaTeX mathematical expressions (powered by KaTeX), and Mermaid diagrams including flowcharts 
and sequence diagrams.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Start the API server
			StartServer()
		},
	}
)

func main() {
	rootCmd.Flags().StringVarP(&portVar, "port", "p", "8080", "Port to run the server on")
	rootCmd.Flags().StringVarP(&hostVar, "host", "", "0.0.0.0", "Host to run the server on")

	if err := rootCmd.Execute(); err != nil {
		logs.Errorf("Error executing command: %v", err)
	}
}

func GetServerConfig() *model.ServerConfig {
	// Parse port to integer
	portNum, _ := strconv.Atoi(portVar)

	config := &model.ServerConfig{
		Port:             portNum,
		Host:             hostVar,
		ReadTimeout:      30,      // 30 seconds
		WriteTimeout:     60,      // 60 seconds
		IdleTimeout:      120,     // 120 seconds
		MaxHeaderSize:    1 << 20, // 1 MB
		MaxImageWidth:    4000,
		MaxImageHeight:   4000,
		DefaultScale:     2.0,
		MaxScale:         4.0,
		MaxContentLength: 1024 * 1024, // 1 MB
		KaTeXEnabled:     true,
		MermaidEnabled:   true,
		StoragePath:      "./storage",
		PublicURL:        "http://" + hostVar + ":" + portVar,
		MaxWorkers:       10,
		QueueSize:        100,
	}

	return config
}
