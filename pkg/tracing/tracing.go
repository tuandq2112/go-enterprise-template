package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer represents the tracing service
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new tracer instance
func NewTracer(serviceName, serviceVersion string) (*Tracer, error) {
	// Create OTLP exporter
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint("http://localhost:4318/v1/traces"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := tp.Tracer(serviceName)

	return &Tracer{
		tracer: tracer,
	}, nil
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a new span with attributes
func (t *Tracer) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]interface{}, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	spanOpts := make([]trace.SpanStartOption, 0, len(opts)+len(attrs))
	spanOpts = append(spanOpts, opts...)

	for k, v := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(
			attribute.String(k, fmt.Sprintf("%v", v)),
		))
	}

	return t.tracer.Start(ctx, name, spanOpts...)
}

// AddEvent adds an event to the current span
func (t *Tracer) AddEvent(ctx context.Context, name string, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	spanOpts := make([]trace.EventOption, 0, len(attrs))
	for k, v := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(
			attribute.String(k, fmt.Sprintf("%v", v)),
		))
	}

	span.AddEvent(name, spanOpts...)
}

// SetAttributes sets attributes on the current span
func (t *Tracer) SetAttributes(ctx context.Context, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	for k, v := range attrs {
		span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", v)))
	}
}

// RecordError records an error on the current span
func (t *Tracer) RecordError(ctx context.Context, err error, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	spanOpts := make([]trace.EventOption, 0, len(attrs))
	for k, v := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(
			attribute.String(k, fmt.Sprintf("%v", v)),
		))
	}

	span.RecordError(err, spanOpts...)
}

// Shutdown gracefully shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}

// TraceFunction traces a function execution
func (t *Tracer) TraceFunction(ctx context.Context, functionName string, fn func(context.Context) error) error {
	ctx, span := t.StartSpan(ctx, functionName)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	t.SetAttributes(ctx, map[string]interface{}{
		"function.duration_ms": duration.Milliseconds(),
	})

	if err != nil {
		t.RecordError(ctx, err, nil)
	}

	return err
}

// TraceFunctionWithResult traces a function execution with result
func TraceFunctionWithResult[T any](t *Tracer, ctx context.Context, functionName string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := t.StartSpan(ctx, functionName)
	defer span.End()

	start := time.Now()
	result, err := fn(ctx)
	duration := time.Since(start)

	t.SetAttributes(ctx, map[string]interface{}{
		"function.duration_ms": duration.Milliseconds(),
	})

	if err != nil {
		t.RecordError(ctx, err, nil)
	}

	return result, err
}
