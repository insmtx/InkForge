package render

import "context"

// Renderer defines the interface for a renderer (Playwright or other)
type Renderer interface {
	RenderHTMLToImage(ctx context.Context, html string, width, height int, scale float64, imgFormat string) ([]byte, error)
	Close() error
}
