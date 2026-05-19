package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadVideos reads videos.jsonl. Returns (nil, nil) if the file does not exist.
func ReadVideos(path string) ([]Video, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var videos []Video
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var v Video
		if err := json.Unmarshal(line, &v); err != nil {
			return nil, fmt.Errorf("invalid videos.jsonl: line %d: %w", lineNo, err)
		}
		videos = append(videos, v)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return videos, nil
}

// AppendVideos appends video records to videos.jsonl, creating the file if needed.
func AppendVideos(path string, videos []Video) error {
	if len(videos) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeVideoLines(f, videos)
}

// WriteVideos writes videos.jsonl atomically by writing to a temp file and renaming.
func WriteVideos(path string, videos []Video) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".videos-*.jsonl.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleaned := false
	defer func() {
		if !cleaned {
			_ = os.Remove(tmpPath)
		}
	}()
	if err := writeVideoLines(tmp, videos); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleaned = true
	return nil
}

func writeVideoLines(w io.Writer, videos []Video) error {
	bw := bufio.NewWriter(w)
	enc := json.NewEncoder(bw)
	enc.SetEscapeHTML(false)
	for _, v := range videos {
		if err := enc.Encode(v); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// WriteJSON writes any value as pretty JSON, atomically.
func WriteJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".out-*.json.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleaned := false
	defer func() {
		if !cleaned {
			_ = os.Remove(tmpPath)
		}
	}()
	enc := json.NewEncoder(tmp)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleaned = true
	return nil
}

// WriteFile writes raw bytes atomically.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".out-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleaned := false
	defer func() {
		if !cleaned {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleaned = true
	return nil
}
