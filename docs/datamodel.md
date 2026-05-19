# activity-heatmap Data Model Specification

## Overview

This document defines the file-based data model for `activity-heatmap`.

The tool writes four files per channel:

```text
outputs/<channel-slug>/
  channel.json
  videos.jsonl
  data.json
  heatmap.html
```

Only `channel.json`, `videos.jsonl`, and `data.json` are covered here. `heatmap.html` is a rendered output derived from `data.json`.

## Common Rules

All timestamps must be ISO 8601 strings.

YouTube API timestamps are stored in UTC when returned by the API.

Generated timestamps may use the system timezone.

All JSON files must include `schemaVersion`.

The initial schema version is:

```json
1
```

## channel.json

`channel.json` stores resolved YouTube channel metadata.

It is regenerated on every run.

### Shape

```json
{
  "schemaVersion": 1,
  "source": "youtube-data-api",
  "input": {
    "url": "https://www.youtube.com/@hololive"
  },
  "channel": {
    "id": "UC...",
    "title": "hololive ホロライブ - VTuber Group",
    "handle": "@hololive",
    "uploadsPlaylistId": "UU..."
  },
  "fetchedAt": "2026-05-19T10:30:00+09:00"
}
```

### Fields

`schemaVersion`: integer. Data schema version.

`source`: string. Must be `youtube-data-api`.

`input.url`: string. Original channel URL passed to the command.

`channel.id`: string. YouTube channel ID.

`channel.title`: string. Channel title returned by the YouTube API.

`channel.handle`: string or null. Channel handle returned by the YouTube API, including the leading `@` when available.

`channel.uploadsPlaylistId`: string. Uploads playlist ID for the channel.

`fetchedAt`: string. Timestamp when the channel metadata was fetched.

## videos.jsonl

`videos.jsonl` stores raw video records as JSON Lines.

Each line is one video record.

The file stores all fetched videos that have `liveStreamingDetails.actualEndTime`.

The file is append-only during normal execution.

Line order has no semantic meaning.

### Shape

```json
{
  "schemaVersion": 1,
  "videoId": "abc123",
  "title": "Example Live",
  "publishedAt": "2026-01-01T12:00:00Z",
  "liveStreamingDetails": {
    "scheduledStartTime": "2026-01-01T12:00:00Z",
    "actualStartTime": "2026-01-01T12:03:00Z",
    "actualEndTime": "2026-01-01T14:10:00Z"
  },
  "fetchedAt": "2026-05-19T10:30:00+09:00"
}
```

### Fields

`schemaVersion`: integer. Data schema version.

`videoId`: string. YouTube video ID. This is the primary key for the record.

`title`: string. Video title.

`publishedAt`: string. Video publish timestamp from YouTube.

`liveStreamingDetails`: object. Live streaming metadata.

`liveStreamingDetails.scheduledStartTime`: string or null. Scheduled live start time when available.

`liveStreamingDetails.actualStartTime`: string or null. Actual live start time when available.

`liveStreamingDetails.actualEndTime`: string. Actual live end time. Required for saved records.

`fetchedAt`: string. Timestamp when the record was fetched.

### Constraints

`videoId` must be unique within `videos.jsonl`.

Records without `liveStreamingDetails.actualEndTime` must not be saved.

Records may have `actualEndTime` without `actualStartTime`, but such records are not included in heatmap aggregation.

Existing records are not modified during normal incremental runs.

`--refresh` may replace the entire file.

## data.json

`data.json` stores the generated aggregation result.

It is regenerated on every run from `videos.jsonl`.

### Shape

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-05-19T10:35:00+09:00",
  "timezone": "Asia/Tokyo",
  "channel": {
    "id": "UC...",
    "title": "hololive ホロライブ - VTuber Group",
    "handle": "@hololive"
  },
  "summary": {
    "totalSavedVideos": 300,
    "liveVideos": 300
  },
  "heatmap": [
    { "weekday": "Mon", "hour": 0, "count": 0 },
    { "weekday": "Mon", "hour": 1, "count": 2 }
  ]
}
```

### Fields

`schemaVersion`: integer. Data schema version.

`generatedAt`: string. Timestamp when the aggregation was generated.

`timezone`: string. System timezone used for aggregation.

`channel.id`: string. YouTube channel ID.

`channel.title`: string. YouTube channel title.

`channel.handle`: string or null. YouTube channel handle, including the leading `@` when available.

`summary.totalSavedVideos`: integer. Number of records in `videos.jsonl`.

`summary.liveVideos`: integer. Number of records included in heatmap aggregation.

`heatmap`: array. Weekday/hour aggregation cells.

## Heatmap Cell

Each `heatmap` item represents one weekday/hour cell.

### Shape

```json
{
  "weekday": "Mon",
  "hour": 21,
  "count": 12
}
```

### Fields

`weekday`: string. One of:

```text
Mon
Tue
Wed
Thu
Fri
Sat
Sun
```

`hour`: integer. Hour of day in the aggregation timezone.

Allowed range:

```text
0-23
```

`count`: integer. Number of live streams whose `actualStartTime` falls into that weekday/hour cell.

### Constraints

`data.json` should contain all 168 cells:

```text
7 weekdays × 24 hours
```

Cells with no live streams must be present with `count: 0`.

Weekday order should be:

```text
Mon, Tue, Wed, Thu, Fri, Sat, Sun
```

Hour order should be ascending from `0` to `23` within each weekday.

## Aggregation Rules

Only records with both fields are included:

```text
liveStreamingDetails.actualStartTime
liveStreamingDetails.actualEndTime
```

For each included record:

1. Parse `liveStreamingDetails.actualStartTime`.
2. Convert it to the system timezone.
3. Extract weekday.
4. Extract hour of day.
5. Increment the matching heatmap cell by `1`.

`actualEndTime` is used only to confirm that the live stream has completed.

Stream duration is not used.

## File Ownership

`channel.json` is generated metadata.

`videos.jsonl` is raw local data and should be preserved across runs.

`data.json` is generated aggregation output.

`heatmap.html` is generated visualization output.

## Compatibility

Readers must ignore unknown fields.

Writers should preserve the schema shapes defined in this document.

Future schema changes must increment `schemaVersion`.
