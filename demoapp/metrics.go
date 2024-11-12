package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"time"
)

func NewMeter(ctx context.Context, name string) (metric.Meter, error) {
	provider, err := newMeterProvider(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("could not create meter provider: %w", err)
	}

	return provider.Meter(name), nil
}

func newMeterProvider(ctx context.Context, serviceName string) (metric.MeterProvider, error) {
	interval := 10 * time.Second

	collectorExporter, _ := getOtelMetricsCollectorExporter(ctx)

	periodicReader := metricsdk.NewPeriodicReader(collectorExporter,
		metricsdk.WithInterval(interval),
	)

	provider := metricsdk.NewMeterProvider(
		metricsdk.WithResource(resource.NewSchemaless(
			[]attribute.KeyValue{
				attribute.String(string(semconv.ServiceNameKey), serviceName),
			}...,
		)),
		metricsdk.WithReader(periodicReader),
	)

	return provider, nil
}

func getOtelMetricsCollectorExporter(ctx context.Context) (metricsdk.Exporter, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint("0.0.0.0:4317"),
		otlpmetricgrpc.WithCompressor("gzip"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create metric exporter: %w", err)
	}
	return exporter, err
}
