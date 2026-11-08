package host_configurations

import (
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logBridgeScope = "github.com/6022-labs/agentic-x402-facilitator"

func ConfigureLogging(container *dig.Container) {
	logLevel := os.Getenv("LOG_LEVEL")

	zapConfig := zap.NewProductionConfig()

	switch logLevel {
	case "debug":
		zapConfig.Level.SetLevel(zap.DebugLevel)
	case "info", "information":
		zapConfig.Level.SetLevel(zap.InfoLevel)
	case "warn", "warning":
		zapConfig.Level.SetLevel(zap.WarnLevel)
	case "error":
		zapConfig.Level.SetLevel(zap.ErrorLevel)
	case "fatal":
		zapConfig.Level.SetLevel(zap.FatalLevel)
	case "panic":
		zapConfig.Level.SetLevel(zap.PanicLevel)
	default:
		zapConfig.Level.SetLevel(zap.InfoLevel)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}

	// Resolves the logger provider lazily, so ConfigureTelemetry may run after this.
	otelCore := otelzap.NewCore(logBridgeScope, otelzap.WithLoggerProvider(logglobal.GetLoggerProvider()))
	logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, otelCore)
	}))

	err = container.Provide(func() *zap.Logger {
		return logger
	})
	if err != nil {
		logger.Fatal("failed to provide logger", zap.Error(err))
	}
}
