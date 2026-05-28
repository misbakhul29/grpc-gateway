package logger

import (
	"fmt"
	"grpc/gateway/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zapLogger *zap.Logger
}

var Log *Logger

func init() {
	var z *zap.Logger
	var err error

	if config.EnvCfg.NodeEnv == "production" {
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		z, err = config.Build()
	} else {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		z, err = config.Build()
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger: %v", err))
	}

	Log = &Logger{
		zapLogger: z,
	}
}

func (l *Logger) Info(service string, message string) {
	l.zapLogger.Info(message, zap.String("service", service))
}

func (l *Logger) Error(service string, message string) {
	l.zapLogger.Error(message, zap.String("service", service))
}

func (l *Logger) Warn(service string, message string) {
	l.zapLogger.Warn(message, zap.String("service", service))
}

func (l *Logger) Debug(service string, message string) {
	if config.EnvCfg.NodeEnv == "production" {
		return
	}
	l.zapLogger.Debug(message, zap.String("service", service))
}

func (l *Logger) Fatal(service string, message string) {
	l.zapLogger.Fatal(message, zap.String("service", service))
}

func (l *Logger) Panic(service string, message string) {
	l.zapLogger.Panic(message, zap.String("service", service))
}
