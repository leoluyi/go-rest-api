package errors

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	pkglog "github.com/leoluyi/go-api-template/pkg/log"
)

// ErrorResponse is the response that represents an error.
type ErrorResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id,omitempty"`
	Details   interface{} `json:"details,omitempty"`
}

// Error is required by the error interface.
func (e ErrorResponse) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code.
func (e ErrorResponse) StatusCode() int {
	return e.Status
}

// InternalServerError creates a new error response representing an internal server error (HTTP 500)
func InternalServerError(msg string) ErrorResponse {
	if msg == "" {
		msg = "We encountered an error while processing your request."
	}
	return ErrorResponse{
		Status:  http.StatusInternalServerError,
		Message: msg,
	}
}

// NotFound creates a new error response representing a resource-not-found error (HTTP 404)
func NotFound(msg string) ErrorResponse {
	if msg == "" {
		msg = "The requested resource was not found."
	}
	return ErrorResponse{
		Status:  http.StatusNotFound,
		Message: msg,
	}
}

// Unauthorized creates a new error response representing an authentication/authorization failure (HTTP 401)
func Unauthorized(msg string) ErrorResponse {
	if msg == "" {
		msg = "You are not authenticated to perform the requested action."
	}
	return ErrorResponse{
		Status:  http.StatusUnauthorized,
		Message: msg,
	}
}

// Forbidden creates a new error response representing an authorization failure (HTTP 403)
func Forbidden(msg string) ErrorResponse {
	if msg == "" {
		msg = "You are not authorized to perform the requested action."
	}
	return ErrorResponse{
		Status:  http.StatusForbidden,
		Message: msg,
	}
}

// BadRequest creates a new error response representing a bad request (HTTP 400)
func BadRequest(msg string) ErrorResponse {
	if msg == "" {
		msg = "Your request is in a bad format."
	}
	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Message: msg,
	}
}

type invalidField struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// InvalidInput creates a new error response representing a data validation error (HTTP 400).
func InvalidInput(errs validator.ValidationErrors) ErrorResponse {
	var details []invalidField
	for _, e := range errs {
		details = append(details, invalidField{
			Field: e.Field(),
			Error: e.Tag(),
		})
	}
	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Message: "There is some problem with the data you submitted.",
		Details: details,
	}
}

// RespondJSON writes a JSON-encoded response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("errors: failed to encode JSON response: %v", err)
	}
}

// RespondWithError writes a JSON error response derived from the given error,
// including the request ID from the request context for client-side correlation.
func RespondWithError(w http.ResponseWriter, r *http.Request, err error) {
	res := buildErrorResponse(err)
	res.RequestID = pkglog.GetRequestID(r.Context())
	RespondJSON(w, res.Status, res)
}
