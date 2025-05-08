package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	logFile *os.File
	Debug   = log("DEBUG")
	Info    = log("INFO")
	Warn    = log("WARN")
	Error   = log("ERROR")
)

func Init() error {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return err
	}

	// 创建日志文件
	filename := filepath.Join("logs", fmt.Sprintf("agent_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logFile = file
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func log(level string) func(v ...interface{}) {
	return func(v ...interface{}) {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		msg := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, fmt.Sprint(v...))
		
		// 同时输出到控制台和文件
		fmt.Print(msg)
		if logFile != nil {
			logFile.WriteString(msg)
		}
	}
}