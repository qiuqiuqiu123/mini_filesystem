package logger

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
)

type logWriter struct {
	calldepth int
	writer    io.Writer
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	// 调整 calldepth 参数跳过封装层数
	_, file, line, ok := runtime.Caller(lw.calldepth)
	if ok {
		msg := fmt.Sprintf("%s:%d %s", filepath.Base(file), line, string(p))
		return lw.writer.Write([]byte(msg))
	}
	return lw.writer.Write(p)
}
