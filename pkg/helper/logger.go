package helper

import (
	"github.com/aliexpressru/alilo-agent/internal/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func ConfigurationLogger(logger *zap.Logger, logFileName string, maxLogSize int, maxLogBackups int, maxLogAge int, compress bool, atomicLevel zap.AtomicLevel) {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    maxLogSize,
		MaxBackups: maxLogBackups,
		MaxAge:     maxLogAge,
		Compress:   compress,
	})

	zap.ReplaceGlobals(logger.WithOptions(
		zap.WrapCore(
			func(core2 zapcore.Core) zapcore.Core {
				return zapcore.NewCore(
					zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
					w,
					atomicLevel)
			})))
}

func CreateLogger(logFileName string, maxLogSize int, maxLogBackups int, maxLogAge int, compress bool) (zap.AtomicLevel, error) {
	atomicLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := zap.Config{
		Level:             atomicLevel,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stdout", logFileName},
		ErrorOutputPaths:  []string{"stderr", logFileName},
		DisableStacktrace: false,
	}.Build()
	zap.ReplaceGlobals(logger)
	ConfigurationLogger(logger, logFileName, maxLogSize, maxLogBackups, maxLogAge, compress, atomicLevel)
	return atomicLevel, err
}

func SetLoggerLevel(atomicLevel zap.AtomicLevel, cfg model.Config) {
	curLevel := atomicLevel.Level()
	newValue := cfg.LogLevel
	if loggErr := curLevel.Set(newValue); loggErr == nil {
		atomicLevel.SetLevel(curLevel)
		zap.S().Debugf("atomicLevel.SetLevel %v -> %v", curLevel.String(), newValue)
		zap.S().Warn("Set Logger Level: ", atomicLevel.Level().String())
	}
}
