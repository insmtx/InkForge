package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/InkForge/internal/model"
)

func SetRouter(eng *gin.Engine) {
	// Add route for root path to serve index.html
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

	// Additional utilities for debugging
	eng.POST("/api/v1/generatehtml", GenerateHTMLHandler)
}

// GenerateHTMLHandler creates rendered HTML for debugging purposes
func GenerateHTMLHandler(ctx *gin.Context) {
	var req model.MarkdownConversionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(int(model.BadRequestCode), "Invalid JSON format"))
		return
	}

	// Validate the request
	if err := renderEngine.ValidateRequest(&req); err != nil {
		logs.Errorf("Request validation failed: %v", err)
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(int(model.ValidationFailedCode), err.Error()))
		return
	}

	// Generate HTML content using the rendering engine
	htmlContent, err := renderEngine.RenderMarkdownAsHTML(&req)
	if err != nil {
		logs.Errorf("HTML rendering failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse(int(model.ConversionFailedCode), err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"html":    htmlContent,
		"title":   req.Title,
		"message": "Generated HTML content for debugging",
	})
}
