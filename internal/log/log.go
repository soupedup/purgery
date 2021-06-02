// Package log implements logging functionality.
package log

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const stampLayout = "2006-01-02 15:04:05.000"

var cfg = zap.Config{
	Encoding:         "console",
	Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
	EncoderConfig: zapcore.EncoderConfig{
		MessageKey:     "message",
		TimeKey:        "time",
		EncodeTime:     zapcore.TimeEncoderOfLayout(stampLayout),
		NameKey:        "logger",
		LevelKey:       "level",
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.StringDurationEncoder,
	},
}

var logger *zap.Logger // setup during init

func init() {
	if strings.ToLower(os.Getenv("TRACE_FORMAT")) == "json" {
		// we're using structured logging
		cfg.Encoding = "json"
		cfg.EncoderConfig.TimeKey = zapcore.OmitKey
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	var level zapcore.Level
	switch strings.ToLower(os.Getenv("TRACE_LEVEL")) {
	default:
		level = zapcore.InfoLevel
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	var err error
	if logger, err = cfg.Build(); err != nil {
		panic(err)
	}
}

// New returns a reference to a Logger with the given name.
func New(name string) *zap.Logger {
	return logger.Named(name)
}

// Checkpoint is shorthand for zap.String("checkpoint", v).
func Checkpoint(v string) zap.Field {
	return zap.String("checkpoint", v)
}

// URL is shorthand for zap.String("url", v).
func URL(v string) zap.Field {
	return zap.String("url", v)
}
