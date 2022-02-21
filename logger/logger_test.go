package logger_test

import (
	"github.com/liuxiaodao666/go-util/logger"
	"testing"
)

func TestLogPrint(t *testing.T) {
	logger.Info("mock info")
	logger.Info("mock info")
	logger.Infof("mock %v", "info")
	logger.Warn("mock warn")
	logger.Warnf("mock %v", "warn")
	logger.Error("mock error")
	logger.Errorf("mock %v", "error")

}
