package logger

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func countNonEmptyLines(t *testing.T, filePath string) int {
	t.Helper()
	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return 0
	}
	lines := strings.Split(text, "\n")
	count := 0
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			count++
		}
	}
	return count
}

func TestNewWithOptions_SamplingInfoBurst(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "sampling-info-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	logPath := tmpFile.Name()
	_ = tmpFile.Close()
	lg, err := NewWithOptions(Options{
		Level:              "debug",
		Output:             logPath,
		SamplingInitial:    100,
		SamplingThereafter: 10,
	})
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	for i := 0; i < 500; i++ {
		lg.Info("sampled-info-burst", zap.Int("idx", i))
	}
	_ = lg.Sync()

	lines := countNonEmptyLines(t, logPath)
	if lines < 120 || lines > 180 {
		t.Fatalf("sampled info lines = %d, want within [120, 180]", lines)
	}
}

func TestNewWithOptions_ErrorNotSampled(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "sampling-error-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	logPath := tmpFile.Name()
	_ = tmpFile.Close()
	lg, err := NewWithOptions(Options{
		Level:              "debug",
		Output:             logPath,
		SamplingInitial:    100,
		SamplingThereafter: 10,
	})
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	for i := 0; i < 500; i++ {
		lg.Error("error-should-not-be-sampled", zap.Int("idx", i))
	}
	_ = lg.Sync()

	lines := countNonEmptyLines(t, logPath)
	if lines != 500 {
		t.Fatalf("error lines = %d, want 500", lines)
	}
}
