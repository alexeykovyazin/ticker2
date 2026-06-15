//go:build windows

package logwriter

import (
	"fmt"
	"os"
)

type TextLog struct {
	file *os.File
}

func OpenText(path string) (*TextLog, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open text log %q: %w", path, err)
	}
	return &TextLog{file: f}, nil
}

func (t *TextLog) WriteLine(line string) error {
	_, err := t.file.WriteString(line)
	if err != nil {
		return fmt.Errorf("write text log: %w", err)
	}
	return nil
}

func (t *TextLog) Close() error {
	if t.file == nil {
		return nil
	}
	return t.file.Close()
}
