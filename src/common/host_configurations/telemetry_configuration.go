package host_configurations

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/settings"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	metricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metrichttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
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

		exporters, err := newOtlpExporterFactory(telemetrySettings).build(ctx)
		if err != nil {
			logger.Error("Failed to create OTLP exporters", zap.Error(err))
			return err
		}

		credMode := "tls"
		if telemetrySettings.Insecure {
			credMode = "insecure"
		}

		authMode := "none"
		if telemetrySettings.AuthScheme != settings.TelemetryAuthSchemeNone {
			authMode = telemetrySettings.AuthScheme

			if telemetrySettings.Insecure {
				logger.Warn("telemetry credential sent over an insecure connection")
			}
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
			sdktrace.WithBatcher(exporters.trace),
			sdktrace.WithResource(res),
		)

		mp := metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exporters.metric)),
			metric.WithResource(res),
		)

		lp := log.NewLoggerProvider(
			log.WithProcessor(log.NewBatchProcessor(exporters.log)),
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
			zap.String("protocol", string(telemetrySettings.Protocol)),
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

// otlpExporters holds one exporter per signal, all speaking the configured OTLP protocol.
type otlpExporters struct {
	trace  sdktrace.SpanExporter
	metric metric.Exporter
	log    log.Exporter
}

// otlpExporterFactory builds the OTLP exporters over gRPC or HTTP from the telemetry settings.
type otlpExporterFactory struct {
	telemetrySettings *settings.TelemetrySettings
}

func newOtlpExporterFactory(telemetrySettings *settings.TelemetrySettings) *otlpExporterFactory {
	return &otlpExporterFactory{telemetrySettings: telemetrySettings}
}

func (f *otlpExporterFactory) build(ctx context.Context) (*otlpExporters, error) {
	if f.telemetrySettings.Protocol == settings.TelemetryProtocolHttp {
		return f.newHttpExporters(ctx)
	}
	return f.newGrpcExporters(ctx)
}

func (f *otlpExporterFactory) newGrpcExporters(ctx context.Context) (*otlpExporters, error) {
	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(f.telemetrySettings.Endpoint)}
	logOpts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(f.telemetrySettings.Endpoint)}
	metricOpts := []metricgrpc.Option{metricgrpc.WithEndpoint(f.telemetrySettings.Endpoint)}

	if f.telemetrySettings.Insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
		logOpts = append(logOpts, otlploggrpc.WithInsecure())
		metricOpts = append(metricOpts, metricgrpc.WithInsecure())
	} else {
		creds := credentials.NewTLS(&tls.Config{})
		traceOpts = append(traceOpts, otlptracegrpc.WithTLSCredentials(creds))
		logOpts = append(logOpts, otlploggrpc.WithTLSCredentials(creds))
		metricOpts = append(metricOpts, metricgrpc.WithTLSCredentials(creds))
	}

	if headers := f.headers(); headers != nil {
		traceOpts = append(traceOpts, otlptracegrpc.WithHeaders(headers))
		logOpts = append(logOpts, otlploggrpc.WithHeaders(headers))
		metricOpts = append(metricOpts, metricgrpc.WithHeaders(headers))
	}

	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP gRPC trace exporter: %w", err)
	}

	logExporter, err := otlploggrpc.New(ctx, logOpts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP gRPC log exporter: %w", err)
	}

	metricExporter, err := metricgrpc.New(ctx, metricOpts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP gRPC metric exporter: %w", err)
	}

	return &otlpExporters{trace: traceExporter, metric: metricExporter, log: logExporter}, nil
}

func (f *otlpExporterFactory) newHttpExporters(ctx context.Context) (*otlpExporters, error) {
	traceOpts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(f.telemetrySettings.Endpoint)}
	logOpts := []otlploghttp.Option{otlploghttp.WithEndpoint(f.telemetrySettings.Endpoint)}
	metricOpts := []metrichttp.Option{metrichttp.WithEndpoint(f.telemetrySettings.Endpoint)}

	if f.telemetrySettings.Insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		logOpts = append(logOpts, otlploghttp.WithInsecure())
		metricOpts = append(metricOpts, metrichttp.WithInsecure())
	} else {
		tlsConfig := &tls.Config{}
		traceOpts = append(traceOpts, otlptracehttp.WithTLSClientConfig(tlsConfig))
		logOpts = append(logOpts, otlploghttp.WithTLSClientConfig(tlsConfig))
		metricOpts = append(metricOpts, metrichttp.WithTLSClientConfig(tlsConfig))
	}

	if headers := f.headers(); headers != nil {
		traceOpts = append(traceOpts, otlptracehttp.WithHeaders(headers))
		logOpts = append(logOpts, otlploghttp.WithHeaders(headers))
		metricOpts = append(metricOpts, metrichttp.WithHeaders(headers))
	}

	traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP trace exporter: %w", err)
	}

	logExporter, err := otlploghttp.New(ctx, logOpts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP log exporter: %w", err)
	}

	metricExporter, err := metrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP metric exporter: %w", err)
	}

	return &otlpExporters{trace: traceExporter, metric: metricExporter, log: logExporter}, nil
}

// headers returns the collector auth header, nil when telemetry sends no credential.
func (f *otlpExporterFactory) headers() map[string]string {
	header, ok := f.telemetrySettings.AuthorizationHeader()
	if !ok {
		return nil
	}
	return map[string]string{"authorization": header}
}
