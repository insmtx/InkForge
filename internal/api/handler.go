package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/InkForge/internal/model"
	"github.com/insmtx/InkForge/internal/render"
)

var (
	renderEngine *render.RenderEngine
)

// InitEngine initializes the rendering engine with specified options
func InitEngine(options model.RenderingOptions) error {
	engine, err := render.NewRenderEngine(options)
	if err != nil {
		return err
	}
	renderEngine = engine

	return nil
}

// MarkdownToImageHandler handles the POST /markdown2img endpoint synchronously
func MarkdownToImageHandler(ctx *gin.Context) {
	var req model.MarkdownConversionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(int(model.BadRequestCode), "Invalid JSON format"))
		return
	}

	// Validate the request
	if err := renderEngine.ValidateRequest(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(int(model.ValidationFailedCode), err.Error()))
		return
	}

	// Start timing the conversion
	startTime := time.Now()

	// Process the conversion synchronously
	result, err := renderEngine.RenderMarkdownToImage(context.Background(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse(int(model.ConversionFailedCode), err.Error()))
		return
	}

	duration := time.Since(startTime)

	// Log the processing time for monitoring purposes
	_ = duration // Will be used for logging in a complete implementation

	// Set image headers and return the image directly
	filename := "inkforge-" + startTime.Format("20060102-150405") + "." + result.ImageFormat
	ctx.Header("Content-Disposition", "inline; filename="+filename)
	ctx.Header("Content-Type", "image/"+result.ImageFormat)

	ctx.Data(http.StatusOK, "image/"+result.ImageFormat, result.ImageData)
}

// HealthCheckHandler handles the GET /health endpoint
func HealthCheckHandler(ctx *gin.Context) {
	response := model.HealthResponse{
		Service:   "InkForge Markdown Renderer",
		Version:   "1.0.0",
		Status:    "healthy",
		Timestamp: time.Now(),
		Stats: map[string]interface{}{
			"uptime": time.Since(time.Unix(0, 0)).Milliseconds(),
		},
	}

	ctx.JSON(http.StatusOK, model.SuccessResponse(response))
}
