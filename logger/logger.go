package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"time"
)

/**
代补充部分：
1、更丰富的日志轮转，比如按天切割
2、自定义日志模块初始化失败时，提供一个默认的日志
*/

//var defaultLogger *zap.Logger
var customLogger *zap.Logger

func init() {
	//初始化customLogger
	customLogger = newCustomLogger()
	if customLogger != nil {
		Info("log module init success!")
		return
	}

	log.Panic("log module init failed, system exit!")

	// fixme 提供一个默认的日志
	//if customLogger == nil {
	//	logger, err := zap.NewProduction()
	//	if err != nil {
	//		log.Panic("init zap SugaredLogger failed, error=", err)
	//	}
	//	customLogger = logger.WithOptions(zap.AddCallerSkip(1))
	//}
}

func Errorf(template string, args ...interface{}) {
	customLogger.Error(fmt.Sprintf(template, args...))
}

func Error(msg string) {
	customLogger.Error(msg)
}

func Infof(template string, args ...interface{}) {
	customLogger.Info(fmt.Sprintf(template, args...))
}

func Info(msg string) {
	customLogger.Info(msg)
}

func Warnf(template string, args ...interface{}) {
	customLogger.Warn(fmt.Sprintf(template, args...))
}

func Warn(msg string) {
	customLogger.Warn(msg)
}

func newCustomLogger() *zap.Logger {
	return zap.New(zapcore.NewCore(getEncoder(), getWriteSyncer(), zapcore.InfoLevel)).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))

}

func getEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()

	cfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder

	return zapcore.NewConsoleEncoder(cfg)
}


//fixme 更丰富的日志轮转，比如按天切割
func getWriteSyncer() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:  "./log/test.log",
		MaxAge:    7,
		MaxSize:   10,
		LocalTime: true,
		Compress:  true,
	}

	return zapcore.AddSync(lumberJackLogger)
}
