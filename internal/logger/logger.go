package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLoggers() []*zap.Logger {
	// Настройка конфигурации для вывода трафика gRPC
	grpcConfig := zap.NewDevelopmentConfig()
	grpcConfig.OutputPaths = []string{"logs/", "stdout"}

	// Создание ядра для вывода трафика gRPC
	grpcCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(grpcConfig.EncoderConfig),
		zapcore.AddSync(getWriteSyncer("logs/grpc.log")), // Файл для вывода трафика gRPC
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)

	// Настройка конфигурации для вывода трафика HTTP
	httpConfig := zap.NewDevelopmentConfig()
	httpConfig.OutputPaths = []string{"logs/", "stdout"}

	// Создание ядра для вывода трафика HTTP
	httpCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(httpConfig.EncoderConfig),
		zapcore.AddSync(getWriteSyncer("logs/http.log")), // Файл для вывода трафика HTTP
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)

	// Настройка конфигурации для вывода логов уровня warning и error
	warningErrorConfig := zap.NewDevelopmentConfig()
	warningErrorConfig.OutputPaths = []string{"logs/", "stdout"}

	// Создание ядра для вывода логов уровня warning и error
	warningErrorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(warningErrorConfig.EncoderConfig),
		zapcore.AddSync(getWriteSyncer("logs/warning_error.log")), // Файл для вывода логов уровня warning и error
		zap.NewAtomicLevelAt(zap.WarnLevel),
	)

	// Настройка конфигурации для вывода отладочной информации
	debugConfig := zap.NewDevelopmentConfig()
	debugConfig.OutputPaths = []string{"logs/", "stdout"}

	// Создание ядра для вывода отладочной информации
	debugCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(debugConfig.EncoderConfig),
		zapcore.AddSync(getWriteSyncer("logs/debug.log")), // Файл для вывода отладочной информации
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)

	log := make([]*zap.Logger, 0)
	logger := zap.New(
		zapcore.NewTee(warningErrorCore, debugCore),
	)

	grpcLogger := zap.New(grpcCore)
	httpLogger := zap.New(httpCore)

	log = append(log, logger)
	log = append(log, grpcLogger)
	log = append(log, httpLogger)

	return log
}

func getWriteSyncer(filename string) zapcore.WriteSyncer {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(file)
}
