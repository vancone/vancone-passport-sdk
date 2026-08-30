package scheduler

import (
	"os"
	"testing"

	"github.com/vancone/vancone-web-common-go/pkg/logger"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// 隔离测试输出：用 nop logger 替代文件日志，同时避免未初始化 logger 引发的 panic
	logger.SugaredLogger = zap.NewNop().Sugar()
	os.Exit(m.Run())
}
