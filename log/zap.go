package log

import (
	"github.com/chencheng8888/GoDo/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

func NewZapSugaredLogger(c *config.LogConfig) (*zap.SugaredLogger, error) {
	logLevel := map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}
	writeSyncer, err := getLogWriter(c.Path, c.FileName, c.MaxSize, c.MaxAge, c.Compress, c.Stdout) // 日志文件配置 文件位置和切割
	if err != nil {
		return nil, err
	}
	encoder := getEncoder(c.Format) // 获取日志输出编码
	level, ok := logLevel[c.Level]  // 日志打印级别
	if !ok {
		level = logLevel["info"]
	}
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger.Sugar(), nil
}

// getEncoder 编码器(如何写入日志)
func getEncoder(format string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 输出level序列化为全大写字符串，如 INFO DEBUG ERROR
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	if format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	if format == "console" {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getLogWriter 获取日志输出方式  日志文件 控制台
func getLogWriter(logPath, fileName string, maxSize, maxAge int, compress, stdout bool) (zapcore.WriteSyncer, error) {
	exist, err := isPathExist(logPath)
	if err != nil {
		return nil, err
	}

	// 判断日志路径是否存在，如果不存在就创建
	if !exist {
		if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// 日志文件 与 日志切割 配置
	lumberJackLogger := newLumberjackLogger(logPath, fileName, maxSize, maxAge, compress)
	if stdout {
		// 日志同时输出到控制台和日志文件中
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout)), nil
	} else {
		// 日志只输出到日志文件
		return zapcore.AddSync(lumberJackLogger), nil
	}
}

func isPathExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsExist(err) {
		return false, nil
	}
	return false, err
}

func newLumberjackLogger(logPath, logFileName string, fileMaxSize, logMaxAge int, logCompress bool) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename: filepath.Join(logPath, logFileName), // 日志文件路径
		MaxSize:  fileMaxSize,                         // 单个日志文件最大多少 mb
		MaxAge:   logMaxAge,                           // 日志最长保留时间
		Compress: logCompress,                         // 是否压缩日志
	}
}
