package logwrapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/teitiago/task-manager-poc/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger Initializes the log to be used on the app.
// This will replace the global zap logger and used these configurations.
func InitLogger() {

	// start configs
	createLogDir()
	writerSync := getLogWriter()
	encoder := getEncoder()

	var logLevel zapcore.Level
	switch level := strings.ToLower(config.GetEnv("LOG_LEVEL", "debug")); level {
	case "debug":
		logLevel = zapcore.DebugLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	// create core for stdout and file
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, writerSync, logLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel),
	)

	log := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(log)
}

// createLogDir Creates the log directory to be used to store the logs.
func createLogDir() {
	path, _ := os.Getwd()
	if _, err := os.Stat(fmt.Sprintf("%s/logs", path)); os.IsNotExist(err) {
		_ = os.Mkdir("logs", os.ModePerm)
	}
}

// getLogWriter Gets writesync object that will store the logs
// in one file.
func getLogWriter() zapcore.WriteSyncer {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filePath := filepath.Clean(path + "/logs/tasks.txt")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(file)
}

// getEncoder Creates a new zapcore encoder.
// This encoder uses a specific time format in UTC and a special encode level
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(t.UTC().Format("2006-01-02T15:04:05Z0700"))
	})
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}
