package model

// ErrorCode represents standard error codes for InkForge API
type ErrorCode int

const (
	// Generic errors
	SuccessCode            ErrorCode = 0
	InternalErrorCode      ErrorCode = 1000
	BadRequestCode         ErrorCode = 1001
	UnauthorizedCode       ErrorCode = 1002
	ForbiddenCode          ErrorCode = 1003
	NotFoundCode           ErrorCode = 1004
	ConflictCode           ErrorCode = 1005
	ValidationFailedCode   ErrorCode = 1006
	RequestTooLargeCode    ErrorCode = 1007
	TimeoutCode            ErrorCode = 1008
	ServiceUnavailableCode ErrorCode = 1009

	// Conversion-specific errors
	ConversionFailedCode  ErrorCode = 2001
	UnsupportedFormatCode ErrorCode = 2002
	RenderEngineError     ErrorCode = 2003
	InvalidMarkdownSyntax ErrorCode = 2004
	InvalidCSSCode        ErrorCode = 2005
	LaTeXProcessingError  ErrorCode = 2006
	MermaidRenderingError ErrorCode = 2007
	ImageGenerationError  ErrorCode = 2008
	UnsupportedThemeCode  ErrorCode = 2009
	URLFetchingError      ErrorCode = 2010
	ResourceLimitExceeded ErrorCode = 2011
)

// ErrorDetail provides detailed error information for validation errors
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// APIError represents a comprehensive API error response
type APIError struct {
	Code      ErrorCode     `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
	Timestamp string        `json:"timestamp"`
}
