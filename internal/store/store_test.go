package store

import (
	"path/filepath"
	"testing"
	"time"
)

func mkVideo(id string) Video {
	t := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	start := t.Add(3 * time.Minute)
	return Video{
		SchemaVersion: SchemaVersion,
		VideoID:       id,
		Title:         "Example " + id,
		PublishedAt:   t,
		LiveStreamingDetails: LiveStreamingDetails{
			ActualStartTime: &start,
			ActualEndTime:   t.Add(2 * time.Hour),
		},
		FetchedAt: t,
	}
}

func TestReadVideos_Missing(t *testing.T) {
	videos, err := ReadVideos(filepath.Join(t.TempDir(), "videos.jsonl"))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if videos != nil {
		t.Fatalf("expected nil, got %v", videos)
	}
}

func TestAppendAndRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "videos.jsonl")
	if err := AppendVideos(path, []Video{mkVideo("a"), mkVideo("b")}); err != nil {
		t.Fatal(err)
	}
	if err := AppendVideos(path, []Video{mkVideo("c")}); err != nil {
		t.Fatal(err)
	}
	got, err := ReadVideos(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d, want 3", len(got))
	}
	if got[0].VideoID != "a" || got[2].VideoID != "c" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestWriteVideos_Atomic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "videos.jsonl")
	if err := WriteVideos(path, []Video{mkVideo("a")}); err != nil {
		t.Fatal(err)
	}
	if err := WriteVideos(path, []Video{mkVideo("x"), mkVideo("y")}); err != nil {
		t.Fatal(err)
	}
	got, err := ReadVideos(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].VideoID != "x" {
		t.Fatalf("unexpected: %+v", got)
	}
}
