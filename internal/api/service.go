package api

import (
	"github.com/insmtx/InkForge/internal/model"
)

// Initialize the API service with the rendering engine
func Initialize() error {
	options := model.RenderingOptions{}

	return InitEngine(options)
}
