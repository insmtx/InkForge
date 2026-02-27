package model

import "time"

// BaseResponse represents a generic API response structure
type BaseResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// TaskStatusResponse provides information about a task status
type TaskStatusResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`   // pending, processing, completed, failed
	Progress  int    `json:"progress"` // 0-100 percentage
	Error     string `json:"error,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// HealthResponse provides health check information
type HealthResponse struct {
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Status    string                 `json:"status"` // healthy, degraded, unhealthy
	Uptime    int64                  `json:"uptime"` // milliseconds since service started
	Stats     map[string]interface{} `json:"stats,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SuccessResponse creates a successful response
func SuccessResponse(data interface{}) *BaseResponse {
	return &BaseResponse{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ErrorResponse creates an error response
func ErrorResponse(code int, message string) *BaseResponse {
	return &BaseResponse{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}
