// Package app orchestrates the activity-heatmap end-to-end run.
package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/takai/activity-heatmap/internal/aggregate"
	"github.com/takai/activity-heatmap/internal/channelurl"
	"github.com/takai/activity-heatmap/internal/render"
	"github.com/takai/activity-heatmap/internal/store"
	"github.com/takai/activity-heatmap/internal/youtube"
)

// Config holds the inputs for a single run.
type Config struct {
	ChannelURL string
	OutputDir  string
	Refresh    bool
	APIKey     string
	// Now returns the current time; defaults to time.Now.
	Now func() time.Time
	// Location is the aggregation timezone; defaults to time.Local.
	Location *time.Location
	// Stdout is the destination for progress messages.
	Stdout io.Writer
}

// Run executes one activity-heatmap invocation.
func Run(ctx context.Context, cfg Config) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("YOUTUBE_API_KEY is required")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Location == nil {
		cfg.Location = time.Local
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "outputs"
	}
	if cfg.Stdout == nil {
		cfg.Stdout = io.Discard
	}

	ref, err := channelurl.Parse(cfg.ChannelURL)
	if err != nil {
		return fmt.Errorf("unsupported YouTube channel URL: %s", cfg.ChannelURL)
	}

	client := youtube.NewClient(cfg.APIKey)
	rc, err := client.ResolveChannel(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to resolve channel: %w", err)
	}

	slug := channelSlug(rc, ref)
	channelDir := filepath.Join(cfg.OutputDir, slug)
	videosPath := filepath.Join(channelDir, "videos.jsonl")

	handleLabel := rc.ID
	if rc.Handle != nil {
		handleLabel = *rc.Handle
	}
	fmt.Fprintf(cfg.Stdout, "resolved channel: %s (%s)\n", rc.Title, handleLabel)

	existing, err := store.ReadVideos(videosPath)
	if err != nil {
		return fmt.Errorf("invalid videos.jsonl: %w", err)
	}
	hasCache := existing != nil
	if cfg.Refresh {
		existing = nil
		hasCache = false
	}

	known := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		known[v.VideoID] = struct{}{}
	}

	if hasCache {
		fmt.Fprintf(cfg.Stdout, "loaded raw data: %s\n", videosPath)
		fmt.Fprintln(cfg.Stdout, "fetching new videos...")
	} else if cfg.Refresh {
		fmt.Fprintln(cfg.Stdout, "refreshing raw data from YouTube Data API...")
	} else {
		fmt.Fprintln(cfg.Stdout, "fetching videos from YouTube Data API...")
	}

	var newIDs []string
	err = client.IterateUploads(ctx, rc.UploadsPlaylistID, func(item youtube.UploadsItem) (bool, error) {
		if _, ok := known[item.VideoID]; ok {
			return true, nil
		}
		newIDs = append(newIDs, item.VideoID)
		return false, nil
	})
	if err != nil {
		if hasCache {
			return fmt.Errorf("failed to fetch new videos: %w", err)
		}
		return fmt.Errorf("failed to fetch videos and no cached videos.jsonl exists: %w", err)
	}

	newRecords, err := fetchEligibleVideos(ctx, client, newIDs, cfg.Now())
	if err != nil {
		return fmt.Errorf("failed to fetch video details: %w", err)
	}

	if cfg.Refresh {
		if err := store.WriteVideos(videosPath, newRecords); err != nil {
			return fmt.Errorf("failed to write videos.jsonl: %w", err)
		}
		fmt.Fprintf(cfg.Stdout, "saved raw data: %s\n", videosPath)
	} else if hasCache {
		if len(newRecords) == 0 {
			fmt.Fprintln(cfg.Stdout, "no new videos found")
		} else {
			if err := store.AppendVideos(videosPath, newRecords); err != nil {
				return fmt.Errorf("failed to append to videos.jsonl: %w", err)
			}
			fmt.Fprintf(cfg.Stdout, "appended %d new records\n", len(newRecords))
		}
	} else {
		if err := store.WriteVideos(videosPath, newRecords); err != nil {
			return fmt.Errorf("failed to write videos.jsonl: %w", err)
		}
		fmt.Fprintf(cfg.Stdout, "saved raw data: %s\n", videosPath)
	}

	all := append([]store.Video{}, existing...)
	all = append(all, newRecords...)

	channelInfo := store.ChannelInfo{
		ID:                rc.ID,
		Title:             rc.Title,
		Handle:            rc.Handle,
		UploadsPlaylistID: rc.UploadsPlaylistID,
	}
	channelFile := store.ChannelFile{
		SchemaVersion: store.SchemaVersion,
		Source:        "youtube-data-api",
		Input:         store.ChannelInput{URL: cfg.ChannelURL},
		Channel:       channelInfo,
		FetchedAt:     cfg.Now(),
	}
	if err := store.WriteJSON(filepath.Join(channelDir, "channel.json"), channelFile); err != nil {
		return fmt.Errorf("failed to write channel.json: %w", err)
	}

	df := aggregate.Build(channelInfo, all, cfg.Location, cfg.Now())
	dataPath := filepath.Join(channelDir, "data.json")
	if err := store.WriteJSON(dataPath, df); err != nil {
		return fmt.Errorf("failed to write data.json: %w", err)
	}
	fmt.Fprintf(cfg.Stdout, "generated: %s\n", dataPath)

	htmlBytes, err := render.HTML(df)
	if err != nil {
		return fmt.Errorf("failed to render heatmap: %w", err)
	}
	htmlPath := filepath.Join(channelDir, "heatmap.html")
	if err := store.WriteFile(htmlPath, htmlBytes); err != nil {
		return fmt.Errorf("failed to write heatmap.html: %w", err)
	}
	fmt.Fprintf(cfg.Stdout, "generated: %s\n", htmlPath)

	return nil
}

func channelSlug(rc youtube.ResolvedChannel, ref channelurl.Ref) string {
	if rc.Handle != nil && *rc.Handle != "" {
		h := *rc.Handle
		if len(h) > 0 && h[0] == '@' {
			h = h[1:]
		}
		if h != "" {
			return h
		}
	}
	if ref.Kind == channelurl.KindHandle && ref.Value != "" {
		return ref.Value
	}
	return rc.ID
}

func fetchEligibleVideos(ctx context.Context, c *youtube.Client, ids []string, fetchedAt time.Time) ([]store.Video, error) {
	var out []store.Video
	for start := 0; start < len(ids); start += 50 {
		end := min(start+50, len(ids))
		details, err := c.FetchVideos(ctx, ids[start:end])
		if err != nil {
			return nil, err
		}
		for _, d := range details {
			if d.ActualEndTime == nil {
				continue
			}
			out = append(out, store.Video{
				SchemaVersion: store.SchemaVersion,
				VideoID:       d.ID,
				Title:         d.Title,
				PublishedAt:   d.PublishedAt,
				LiveStreamingDetails: store.LiveStreamingDetails{
					ScheduledStartTime: d.ScheduledStartTime,
					ActualStartTime:    d.ActualStartTime,
					ActualEndTime:      *d.ActualEndTime,
				},
				FetchedAt: fetchedAt,
			})
		}
	}
	return out, nil
}
