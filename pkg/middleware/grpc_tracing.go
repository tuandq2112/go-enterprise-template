package middleware

import (
	"context"
	"time"

	"go-clean-ddd-es-template/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCTracingInterceptor creates a gRPC interceptor that adds tracing
func GRPCTracingInterceptor(tracer *tracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if tracer == nil {
			return handler(ctx, req)
		}

		// Extract metadata
		md, _ := metadata.FromIncomingContext(ctx)

		// Start span
		ctx, span := tracer.StartSpan(ctx, info.FullMethod,
			trace.WithAttributes(
				attribute.String("grpc.method", info.FullMethod),
				attribute.String("grpc.type", "unary"),
			),
		)
		defer span.End()

		// Add metadata attributes
		if userAgent := getMetadataValue(md, "user-agent"); userAgent != "" {
			span.SetAttributes(attribute.String("grpc.user_agent", userAgent))
		}
		if requestID := getMetadataValue(md, "x-request-id"); requestID != "" {
			span.SetAttributes(attribute.String("request.id", requestID))
		}

		// Record start time
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Record end time and duration
		duration := time.Since(start)
		span.SetAttributes(attribute.Int64("grpc.duration_ms", duration.Milliseconds()))

		// Record status
		if err != nil {
			st := status.Convert(err)
			span.SetAttributes(
				attribute.String("grpc.status_code", st.Code().String()),
				attribute.String("grpc.status_message", st.Message()),
			)
			span.RecordError(err)
		} else {
			span.SetAttributes(attribute.String("grpc.status_code", "OK"))
		}

		// Add events
		span.AddEvent("grpc.request.start")
		span.AddEvent("grpc.request.end", trace.WithAttributes(
			attribute.Int64("duration_ms", duration.Milliseconds()),
		))

		return resp, err
	}
}

// GRPCStreamTracingInterceptor creates a gRPC stream interceptor that adds tracing
func GRPCStreamTracingInterceptor(tracer *tracing.Tracer) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if tracer == nil {
			return handler(srv, ss)
		}

		// Extract metadata
		md, _ := metadata.FromIncomingContext(ss.Context())

		// Start span
		ctx, span := tracer.StartSpan(ss.Context(), info.FullMethod,
			trace.WithAttributes(
				attribute.String("grpc.method", info.FullMethod),
				attribute.String("grpc.type", "stream"),
				attribute.Bool("grpc.is_client_stream", info.IsClientStream),
				attribute.Bool("grpc.is_server_stream", info.IsServerStream),
			),
		)
		defer span.End()

		// Add metadata attributes
		if userAgent := getMetadataValue(md, "user-agent"); userAgent != "" {
			span.SetAttributes(attribute.String("grpc.user_agent", userAgent))
		}
		if requestID := getMetadataValue(md, "x-request-id"); requestID != "" {
			span.SetAttributes(attribute.String("request.id", requestID))
		}

		// Record start time
		start := time.Now()

		// Create wrapped stream
		wrappedStream := &tracingServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Call handler
		err := handler(srv, wrappedStream)

		// Record end time and duration
		duration := time.Since(start)
		span.SetAttributes(attribute.Int64("grpc.duration_ms", duration.Milliseconds()))

		// Record status
		if err != nil {
			st := status.Convert(err)
			span.SetAttributes(
				attribute.String("grpc.status_code", st.Code().String()),
				attribute.String("grpc.status_message", st.Message()),
			)
			span.RecordError(err)
		} else {
			span.SetAttributes(attribute.String("grpc.status_code", "OK"))
		}

		// Add events
		span.AddEvent("grpc.stream.start")
		span.AddEvent("grpc.stream.end", trace.WithAttributes(
			attribute.Int64("duration_ms", duration.Milliseconds()),
		))

		return err
	}
}

// tracingServerStream wraps grpc.ServerStream to provide tracing context
type tracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *tracingServerStream) Context() context.Context {
	return s.ctx
}

// getMetadataValue extracts a value from metadata
func getMetadataValue(md metadata.MD, key string) string {
	if values := md.Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}
