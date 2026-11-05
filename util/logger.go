package util

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogFormat int

const (
	LogFormatText LogFormat = iota
	LogFormatJSON
	LogFormatCSV
)

func (f LogFormat) String() string {
	switch f {
	case LogFormatJSON:
		return "json"
	case LogFormatCSV:
		return "csv"
	default:
		return "text"
	}
}

func ParseLogFormat(s string) LogFormat {
	switch s {
	case "json", "JSON":
		return LogFormatJSON
	case "csv", "CSV":
		return LogFormatCSV
	default:
		return LogFormatText
	}
}

type Logger interface {
	General(format string, args ...any)
	Audit(format string, args ...any)
	Crash(msg string, recovered any, stack []byte)
	Close() error
}

type LoggerConfig struct {
	Dir         string
	Format      LogFormat
	Ext         string
	GeneralFile string
	AuditFile   string
	CrashFile   string
	Console     bool
	Writer      io.Writer
}

type FileLogger struct {
	mu          sync.Mutex
	cfg         LoggerConfig
	generalFile *os.File
	auditFile   *os.File
	crashFile   *os.File
	writer      io.Writer
	closed      bool
}

func NewLogger(configs ...LoggerConfig) (*FileLogger, error) {
	cfg := LoggerConfig{
		Dir:     ".logs",
		Format:  LogFormatText,
		Console: false,
	}
	if len(configs) > 0 {
		c := configs[0]
		if c.Dir != "" {
			cfg.Dir = c.Dir
		}
		cfg.Format = c.Format
		cfg.Ext = c.Ext
		cfg.GeneralFile = c.GeneralFile
		cfg.AuditFile = c.AuditFile
		cfg.CrashFile = c.CrashFile
		cfg.Console = c.Console
		cfg.Writer = c.Writer
	}

	ext := cfg.Ext
	if ext == "" {
		switch cfg.Format {
		case LogFormatJSON:
			ext = ".json"
		case LogFormatCSV:
			ext = ".csv"
		default:
			ext = ".log"
		}
	}
	if !filepath.IsAbs(ext) && ext[0] != '.' {
		ext = "." + ext
	}

	if cfg.GeneralFile == "" {
		cfg.GeneralFile = "general" + ext
	}
	if cfg.AuditFile == "" {
		cfg.AuditFile = "audit" + ext
	}
	if cfg.CrashFile == "" {
		cfg.CrashFile = "crash" + ext
	}

	l := &FileLogger{
		cfg:    cfg,
		writer: cfg.Writer,
	}

	if l.writer == nil {
		if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
			return nil, err
		}

		genPath := filepath.Join(cfg.Dir, cfg.GeneralFile)
		gf, err := os.OpenFile(genPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		l.generalFile = gf

		auditPath := filepath.Join(cfg.Dir, cfg.AuditFile)
		af, err := os.OpenFile(auditPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			_ = l.Close()
			return nil, err
		}
		l.auditFile = af

		crashPath := filepath.Join(cfg.Dir, cfg.CrashFile)
		cf, err := os.OpenFile(crashPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			_ = l.Close()
			return nil, err
		}
		l.crashFile = cf
	}

	return l, nil
}

func (l *FileLogger) General(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	data := l.formatLine("INFO", "general", msg, nil)
	l.write("general", data)
}

func (l *FileLogger) Audit(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	data := l.formatLine("AUDIT", "audit", msg, nil)
	l.write("audit", data)
}

func (l *FileLogger) Crash(msg string, recovered any, stack []byte) {
	fullMsg := msg
	if recovered != nil {
		fullMsg = fmt.Sprintf("%s: %v", msg, recovered)
	}
	data := l.formatLine("CRASH", "crash", fullMsg, stack)
	l.write("crash", data)
}

func (l *FileLogger) write(category string, data []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return
	}

	if l.writer != nil {
		_, _ = l.writer.Write(data)
	} else {
		var f *os.File
		switch category {
		case "general":
			f = l.generalFile
		case "audit":
			f = l.auditFile
		case "crash":
			f = l.crashFile
		}
		if f != nil {
			_, _ = f.Write(data)
		}
	}

	if l.cfg.Console {
		_, _ = os.Stderr.Write(data)
	}
}

func (l *FileLogger) formatLine(level, category, message string, stack []byte) []byte {
	now := time.Now().UTC()

	switch l.cfg.Format {
	case LogFormatJSON:
		entry := map[string]any{
			"time":     now.Format(time.RFC3339),
			"level":    level,
			"category": category,
			"message":  message,
		}
		if len(stack) > 0 {
			entry["stack"] = string(stack)
		}
		b, _ := json.Marshal(entry)
		b = append(b, '\n')
		return b

	case LogFormatCSV:
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write([]string{
			now.Format(time.RFC3339),
			level,
			category,
			message,
			string(stack),
		})
		w.Flush()
		return buf.Bytes()

	default:
		if len(stack) > 0 {
			return fmt.Appendf(nil, "%s [%s] [%s] %s\nstack:\n%s\n",
				now.Format("2006-01-02 15:04:05"), level, category, message, string(stack))
		}
		return fmt.Appendf(nil, "%s [%s] [%s] %s\n",
			now.Format("2006-01-02 15:04:05"), level, category, message)
	}
}

func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}
	l.closed = true

	var errs []error
	if l.generalFile != nil {
		if err := l.generalFile.Close(); err != nil {
			errs = append(errs, err)
		}
		l.generalFile = nil
	}
	if l.auditFile != nil {
		if err := l.auditFile.Close(); err != nil {
			errs = append(errs, err)
		}
		l.auditFile = nil
	}
	if l.crashFile != nil {
		if err := l.crashFile.Close(); err != nil {
			errs = append(errs, err)
		}
		l.crashFile = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
