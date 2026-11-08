package host_configurations

import (
	"context"
	"crypto/tls"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/settings"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	metricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.uber.org/dig"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// ConfigureTelemetry wires the global OTel SDK from settings and returns a shutdown func (noop when disabled).
func ConfigureTelemetry(container *dig.Container, serviceName string) func(context.Context) error {
	shutdown := func(ctx context.Context) error { return nil }

	if err := container.Provide(settings.NewTelemetrySettings); err != nil {
		panic(err)
	}

	err := container.Invoke(func(telemetrySettings *settings.TelemetrySettings, logger *zap.Logger) error {
		if telemetrySettings == nil || !telemetrySettings.Enabled {
			logger.Info("OpenTelemetry telemetry disabled")
			return nil
		}

		ctx := context.Background()

		credMode := "tls"
		traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(telemetrySettings.Endpoint)}
		logOpts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(telemetrySettings.Endpoint)}
		metricOpts := []metricgrpc.Option{metricgrpc.WithEndpoint(telemetrySettings.Endpoint)}

		if telemetrySettings.Insecure {
			traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
			logOpts = append(logOpts, otlploggrpc.WithInsecure())
			metricOpts = append(metricOpts, metricgrpc.WithInsecure())
			credMode = "insecure"
		} else {
			creds := credentials.NewTLS(&tls.Config{})
			traceOpts = append(traceOpts, otlptracegrpc.WithTLSCredentials(creds))
			logOpts = append(logOpts, otlploggrpc.WithTLSCredentials(creds))
			metricOpts = append(metricOpts, metricgrpc.WithTLSCredentials(creds))
		}

		authMode := "none"
		if header, ok := telemetrySettings.AuthorizationHeader(); ok {
			headers := map[string]string{"authorization": header}
			traceOpts = append(traceOpts, otlptracegrpc.WithHeaders(headers))
			logOpts = append(logOpts, otlploggrpc.WithHeaders(headers))
			metricOpts = append(metricOpts, metricgrpc.WithHeaders(headers))
			authMode = telemetrySettings.AuthScheme

			if telemetrySettings.Insecure {
				logger.Warn("telemetry credential sent over an insecure connection")
			}
		}

		traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
		if err != nil {
			logger.Error("Failed to create OTLP trace exporter", zap.Error(err))
			return err
		}

		logExporter, err := otlploggrpc.New(ctx, logOpts...)
		if err != nil {
			logger.Error("Failed to create OTLP log exporter", zap.Error(err))
			return err
		}

		metricExporter, err := metricgrpc.New(ctx, metricOpts...)
		if err != nil {
			logger.Error("Failed to create OTLP metric exporter", zap.Error(err))
			return err
		}

		res, err := resource.New(ctx,
			resource.WithSchemaURL(semconv.SchemaURL),
			resource.WithAttributes(
				semconv.ServiceName(serviceName),
				semconv.ServiceNamespace("agentic"),
				// Tags every signal with the emitting instance and environment so a shared Grafana can filter by them.
				semconv.ServiceInstanceID(telemetrySettings.InstanceId),
				semconv.DeploymentEnvironmentName(telemetrySettings.Environment),
			),
		)
		if err != nil {
			logger.Error("Failed to build telemetry resource", zap.Error(err))
			return err
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)

		mp := metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(metricExporter)),
			metric.WithResource(res),
		)

		lp := log.NewLoggerProvider(
			log.WithProcessor(log.NewBatchProcessor(logExporter)),
			log.WithResource(res),
		)

		otel.SetTracerProvider(tp)
		otel.SetMeterProvider(mp)
		logglobal.SetLoggerProvider(lp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

		shutdown = func(ctx context.Context) error {
			var shutdownErr error
			if err := tp.Shutdown(ctx); err != nil {
				shutdownErr = multierr.Append(shutdownErr, err)
			}
			if err := mp.Shutdown(ctx); err != nil {
				shutdownErr = multierr.Append(shutdownErr, err)
			}
			if err := lp.Shutdown(ctx); err != nil {
				shutdownErr = multierr.Append(shutdownErr, err)
			}
			return shutdownErr
		}

		logger.Info("OpenTelemetry telemetry enabled",
			zap.String("endpoint", telemetrySettings.Endpoint),
			zap.String("service", serviceName),
			zap.String("creds", credMode),
			zap.String("auth", authMode),
		)
		return nil
	})
	if err != nil {
		panic(err)
	}

	return shutdown
}
