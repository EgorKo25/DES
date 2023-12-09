package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getWriteSyncer(filename string) zapcore.WriteSyncer {
	err := os.MkdirAll("logs", 0750)
	if err != nil {
		panic(err)
	}

	filename = filepath.Join("logs/", filepath.Clean(filename))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(file)
}

func NewLogger() map[string]*zap.Logger {
	warErrLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if lvl >= zapcore.WarnLevel && lvl < zapcore.DPanicLevel {
			return true
		}
		return false
	})

	console := zapcore.Lock(os.Stdout)

	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, console, zap.DebugLevel),
		zapcore.NewCore(consoleEncoder, console, warErrLvl),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("debug.log")),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("warning_error.log")),
			warErrLvl,
		),
	)

	logger := zap.New(core)

	core = zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, console, zap.InfoLevel),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("http.log")),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	httpLogger := zap.New(core)

	core = zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, console, zap.InfoLevel),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("grpc.log")),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	grpcLogger := zap.New(core)

	loggerMap := map[string]*zap.Logger{}

	loggerMap["logger"] = logger
	loggerMap["grpc"] = grpcLogger
	loggerMap["http"] = httpLogger

	return loggerMap
}
