package util

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestParseLogFormat(t *testing.T) {
	if ParseLogFormat("json") != LogFormatJSON {
		t.Fatalf("expected LogFormatJSON")
	}
	if ParseLogFormat("JSON") != LogFormatJSON {
		t.Fatalf("expected LogFormatJSON")
	}
	if ParseLogFormat("csv") != LogFormatCSV {
		t.Fatalf("expected LogFormatCSV")
	}
	if ParseLogFormat("CSV") != LogFormatCSV {
		t.Fatalf("expected LogFormatCSV")
	}
	if ParseLogFormat("text") != LogFormatText {
		t.Fatalf("expected LogFormatText")
	}
	if ParseLogFormat("other") != LogFormatText {
		t.Fatalf("expected LogFormatText")
	}

	if LogFormatJSON.String() != "json" {
		t.Fatalf("expected json, got %s", LogFormatJSON.String())
	}
	if LogFormatCSV.String() != "csv" {
		t.Fatalf("expected csv, got %s", LogFormatCSV.String())
	}
	if LogFormatText.String() != "text" {
		t.Fatalf("expected text, got %s", LogFormatText.String())
	}
}

func TestLogger_DefaultText(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(LoggerConfig{Dir: dir})
	if err != nil {
		t.Fatalf("unexpected error creating logger: %v", err)
	}

	l.General("hello %s", "world")
	l.Audit("user %d changed %s", 123, "role")
	l.Crash("panic recovered", "nil pointer", []byte("goroutine 1\nmain.go:42"))

	if err := l.Close(); err != nil {
		t.Fatalf("unexpected error closing logger: %v", err)
	}

	genContent, err := os.ReadFile(filepath.Join(dir, "general.log"))
	if err != nil {
		t.Fatalf("failed to read general.log: %v", err)
	}
	if !strings.Contains(string(genContent), "[INFO] [general] hello world") {
		t.Fatalf("unexpected general log content: %s", string(genContent))
	}

	auditContent, err := os.ReadFile(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("failed to read audit.log: %v", err)
	}
	if !strings.Contains(string(auditContent), "[AUDIT] [audit] user 123 changed role") {
		t.Fatalf("unexpected audit log content: %s", string(auditContent))
	}

	crashContent, err := os.ReadFile(filepath.Join(dir, "crash.log"))
	if err != nil {
		t.Fatalf("failed to read crash.log: %v", err)
	}
	if !strings.Contains(string(crashContent), "[CRASH] [crash] panic recovered: nil pointer") {
		t.Fatalf("unexpected crash log content: %s", string(crashContent))
	}
	if !strings.Contains(string(crashContent), "stack:\ngoroutine 1\nmain.go:42") {
		t.Fatalf("unexpected crash log stack: %s", string(crashContent))
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(LoggerConfig{
		Dir:    dir,
		Format: LogFormatJSON,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.General("started service")
	l.Crash("system failure", "oom", []byte("stack trace"))

	if err := l.Close(); err != nil {
		t.Fatalf("unexpected error closing logger: %v", err)
	}

	genContent, err := os.ReadFile(filepath.Join(dir, "general.json"))
	if err != nil {
		t.Fatalf("failed to read general.json: %v", err)
	}

	var genEntry map[string]any
	if err := json.Unmarshal(genContent, &genEntry); err != nil {
		t.Fatalf("invalid json in general.json: %v", err)
	}
	if genEntry["level"] != "INFO" || genEntry["category"] != "general" || genEntry["message"] != "started service" {
		t.Fatalf("unexpected json fields: %v", genEntry)
	}

	crashContent, err := os.ReadFile(filepath.Join(dir, "crash.json"))
	if err != nil {
		t.Fatalf("failed to read crash.json: %v", err)
	}

	var crashEntry map[string]any
	if err := json.Unmarshal(crashContent, &crashEntry); err != nil {
		t.Fatalf("invalid json in crash.json: %v", err)
	}
	if crashEntry["level"] != "CRASH" || crashEntry["category"] != "crash" || crashEntry["message"] != "system failure: oom" || crashEntry["stack"] != "stack trace" {
		t.Fatalf("unexpected crash json fields: %v", crashEntry)
	}
}

func TestLogger_CSVFormat(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(LoggerConfig{
		Dir:    dir,
		Format: LogFormatCSV,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.Audit("user deleted message")

	if err := l.Close(); err != nil {
		t.Fatalf("unexpected error closing: %v", err)
	}

	auditContent, err := os.ReadFile(filepath.Join(dir, "audit.csv"))
	if err != nil {
		t.Fatalf("failed to read audit.csv: %v", err)
	}

	r := csv.NewReader(bytes.NewReader(auditContent))
	record, err := r.Read()
	if err != nil {
		t.Fatalf("failed to parse csv record: %v", err)
	}
	if len(record) != 5 {
		t.Fatalf("expected 5 columns in csv record, got %d", len(record))
	}
	if record[1] != "AUDIT" || record[2] != "audit" || record[3] != "user deleted message" {
		t.Fatalf("unexpected csv columns: %v", record)
	}
}

func TestLogger_CustomNamesAndExt(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(LoggerConfig{
		Dir:         dir,
		Ext:         "txt",
		GeneralFile: "custom_gen.txt",
		AuditFile:   "custom_audit.txt",
		CrashFile:   "custom_crash.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.General("custom")
	_ = l.Close()

	if _, err := os.Stat(filepath.Join(dir, "custom_gen.txt")); err != nil {
		t.Fatalf("expected custom_gen.txt to exist: %v", err)
	}
}

func TestLogger_CustomWriter(t *testing.T) {
	var buf bytes.Buffer
	l, err := NewLogger(LoggerConfig{
		Writer: &buf,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l.General("buffered message")
	_ = l.Close()

	if !strings.Contains(buf.String(), "buffered message") {
		t.Fatalf("expected buffer to contain message, got %s", buf.String())
	}
}

func TestLogger_Concurrent(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(LoggerConfig{Dir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			l.General("msg %d", idx)
		}(i)
		go func(idx int) {
			defer wg.Done()
			l.Audit("audit %d", idx)
		}(i)
	}
	wg.Wait()

	if err := l.Close(); err != nil {
		t.Fatalf("unexpected error closing: %v", err)
	}

	if err := l.Close(); err != nil {
		t.Fatalf("expected idempotent close, got: %v", err)
	}

	l.General("after close")
}

