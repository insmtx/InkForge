package api

import (
	"github.com/gin-gonic/gin"
)

func SetRouter(eng *gin.Engine) {
	// Serve static files for the demo interface from a subroute to avoid conflicts
	eng.Static("/static", "./demo")
	eng.StaticFS("/demo", gin.Dir("./demo", false))

	// Root path serves index.html from demo directory
	eng.GET("/", func(c *gin.Context) {
		c.File("./demo/index.html")
	})

	v1 := eng.Group("/api/v1")
	{
		v1.POST("/markdown2image", MarkdownToImageHandler)
		v1.GET("/health", HealthCheckHandler)
	}

	// For backward compatibility
	eng.POST("/markdown2img", MarkdownToImageHandler)
	eng.GET("/health", HealthCheckHandler)
}
