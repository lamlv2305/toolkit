package rok

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger

func Setup(debug bool) {
	var encoderConfig zapcore.EncoderConfig
	var config zap.Config

	if debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		config = zap.NewDevelopmentConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		config = zap.NewProductionConfig()
	}

	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderConfig.StacktraceKey = "" // to hide stacktrace info
	config.EncoderConfig = encoderConfig

	var err error
	zapLog, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(message string, fields ...zap.Field) {
	if zapLog == nil {
		return
	}

	zapLog.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	if zapLog == nil {
		return
	}

	zapLog.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	if zapLog == nil {
		return
	}

	zapLog.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	if zapLog == nil {
		return
	}

	zapLog.Fatal(message, fields...)
}
