package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 封装 zap.Logger 的接口
type Logger struct {
	*zap.Logger
}

var defaultLogger *Logger

// Init 初始化日志配置
func Init(level string, format string) error {
	var cfg zap.Config

	// 设置日志级别
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	// 设置日志格式
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// 配置日志
	cfg.Level = zap.NewAtomicLevelAt(logLevel)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.Development = true

	// 创建 logger 时启用调用者信息和堆栈跟踪
	logger, err := cfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	defaultLogger = &Logger{logger}
	return nil
}

// Sync 同步日志
func Sync() error {
	return defaultLogger.Sync()
}

// L 获取默认日志实例
func L() *Logger {
	return defaultLogger
}

// Named 创建命名日志实例
func Named(name string) *Logger {
	return &Logger{defaultLogger.Named(name)}
}

// With 创建带有字段的日志实例
func With(fields ...zap.Field) *Logger {
	return &Logger{defaultLogger.With(fields...)}
}

// Debug 输出调试日志
func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

// Info 输出信息日志
func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn 输出警告日志
func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error 输出错误日志
func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

// Fatal 输出致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	defaultLogger.Fatal(msg, fields...)
}
