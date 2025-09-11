package dtos

/*
Response ErrorResponseDTO dto
Standard error response format for HTTP endpoints
*/
type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// NewErrorResponse creates a new ErrorResponseDTO
func NewErrorResponse(code int, error string, message string) *ErrorResponseDTO {
	return &ErrorResponseDTO{
		Code:    code,
		Error:   error,
		Message: message,
	}
}

/*
Response ValidationErrorResponseDTO dto
Validation error response with field-specific error details
*/
type ValidationErrorResponseDTO struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// NewValidationErrorResponse creates a new ValidationErrorResponseDTO
func NewValidationErrorResponse(message string, fields map[string]any) *ValidationErrorResponseDTO {
	return &ValidationErrorResponseDTO{
		Code:    400,
		Error:   "validation_error",
		Message: message,
		Fields:  fields,
	}
}