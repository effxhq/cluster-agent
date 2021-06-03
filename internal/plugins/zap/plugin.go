package zap_plugin

import (
	"context"

	"github.com/effxhq/go-lifecycle"
	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

const contextKey = lifecycle.ContextKey("logger")

var defaultLogger = zap.NewNop()

func FromContext(ctx context.Context) *zap.Logger {
	val := ctx.Value(contextKey)
	if val == nil {
		return defaultLogger
	}
	return val.(*zap.Logger)
}

func Plugin() lifecycle.Plugin {
	var logger *zap.Logger

	return &lifecycle.PluginFuncs{
		InitializeFunc: func(app *lifecycle.Application) (err error) {
			cfg := zap.NewProductionConfig()
			cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

			logger, err = cfg.Build()
			if err != nil {
				return err
			}

			app.WithValue(contextKey, logger)
			app.WithHook(func(phase string, err error) {
				if err != nil {
					logger.Error("encountered error", zap.String("phase", phase), zap.Error(err))
				}
			})

			return nil
		},
		StartFunc: func(app *lifecycle.Application) error {
			logger.Info("starting agent")
			return nil
		},
		RunFunc: func(app *lifecycle.Application) error {
			logger.Info("running job")
			return nil
		},
		ShutdownFunc: func(app *lifecycle.Application) error {
			logger.Info("shutting down application")
			return nil
		},
	}
}
