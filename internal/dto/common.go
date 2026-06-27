package dto

import "time"

type ErrorResponse struct {
	Status    string `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func NewErrorResponse(code string, message string) ErrorResponse {
	return ErrorResponse{
		Status:    "error",
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

type SuccessResponse struct {
	Status    string `json:"status"`
	Data      any    `json:"data"`
	Timestamp string `json:"timestamp"`
}

func NewSuccessResponse(data any) SuccessResponse {
	return SuccessResponse{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	Limit       int   `json:"limit"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int64 `json:"total_pages"`
}

type PaginatedResponse struct {
	Status     string         `json:"status"`
	Data       any            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Timestamp  string         `json:"timestamp"`
}

func NewPaginatedResponse(data any, pagination PaginationMeta) PaginatedResponse {
	return PaginatedResponse{
		Status:     "success",
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}
