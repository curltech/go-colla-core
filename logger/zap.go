package logger

import (
	"github.com/curltech/go-colla-core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {
	filePath, _ := config.GetString("log.filePath", "./logs/spikeProxy1.log")
	hook := lumberjack.Logger{
		Filename:   filePath, // 日志文件路径
		MaxSize:    128,      // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,       // 日志文件最多保存多少个备份
		MaxAge:     7,        // 文件最多保存多少天
		Compress:   true,     // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	level, _ := config.GetString("log.level", "debug")
	err := atomicLevel.UnmarshalText([]byte(level))
	if err != nil {
		atomicLevel.SetLevel(zap.InfoLevel)
	}
	encode, _ := config.GetString("log.encoder", "json")
	var encoder zapcore.Encoder
	if encode == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	core := zapcore.NewCore(
		encoder, // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段
	fs := make([]zap.Field, 0)
	serviceName, _ := config.GetString("log.serviceName")
	if serviceName != "" {
		field := zap.String("serviceName", serviceName)
		fs = append(fs, field)
	}
	fields := zap.Fields(fs...)
	// 构造日志
	Logger = zap.New(core, caller, development, fields)
	Sugar = Logger.Sugar()
}
