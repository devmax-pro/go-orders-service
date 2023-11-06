package logger

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"os"

	"go.uber.org/zap"
)

var log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	log = createLogger(lvl)
	return nil
}

func createLogger(lvl zap.AtomicLevel) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             lvl,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console", // "json"
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		InitialFields:     map[string]interface{}{"pid": os.Getpid()},
	}

	return zap.Must(config.Build())
}

func Debug(msg string) {
	log.Debug(msg)
}

func Debugf(format string, a ...any) {
	log.Debug(fmt.Sprintf(format, a))
}

func Info(msg string) {
	log.Info(msg)
}

func Infof(format string, a ...any) {
	log.Info(fmt.Sprintf(format, a))
}

func Warn(msg string) {
	log.Warn(msg)
}

// Error log message with error as option
func Error(msg string, err ...error) {
	if len(err) == 0 {
		log.Error(msg)
		return
	}
}

func Fatal(msg string, err error) {
	log.Error(
		msg,
		zap.Error(err),
	)
	os.Exit(1)
}
