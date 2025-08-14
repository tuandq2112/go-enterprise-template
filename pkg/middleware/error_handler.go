package middleware

import (
	"context"
	"net/http"

	"go-clean-ddd-es-template/pkg/errors"
	"go-clean-ddd-es-template/pkg/i18n"
	"go-clean-ddd-es-template/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandler handles errors with i18n support
type ErrorHandler struct {
	translator *i18n.Translator
	logger     logger.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(translator *i18n.Translator, logger logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		translator: translator,
		logger:     logger,
	}
}

// HandleError handles an error and returns appropriate response
func (h *ErrorHandler) HandleError(err error, locale string) *ErrorResponse {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	if appErr, ok := err.(*errors.AppError); ok {
		// Translate the error message
		translatedErr := h.translator.TranslateError(appErr, locale)

		// Log the error
		h.logger.Error("Application error occurred", map[string]interface{}{
			"error_code":    appErr.Code,
			"error_message": appErr.Message,
			"locale":        locale,
			"http_status":   appErr.HTTPStatus,
			"details":       appErr.Details,
		})

		return &ErrorResponse{
			Code:    string(translatedErr.Code),
			Message: translatedErr.Message,
			Details: translatedErr.Details,
		}
	}

	// Try to convert to AppError
	if appErr, ok := errors.AsAppError(err); ok {
		translatedErr := h.translator.TranslateError(appErr, locale)

		h.logger.Error("Converted error to AppError", map[string]interface{}{
			"error_code":    appErr.Code,
			"error_message": appErr.Message,
			"locale":        locale,
		})

		return &ErrorResponse{
			Code:    string(translatedErr.Code),
			Message: translatedErr.Message,
			Details: translatedErr.Details,
		}
	}

	// Handle unknown errors
	h.logger.Error("Unknown error occurred", map[string]interface{}{
		"error":  err.Error(),
		"locale": locale,
	})

	return &ErrorResponse{
		Code:    string(errors.ErrInternalServer),
		Message: h.translator.Translate(string(errors.ErrInternalServer), locale),
	}
}

// HTTPErrorHandler returns an HTTP middleware for error handling
func (h *ErrorHandler) HTTPErrorHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract locale from request headers or query params
			locale := h.extractLocale(r)

			// Create a custom response writer to capture errors
			errorWriter := &errorResponseWriter{
				ResponseWriter: w,
				errorHandler:   h,
				locale:         locale,
			}

			next.ServeHTTP(errorWriter, r)
		})
	}
}

// GRPCErrorHandler returns a gRPC interceptor for error handling
func (h *ErrorHandler) GRPCErrorHandler() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract locale from context or metadata
		locale := h.extractLocaleFromContext(ctx)

		resp, err := handler(ctx, req)
		if err != nil {
			// Handle the error
			appErr := h.handleGRPCError(err, locale)
			return nil, appErr
		}

		return resp, nil
	}
}

// handleGRPCError handles gRPC errors
func (h *ErrorHandler) handleGRPCError(err error, locale string) error {
	// Check if it's already a gRPC status
	if st, ok := status.FromError(err); ok {
		// Convert gRPC status to AppError
		appErr := h.convertGRPCStatusToAppError(st, locale)
		return status.Error(codes.Code(appErr.HTTPStatus), appErr.Message)
	}

	// Handle AppError
	if appErr, ok := err.(*errors.AppError); ok {
		translatedErr := h.translator.TranslateError(appErr, locale)
		return status.Error(codes.Code(translatedErr.HTTPStatus), translatedErr.Message)
	}

	// Handle unknown errors
	h.logger.Error("Unknown gRPC error", map[string]interface{}{
		"error":  err.Error(),
		"locale": locale,
	})

	internalErr := h.translator.Translate(string(errors.ErrInternalServer), locale)
	return status.Error(codes.Internal, internalErr)
}

// convertGRPCStatusToAppError converts gRPC status to AppError
func (h *ErrorHandler) convertGRPCStatusToAppError(st *status.Status, locale string) *errors.AppError {
	var code errors.ErrorCode
	var message string

	switch st.Code() {
	case codes.InvalidArgument:
		code = errors.ErrBadRequest
		message = h.translator.Translate(string(errors.ErrBadRequest), locale)
	case codes.NotFound:
		code = errors.ErrNotFound
		message = h.translator.Translate(string(errors.ErrNotFound), locale)
	case codes.PermissionDenied:
		code = errors.ErrForbidden
		message = h.translator.Translate(string(errors.ErrForbidden), locale)
	case codes.Unauthenticated:
		code = errors.ErrUnauthorized
		message = h.translator.Translate(string(errors.ErrUnauthorized), locale)
	case codes.DeadlineExceeded:
		code = errors.ErrTimeout
		message = h.translator.Translate(string(errors.ErrTimeout), locale)
	case codes.Unavailable:
		code = errors.ErrServiceUnavailable
		message = h.translator.Translate(string(errors.ErrServiceUnavailable), locale)
	default:
		code = errors.ErrInternalServer
		message = h.translator.Translate(string(errors.ErrInternalServer), locale)
	}

	return &errors.AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: int(st.Code()),
		Locale:     locale,
	}
}

// extractLocale extracts locale from HTTP request
func (h *ErrorHandler) extractLocale(r *http.Request) string {
	// Try to get locale from Accept-Language header
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		// Parse Accept-Language header (simplified)
		if len(acceptLang) >= 2 {
			locale := acceptLang[:2]
			if h.translator.IsLocaleSupported(locale) {
				return locale
			}
		}
	}

	// Try to get locale from query parameter
	if locale := r.URL.Query().Get("locale"); locale != "" {
		if h.translator.IsLocaleSupported(locale) {
			return locale
		}
	}

	// Default to English
	return "en"
}

// extractLocaleFromContext extracts locale from gRPC context
func (h *ErrorHandler) extractLocaleFromContext(ctx context.Context) string {
	// This is a simplified implementation
	// In a real application, you would extract locale from gRPC metadata
	return "en"
}

// errorResponseWriter is a custom response writer that captures errors
type errorResponseWriter struct {
	http.ResponseWriter
	errorHandler *ErrorHandler
	locale       string
	statusCode   int
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *errorResponseWriter) Write(data []byte) (int, error) {
	// If this is an error response, we might want to transform it
	if w.statusCode >= 400 {
		// You could transform the error response here if needed
	}
	return w.ResponseWriter.Write(data)
}
