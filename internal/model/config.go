package model

// ServerConfig holds server configuration values
type ServerConfig struct {
	Port          int    `json:"port" yaml:"port" env:"PORT"`
	Host          string `json:"host" yaml:"host" env:"HOST"`
	ReadTimeout   int    `json:"read_timeout" yaml:"read_timeout"`   // in seconds
	WriteTimeout  int    `json:"write_timeout" yaml:"write_timeout"` // in seconds
	IdleTimeout   int    `json:"idle_timeout" yaml:"idle_timeout"`   // in seconds
	MaxHeaderSize int    `json:"max_header_size" yaml:"max_header_size"`

	// Image generation settings
	MaxImageWidth    int     `json:"max_image_width" yaml:"max_image_width"`
	MaxImageHeight   int     `json:"max_image_height" yaml:"max_image_height"`
	DefaultScale     float64 `json:"default_scale" yaml:"default_scale"`
	MaxScale         float64 `json:"max_scale" yaml:"max_scale"`
	MaxContentLength int     `json:"max_content_length" yaml:"max_content_length"` // in bytes

	// Rendering settings
	KaTeXEnabled   bool `json:"katex_enabled" yaml:"katex_enabled"`
	MermaidEnabled bool `json:"mermaid_enabled" yaml:"mermaid_enabled"`

	// Storage settings
	StoragePath string `json:"storage_path" yaml:"storage_path"`
	PublicURL   string `json:"public_url" yaml:"public_url"`

	// Concurrency settings
	MaxWorkers int `json:"max_workers" yaml:"max_workers"`
	QueueSize  int `json:"queue_size" yaml:"queue_size"`
}

// RenderingOptions defines options for Markdown rendering
type RenderingOptions struct {
	EnableSyntaxHighlighting bool              `json:"enable_syntax_highlighting"`
	EnableKaTeX              bool              `json:"enable_katex"`
	EnableMermaid            bool              `json:"enable_mermaid"`
	DisableExternalResources bool              `json:"disable_external_resources"`
	CustomJS                 []string          `json:"custom_js,omitempty"`
	CustomCSS                []string          `json:"custom_css,omitempty"`
	ResourceHints            map[string]string `json:"resource_hints,omitempty"`
}
