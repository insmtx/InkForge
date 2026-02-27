package api

import (
	"github.com/gin-gonic/gin"
)

func SetRouter(eng *gin.Engine) {
	v1 := eng.Group("/api/v1")
	{
		v1.POST("/markdown2image", MarkdownToImageHandler)
		v1.GET("/health", HealthCheckHandler)
	}

	// For backward compatibility
	eng.POST("/markdown2img", MarkdownToImageHandler)
	eng.GET("/health", HealthCheckHandler)
}
