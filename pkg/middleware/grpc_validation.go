package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go-clean-ddd-es-template/pkg/errors"
)

// GRPCValidationInterceptor creates a gRPC unary interceptor for validation
func GRPCValidationInterceptor(validationMiddleware *ValidationMiddleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Validate metadata (headers)
		if err := validateGRPCMetadata(ctx, validationMiddleware); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "metadata validation failed: %v", err)
		}

		// Validate request payload
		if err := validateGRPCRequest(req, validationMiddleware); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "request validation failed: %v", err)
		}

		// Continue to handler
		return handler(ctx, req)
	}
}

// GRPCStreamValidationInterceptor creates a gRPC stream interceptor for validation
func GRPCStreamValidationInterceptor(validationMiddleware *ValidationMiddleware) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Validate metadata (headers)
		if err := validateGRPCMetadata(stream.Context(), validationMiddleware); err != nil {
			return status.Errorf(codes.InvalidArgument, "metadata validation failed: %v", err)
		}

		// Wrap stream to validate messages
		wrappedStream := &validatedServerStream{
			ServerStream:         stream,
			validationMiddleware: validationMiddleware,
		}

		// Continue to handler
		return handler(srv, wrappedStream)
	}
}

// validatedServerStream wraps grpc.ServerStream to validate messages
type validatedServerStream struct {
	grpc.ServerStream
	validationMiddleware *ValidationMiddleware
}

// RecvMsg validates incoming messages
func (s *validatedServerStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	// Validate message
	if err := validateGRPCRequest(m, s.validationMiddleware); err != nil {
		return status.Errorf(codes.InvalidArgument, "message validation failed: %v", err)
	}

	return nil
}

// validateGRPCMetadata validates gRPC metadata (headers)
func validateGRPCMetadata(ctx context.Context, vm *ValidationMiddleware) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil // No metadata to validate
	}

	// Check metadata size
	var metadataSize int64
	for key, values := range md {
		metadataSize += int64(len(key))
		for _, value := range values {
			metadataSize += int64(len(value))
		}
	}

	if metadataSize > vm.config.MaxHeaderSize {
		return errors.New(errors.ErrBadRequest, "Metadata too large")
	}

	// Check for suspicious metadata keys
	suspiciousKeys := []string{
		"x-forwarded-for", "x-real-ip", "x-forwarded-host",
		"x-forwarded-proto", "x-forwarded-port",
	}

	for key := range md {
		keyLower := strings.ToLower(key)
		for _, suspicious := range suspiciousKeys {
			if keyLower == suspicious {
				vm.logger.Warn("Suspicious metadata key detected: %s", key)
			}
		}
	}

	return nil
}

// validateGRPCRequest validates gRPC request payload
func validateGRPCRequest(req interface{}, vm *ValidationMiddleware) error {
	if req == nil {
		return nil
	}

	// Convert request to string for pattern checking
	reqStr := requestToString(req)
	if reqStr == "" {
		return nil
	}

	// Check for blocked patterns
	if vm.containsBlockedPatterns(reqStr) {
		return errors.New(errors.ErrBadRequest, "Request contains blocked patterns")
	}

	// Check for null bytes
	if strings.Contains(reqStr, "\x00") {
		return errors.New(errors.ErrBadRequest, "Request contains null bytes")
	}

	// Check request size (approximate)
	if int64(len(reqStr)) > vm.config.MaxRequestSize {
		return errors.New(errors.ErrBadRequest, "Request too large")
	}

	return nil
}

// requestToString converts gRPC request to string for validation
func requestToString(req interface{}) string {
	// Try to get string representation of the request
	switch v := req.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case io.Reader:
		if data, err := io.ReadAll(v); err == nil {
			return string(data)
		}
	default:
		// For protobuf messages, try to get their string representation
		if stringer, ok := req.(interface{ String() string }); ok {
			return stringer.String()
		}
		// Fallback to fmt.Sprintf
		result := fmt.Sprintf("%+v", req)
		result = strings.ReplaceAll(result, "\n", " ")
		result = strings.ReplaceAll(result, "\t", " ")
		result = strings.ReplaceAll(result, "  ", " ")
		return strings.TrimSpace(result)
	}
	return ""
}

// GRPCRateLimitInterceptor creates a gRPC interceptor for rate limiting
func GRPCRateLimitInterceptor(validationMiddleware *ValidationMiddleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Create a mock HTTP request for rate limiting
		mockReq := &http.Request{
			Method: "POST", // gRPC requests are typically POST
			Header: make(http.Header),
		}

		// Extract client IP from gRPC metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if forwardedFor := md.Get("x-forwarded-for"); len(forwardedFor) > 0 {
				mockReq.Header.Set("X-Forwarded-For", forwardedFor[0])
			}
			if realIP := md.Get("x-real-ip"); len(realIP) > 0 {
				mockReq.Header.Set("X-Real-IP", realIP[0])
			}
		}

		// Check rate limit
		if err := validationMiddleware.checkRateLimit(mockReq); err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
		}

		// Continue to handler
		return handler(ctx, req)
	}
}

// GRPCStreamRateLimitInterceptor creates a gRPC stream interceptor for rate limiting
func GRPCStreamRateLimitInterceptor(validationMiddleware *ValidationMiddleware) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Create a mock HTTP request for rate limiting
		mockReq := &http.Request{
			Method: "POST", // gRPC requests are typically POST
			Header: make(http.Header),
		}

		// Extract client IP from gRPC metadata
		if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
			if forwardedFor := md.Get("x-forwarded-for"); len(forwardedFor) > 0 {
				mockReq.Header.Set("X-Forwarded-For", forwardedFor[0])
			}
			if realIP := md.Get("x-real-ip"); len(realIP) > 0 {
				mockReq.Header.Set("X-Real-IP", realIP[0])
			}
		}

		// Check rate limit
		if err := validationMiddleware.checkRateLimit(mockReq); err != nil {
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
		}

		// Continue to handler
		return handler(srv, stream)
	}
}
