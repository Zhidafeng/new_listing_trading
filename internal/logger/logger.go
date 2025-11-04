package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

// Init 初始化日志
func Init(level, filePath string, maxSize, maxAge int, compress bool) error {
	log = logrus.New()

	// 设置日志级别
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	// 设置日志格式
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	// 设置输出
	var writers []io.Writer

	// 始终输出到控制台
	writers = append(writers, os.Stdout)

	// 如果配置了日志文件，也输出到文件
	if filePath != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(filePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 使用lumberjack进行日志轮转
		logFile := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    maxSize, // MB
			MaxAge:     maxAge,  // days
			MaxBackups: 10,      // 保留的旧日志文件数量
			Compress:   compress,
		}
		writers = append(writers, logFile)
	}

	// 设置多输出
	log.SetOutput(io.MultiWriter(writers...))

	return nil
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if log == nil {
		// 如果没有初始化，使用默认配置
		log = logrus.New()
		log.SetLevel(logrus.InfoLevel)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		log.SetOutput(os.Stdout)
	}
	return log
}

// Trace 记录Trace级别日志
func Trace(args ...interface{}) {
	GetLogger().Trace(args...)
}

// Tracef 记录Trace级别格式化日志
func Tracef(format string, args ...interface{}) {
	GetLogger().Tracef(format, args...)
}

// Debug 记录Debug级别日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 记录Debug级别格式化日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 记录Info级别日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 记录Info级别格式化日志
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 记录Warn级别日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 记录Warn级别格式化日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 记录Error级别日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 记录Error级别格式化日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 记录Fatal级别日志并退出
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 记录Fatal级别格式化日志并退出
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic 记录Panic级别日志并panic
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf 记录Panic级别格式化日志并panic
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}
