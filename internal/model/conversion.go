package model

import "time"

// MarkdownConversionRequest represents request to convert Markdown to image
type MarkdownConversionRequest struct {
	Content     string            `json:"content" validate:"required"`            // Markdown content to be converted
	Title       string            `json:"title,omitempty"`                        // Optional title for the document
	CSS         string            `json:"css,omitempty"`                          // Optional custom CSS
	Theme       string            `json:"theme,omitempty" default:"light"`        // Theme: light, dark
	ImageFormat string            `json:"image_format,omitempty" default:"png"`   // Image format: png, jpg, webp
	Width       int               `json:"width,omitempty" default:"1200"`         // Width of the output image
	Height      int               `json:"height,omitempty" default:"800"`         // Height of the output image
	Scale       float64           `json:"scale,omitempty" default:"2.0"`          // Scale factor for high-DPI output
	Quality     int               `json:"quality,omitempty" default:"90"`         // Quality percentage for JPEG/WebP
	IncludeMeta bool              `json:"include_meta,omitempty" default:"false"` // Include metadata in the output
	Headers     map[string]string `json:"headers,omitempty"`                      // Additional headers for the rendered document
}

// MarkdownConversionResponse represents response from Markdown to image conversion
type MarkdownConversionResponse struct {
	TaskID      string    `json:"task_id"`              // Unique identifier for the conversion task
	ImageURL    string    `json:"image_url,omitempty"`  // URL to the generated image, if available
	ImageData   []byte    `json:"image_data,omitempty"` // Image binary data
	ImageFormat string    `json:"image_format"`         // Format of the generated image
	Size        ImageSize `json:"size"`                 // Actual dimensions of the image
	Duration    int64     `json:"duration_ms"`          // Processing duration in milliseconds
	Error       string    `json:"error,omitempty"`      // Error message if conversion failed
}

// ImageSize represents dimensions of an image
type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ConversionTask represents a conversion task in the system
type ConversionTask struct {
	TaskID      string                      `json:"task_id"`
	Request     MarkdownConversionRequest   `json:"request"`
	Status      string                      `json:"status"` // pending, processing, completed, failed
	Result      *MarkdownConversionResponse `json:"result,omitempty"`
	CreatedAt   time.Time                   `json:"created_at"`
	ProcessedAt *time.Time                  `json:"processed_at,omitempty"`
	Error       string                      `json:"error,omitempty"`
}
