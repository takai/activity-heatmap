package aggregate

import (
	"testing"
	"time"

	"github.com/takai/activity-heatmap/internal/store"
)

func TestBuild_CellsAndCounts(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	// 2026-05-19 is a Tuesday.
	mkVideo := func(id string, start time.Time) store.Video {
		return store.Video{
			VideoID: id,
			LiveStreamingDetails: store.LiveStreamingDetails{
				ActualStartTime: new(start),
				ActualEndTime:   start.Add(2 * time.Hour),
			},
		}
	}

	videos := []store.Video{
		mkVideo("a", time.Date(2026, 5, 19, 21, 30, 0, 0, loc)),     // Tue 21
		mkVideo("b", time.Date(2026, 5, 19, 21, 45, 0, 0, loc)),     // Tue 21
		mkVideo("c", time.Date(2026, 5, 19, 12, 5, 0, 0, time.UTC)), // 21:05 JST -> Tue 21
		mkVideo("d", time.Date(2026, 5, 18, 9, 0, 0, 0, loc)),       // Mon 9
		{ // no ActualStartTime - skipped
			VideoID: "skipped",
			LiveStreamingDetails: store.LiveStreamingDetails{
				ActualEndTime: time.Now(),
			},
		},
	}

	ch := store.ChannelInfo{ID: "UC1", Title: "T", Handle: new("@h")}
	df := Build(ch, videos, loc, time.Date(2026, 5, 19, 22, 0, 0, 0, loc))

	if df.Summary.TotalSavedVideos != len(videos) {
		t.Errorf("TotalSavedVideos: got %d, want %d", df.Summary.TotalSavedVideos, len(videos))
	}
	if df.Summary.LiveVideos != 4 {
		t.Errorf("LiveVideos: got %d, want 4", df.Summary.LiveVideos)
	}
	if len(df.Heatmap) != 168 {
		t.Fatalf("want 168 cells, got %d", len(df.Heatmap))
	}
	if df.Heatmap[0].Weekday != "Mon" || df.Heatmap[0].Hour != 0 {
		t.Errorf("first cell unexpected: %+v", df.Heatmap[0])
	}
	if df.Heatmap[167].Weekday != "Sun" || df.Heatmap[167].Hour != 23 {
		t.Errorf("last cell unexpected: %+v", df.Heatmap[167])
	}

	find := func(wd string, hour int) store.HeatmapCell {
		for _, c := range df.Heatmap {
			if c.Weekday == wd && c.Hour == hour {
				return c
			}
		}
		t.Fatalf("missing cell %s %d", wd, hour)
		return store.HeatmapCell{}
	}
	if c := find("Tue", 21); c.Count != 3 {
		t.Errorf("Tue 21 = %d, want 3", c.Count)
	}
	if c := find("Mon", 9); c.Count != 1 {
		t.Errorf("Mon 9 = %d, want 1", c.Count)
	}
	if c := find("Wed", 0); c.Count != 0 {
		t.Errorf("Wed 0 = %d, want 0", c.Count)
	}
	if df.Timezone != "Asia/Tokyo" {
		t.Errorf("Timezone = %q, want Asia/Tokyo", df.Timezone)
	}
}
