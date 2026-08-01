package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu           sync.Mutex
	out          io.Writer
	logFile      *os.File
	maskedTokens []string
}

func NewLogger(logFilePath string, secretsToMask ...string) (*Logger, error) {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	var logFile *os.File
	if logFilePath != "" {
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", logFilePath, err)
		}
		logFile = f
		writers = append(writers, f)
	}

	var cleanedSecrets []string
	for _, s := range secretsToMask {
		if strings.TrimSpace(s) != "" {
			cleanedSecrets = append(cleanedSecrets, strings.TrimSpace(s))
		}
	}

	return &Logger{
		out:          io.MultiWriter(writers...),
		logFile:      logFile,
		maskedTokens: cleanedSecrets,
	}, nil
}

func (l *Logger) logMsg(level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	sanitized := msg
	for _, token := range l.maskedTokens {
		if token != "" {
			sanitized = strings.ReplaceAll(sanitized, token, "[REDACTED_SECRET]")
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, sanitized)
	_, _ = l.out.Write([]byte(entry))
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.logMsg("INFO", fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.logMsg("WARN", fmt.Sprintf(format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.logMsg("ERROR", fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.logMsg("DEBUG", fmt.Sprintf(format, args...))
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logFile != nil {
		_ = l.logFile.Close()
	}
}

// Global logger helper for standard log package integration
func SetGlobalLogger(l *Logger) {
	log.SetOutput(l.out)
	log.SetFlags(0)
}
