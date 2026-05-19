// Package aggregate builds the heatmap data file from raw video records.
package aggregate

import (
	"time"

	"github.com/takai/activity-heatmap/internal/store"
)

var Weekdays = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

func weekdayIndex(t time.Time) int {
	// time.Weekday: Sunday=0..Saturday=6. Convert to Mon=0..Sun=6.
	wd := int(t.Weekday())
	return (wd + 6) % 7
}

// Build produces a DataFile from saved videos. The channel info, location and
// generation timestamp are provided by the caller.
func Build(ch store.ChannelInfo, videos []store.Video, loc *time.Location, generatedAt time.Time) store.DataFile {
	counts := make([][]int, 7)
	for i := range counts {
		counts[i] = make([]int, 24)
	}

	liveVideos := 0
	for _, v := range videos {
		if v.LiveStreamingDetails.ActualStartTime == nil {
			continue
		}
		// actualEndTime is required when stored, but be defensive.
		if v.LiveStreamingDetails.ActualEndTime.IsZero() {
			continue
		}
		t := v.LiveStreamingDetails.ActualStartTime.In(loc)
		counts[weekdayIndex(t)][t.Hour()]++
		liveVideos++
	}

	cells := make([]store.HeatmapCell, 0, 7*24)
	for wi, name := range Weekdays {
		for h := range 24 {
			cells = append(cells, store.HeatmapCell{
				Weekday: name,
				Hour:    h,
				Count:   counts[wi][h],
			})
		}
	}

	tzName, _ := generatedAt.In(loc).Zone()
	if locName := loc.String(); locName != "" {
		tzName = locName
	}

	return store.DataFile{
		SchemaVersion: store.SchemaVersion,
		GeneratedAt:   generatedAt,
		Timezone:      tzName,
		Channel: store.DataFileChannel{
			ID:     ch.ID,
			Title:  ch.Title,
			Handle: ch.Handle,
		},
		Summary: store.DataFileSummary{
			TotalSavedVideos: len(videos),
			LiveVideos:       liveVideos,
		},
		Heatmap: cells,
	}
}
