// Package store defines the persistent data model for activity-heatmap and
// helpers for reading and writing it.
package store

import "time"

const SchemaVersion = 1

type ChannelFile struct {
	SchemaVersion int           `json:"schemaVersion"`
	Source        string        `json:"source"`
	Input         ChannelInput  `json:"input"`
	Channel       ChannelInfo   `json:"channel"`
	FetchedAt     time.Time     `json:"fetchedAt"`
}

type ChannelInput struct {
	URL string `json:"url"`
}

type ChannelInfo struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Handle            *string `json:"handle"`
	UploadsPlaylistID string  `json:"uploadsPlaylistId"`
}

type Video struct {
	SchemaVersion        int                  `json:"schemaVersion"`
	VideoID              string               `json:"videoId"`
	Title                string               `json:"title"`
	PublishedAt          time.Time            `json:"publishedAt"`
	LiveStreamingDetails LiveStreamingDetails `json:"liveStreamingDetails"`
	FetchedAt            time.Time            `json:"fetchedAt"`
}

type LiveStreamingDetails struct {
	ScheduledStartTime *time.Time `json:"scheduledStartTime"`
	ActualStartTime    *time.Time `json:"actualStartTime"`
	ActualEndTime      time.Time  `json:"actualEndTime"`
}

type DataFile struct {
	SchemaVersion int             `json:"schemaVersion"`
	GeneratedAt   time.Time       `json:"generatedAt"`
	Timezone      string          `json:"timezone"`
	Channel       DataFileChannel `json:"channel"`
	Summary       DataFileSummary `json:"summary"`
	Heatmap       []HeatmapCell   `json:"heatmap"`
}

type DataFileChannel struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Handle *string `json:"handle"`
}

type DataFileSummary struct {
	TotalSavedVideos int `json:"totalSavedVideos"`
	LiveVideos       int `json:"liveVideos"`
}

type HeatmapCell struct {
	Weekday string `json:"weekday"`
	Hour    int    `json:"hour"`
	Count   int    `json:"count"`
}
