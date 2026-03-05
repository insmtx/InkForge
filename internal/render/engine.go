package render

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/InkForge/internal/model"
)

// RenderEngine manages the complete rendering process from Markdown to image
type RenderEngine struct {
	renderer  *PlaywrightRenderer
	templates map[string]*template.Template
	mutex     sync.RWMutex
	options   model.RenderingOptions
}

// NewRenderEngine creates a new render engine instance
func NewRenderEngine(options model.RenderingOptions) (*RenderEngine, error) {
	renderer, err := NewPlaywrightRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Playwright renderer: %w", err)
	}

	engine := &RenderEngine{
		renderer:  renderer,
		templates: make(map[string]*template.Template),
		options:   options,
	}

	// Load the base template
	if err := engine.loadTemplates(); err != nil {
		renderer.Close()
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return engine, nil
}

// loadTemplates loads HTML templates needed for rendering
func (e *RenderEngine) loadTemplates() error {
	// Read the base template file content - find in multiple possible locations
	templatePaths := []string{
		filepath.Join("template", "base.html"),                       // Original location
		filepath.Join("internal", "render", "template", "base.html"), // Deployed location
		filepath.Join(".", "template", "base.html"),                  // Fallback
		"./template/base.html",                                       // Direct path
	}

	var content []byte
	var baseTemplatePath string
	var err error

	// Find the first path that exists
	for _, path := range templatePaths {
		content, err = ioutil.ReadFile(path)
		if err == nil {
			baseTemplatePath = path
			break
		}
		logs.Debugf("Trying template path %s, failed: %v", path, err)
	}

	if err != nil {
		return fmt.Errorf("failed to find base template in any expected location: %w", err)
	}

	logs.Infof("Loading template from path: %s", baseTemplatePath)
	logs.Infof("Successfully read base template, size: %d bytes", len(content))

	tmpl, err := template.New("base").Parse(string(content))
	if err != nil {
		logs.Errorf("Failed to parse base template: %v", err)
		return fmt.Errorf("failed to parse base template: %w", err)
	}
	logs.Info("Successfully parsed base template")

	e.templates["base"] = tmpl
	logs.Info("Template successfully loaded into engine cache")
	return nil
}

// RenderMarkdownToImage converts Markdown content to an image
func (e *RenderEngine) RenderMarkdownToImage(ctx context.Context, req *model.MarkdownConversionRequest) (*model.MarkdownConversionResponse, error) {
	startTime := time.Now()

	// Generate the HTML content
	finalHTML, err := e.RenderMarkdownAsHTML(req)
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to prepare HTML for image rendering: %v", err)
		return nil, fmt.Errorf("failed to prepare HTML: %w", err)
	}

	width := req.Width
	if width == 0 {
		width = 1200
	}
	height := req.Height
	if height == 0 {
		height = 800
	}
	scale := req.Scale
	if scale == 0 {
		scale = 2.0
	}
	imgFormat := req.ImageFormat
	if imgFormat == "" {
		imgFormat = "png"
	}

	// Generate the image using Playwright
	imageData, err := e.renderer.RenderImage(ctx, finalHTML, width, height, scale, imgFormat)
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to render image after preparing parameters: %v", err)
		return nil, fmt.Errorf("failed to render image: %w", err)
	}

	duration := time.Since(startTime)

	response := &model.MarkdownConversionResponse{
		ImageData:   imageData,
		ImageFormat: imgFormat,
		Size: model.ImageSize{
			Width:  width,
			Height: height,
		},
		Duration: duration.Milliseconds(),
	}

	logs.InfoContextf(ctx, "Successfully completed conversion in %v", duration)
	return response, nil
}

// RenderMarkdownAsHTML converts Markdown content to raw HTML string (for debugging)
func (e *RenderEngine) RenderMarkdownAsHTML(req *model.MarkdownConversionRequest) (string, error) {
	// Convert Markdown to HTML
	htmlContent, err := e.convertMarkdownToHTML(req.Content)
	if err != nil {
		logs.Errorf("Failed to convert markdown to HTML: %v", err)
		return "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Prepare template data - all features always enabled
	templateData := map[string]interface{}{
		"Title":          req.Title,
		"Content":        template.HTML(htmlContent),
		"CSS":            req.CSS,
		"Theme":          req.Theme,
		"KaTeXEnabled":   true, // Always enabled
		"MermaidEnabled": true, // Always enabled
	}

	var buf bytes.Buffer
	tmpl, exists := e.templates["base"]
	if !exists {
		logs.Errorf("Base template not found in cache")
		return "", fmt.Errorf("base template not found")
	}

	logs.Infof("Executing template for HTML generation with title: %s", req.Title)
	if err := tmpl.Execute(&buf, templateData); err != nil {
		logs.Errorf("Failed to execute template: %v", err)
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Process custom CSS/JS if provided in request
	finalHTML := buf.String()
	if len(req.Headers) > 0 {
		// Add custom headers if needed - implementation would go here
		logs.Infof("Applying %d custom headers", len(req.Headers))
	}

	return finalHTML, nil
}

// convertMarkdownToHTML converts Markdown text to HTML string while preserving math expressions
func (e *RenderEngine) convertMarkdownToHTML(markdownText string) (string, error) {
	// Parse Markdown to HTML using gomarkdown library
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.FencedCode |
		parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags |
			html.HrefTargetBlank |
			html.FootnoteReturnLinks,
	})

	doc := markdown.Parse([]byte(markdownText), p)
	htmlContent := string(markdown.Render(doc, renderer))

	return htmlContent, nil
}

// Close closes the render engine and cleans up resources
func (e *RenderEngine) Close() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.renderer != nil {
		return e.renderer.Close()
	}
	return nil
}

// ValidateRequest validates the conversion request
func (e *RenderEngine) ValidateRequest(req *model.MarkdownConversionRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return fmt.Errorf("markdown content cannot be empty")
	}

	if req.Width < 100 || req.Width > 10000 {
		return fmt.Errorf("width must be between 100 and 10000 pixels")
	}

	if req.Height < 100 || req.Height > 10000 {
		return fmt.Errorf("height must be between 100 and 10000 pixels")
	}

	imgFormat := strings.ToLower(req.ImageFormat)
	if imgFormat != "" && imgFormat != "png" && imgFormat != "jpeg" && imgFormat != "jpg" && imgFormat != "webp" {
		return fmt.Errorf("image format must be one of: png, jpeg, jpg, webp")
	}

	theme := strings.ToLower(req.Theme)
	if theme != "" && theme != "light" && theme != "dark" {
		return fmt.Errorf("theme must be one of: light, dark")
	}

	return nil
}

// ProcessTask processes a conversion task and updates its status
func (e *RenderEngine) ProcessTask(ctx context.Context, task *model.ConversionTask) error {
	task.Status = "processing"
	task.ProcessedAt = &time.Time{}

	// Perform the actual rendering
	result, err := e.RenderMarkdownToImage(ctx, &task.Request)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		return err
	}

	task.Status = "completed"
	task.Result = result
	now := time.Now()
	task.ProcessedAt = &now

	return nil
}
