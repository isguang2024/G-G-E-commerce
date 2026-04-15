package audit

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type degradedSink interface {
	Append(row *AuditLog) error
	Replay(apply func(*AuditLog) error) error
	Close() error
}

type jsonlDegradedSink struct {
	path string
	mu   sync.Mutex
	file *os.File
}

func newJSONLDegradedSink(path string) (*jsonlDegradedSink, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		return nil, errors.New("audit degraded sink path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, fmt.Errorf("create degraded sink dir: %w", err)
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open degraded sink: %w", err)
	}
	return &jsonlDegradedSink{path: target, file: f}, nil
}

func (s *jsonlDegradedSink) Append(row *AuditLog) error {
	if row == nil {
		return nil
	}
	raw, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("marshal degraded row: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return errors.New("degraded sink is closed")
	}
	if _, err := s.file.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("append degraded row: %w", err)
	}
	return s.file.Sync()
}

func (s *jsonlDegradedSink) Replay(apply func(*AuditLog) error) error {
	if apply == nil {
		return errors.New("replay apply func is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return errors.New("degraded sink is closed")
	}
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek degraded sink: %w", err)
	}

	scanner := bufio.NewScanner(s.file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	rows := make([]*AuditLog, 0, 128)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		row := &AuditLog{}
		if err := json.Unmarshal([]byte(line), row); err != nil {
			return fmt.Errorf("decode degraded row: %w", err)
		}
		rows = append(rows, row)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan degraded sink: %w", err)
	}

	sort.Slice(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.Ts.Equal(right.Ts) {
			return left.ID < right.ID
		}
		return left.Ts.Before(right.Ts)
	})

	for _, row := range rows {
		if row == nil {
			continue
		}
		if err := apply(row); err != nil {
			return err
		}
	}

	if err := s.file.Truncate(0); err != nil {
		return fmt.Errorf("truncate degraded sink: %w", err)
	}
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind degraded sink: %w", err)
	}
	return s.file.Sync()
}

func (s *jsonlDegradedSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return nil
	}
	err := s.file.Close()
	s.file = nil
	return err
}
