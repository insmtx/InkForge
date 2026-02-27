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
	// Read the base template file content
	templateDir := filepath.Join("internal", "render", "template")
	baseTemplatePath := filepath.Join(templateDir, "base.html")

	content, err := ioutil.ReadFile(baseTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read base template: %w", err)
	}

	tmpl, err := template.New("base").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	e.templates["base"] = tmpl
	return nil
}

// RenderMarkdownToImage converts Markdown content to an image
func (e *RenderEngine) RenderMarkdownToImage(ctx context.Context, req *model.MarkdownConversionRequest) (*model.MarkdownConversionResponse, error) {
	startTime := time.Now()

	// Convert Markdown to HTML
	htmlContent, err := e.convertMarkdownToHTML(req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Prepare template data
	templateData := map[string]interface{}{
		"Title":          req.Title,
		"Content":        template.HTML(htmlContent),
		"CSS":            req.CSS,
		"Theme":          req.Theme,
		"KaTeXEnabled":   e.options.EnableKaTeX,
		"MermaidEnabled": e.options.EnableMermaid,
	}

	var buf bytes.Buffer
	tmpl, exists := e.templates["base"]
	if !exists {
		return nil, fmt.Errorf("base template not found")
	}

	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Process custom CSS/JS if provided in request
	finalHTML := buf.String()
	if len(req.Headers) > 0 {
		// Add custom headers if needed - implementation would go here
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

	return response, nil
}

// convertMarkdownToHTML converts Markdown text to HTML string
func (e *RenderEngine) convertMarkdownToHTML(markdownText string) (string, error) {
	// Create a parser with extended options
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.FencedCode |
		parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// Create a renderer with HTML options
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags |
			html.HrefTargetBlank |
			html.FootnoteReturnLinks,
	})

	// Parse and render the markdown
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
