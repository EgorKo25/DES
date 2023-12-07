package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getWriteSyncer(filename string) zapcore.WriteSyncer {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
			zapcore.AddSync(getWriteSyncer("logs/debug.log")),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("logs/warning_error.log")),
			warErrLvl,
		),
	)

	logger := zap.New(core)

	core = zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, console, zap.InfoLevel),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("logs/http.log")),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	httpLogger := zap.New(core)

	core = zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, console, zap.InfoLevel),
		zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(getWriteSyncer("logs/grpc.log")),
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
