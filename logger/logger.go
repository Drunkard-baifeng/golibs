package logger

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// ANSI 颜色码
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
)

// 级别对应的消息颜色
var levelColors = map[zapcore.Level]string{
	zapcore.DebugLevel: colorMagenta,
	zapcore.InfoLevel:  colorBlue,
	zapcore.WarnLevel:  colorYellow,
	zapcore.ErrorLevel: colorRed,
	zapcore.FatalLevel: colorRed,
}

func init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		&colorEncoder{Encoder: zapcore.NewConsoleEncoder(encoderConfig)},
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	log = logger.Sugar()
}

// colorEncoder 自定义编码器
type colorEncoder struct {
	zapcore.Encoder
}

func (e *colorEncoder) Clone() zapcore.Encoder {
	return &colorEncoder{Encoder: e.Encoder.Clone()}
}

func (e *colorEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	color := levelColors[entry.Level]
	if color != "" {
		entry.Message = color + entry.Message + colorReset
	}
	return e.Encoder.EncodeEntry(entry, fields)
}

// ==================== 日志方法 ====================

// Debug 调试日志（洋红色）
func Debug(msg string, keysAndValues ...interface{}) {
	log.Debugw(msg, keysAndValues...)
}

// Debugf 格式化调试日志
func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Info 信息日志（蓝色）
func Info(msg string, keysAndValues ...interface{}) {
	log.Infow(msg, keysAndValues...)
}

// Infof 格式化信息日志
func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Success 成功日志（绿色）- 用 INFO 级别但显示绿色
func Success(msg string, keysAndValues ...interface{}) {
	// 手动加绿色
	log.Infow(colorGreen+msg+colorReset, keysAndValues...)
}

// Successf 格式化成功日志
func Successf(format string, v ...interface{}) {
	log.Infof(colorGreen+format+colorReset, v...)
}

// Error 错误日志（红色）
func Error(msg string, keysAndValues ...interface{}) {
	log.Errorw(msg, keysAndValues...)
}

// Errorf 格式化错误日志
func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Warn 警告日志（黄色）
func Warn(msg string, keysAndValues ...interface{}) {
	log.Warnw(msg, keysAndValues...)
}

// Warnf 格式化警告日志
func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}
