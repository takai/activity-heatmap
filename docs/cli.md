# activity-heatmap CLI Specification

## Overview

`activity-heatmap` generates a weekday/hour heatmap from a public YouTube channel’s live stream archives.

The tool accepts one YouTube channel URL, fetches public video metadata through the YouTube Data API, stores raw video records as JSON Lines, and writes a static HTML report.

## Command

```sh
activity-heatmap <youtube-channel-url>
```

Example:

```sh
activity-heatmap https://www.youtube.com/@hololive
```

## Authentication

The tool reads the YouTube Data API key from the environment variable:

```sh
YOUTUBE_API_KEY
```

If `YOUTUBE_API_KEY` is missing or empty, the command must fail before making any API request.

There is no `--api-key` option.

## Arguments

### `<youtube-channel-url>`

Required.

The YouTube channel URL to analyze.

Supported formats:

```text
https://www.youtube.com/@handle
https://youtube.com/@handle
https://www.youtube.com/channel/UC...
```

Unsupported formats must result in an error.

## Options

### `--refresh`

Rebuild raw data from the YouTube Data API.

Default behavior reuses `videos.jsonl` when it exists and only appends newly discovered videos.

When `--refresh` is specified, the tool ignores the existing `videos.jsonl`, fetches channel videos again, and replaces `videos.jsonl` only after the new fetch succeeds.

Example:

```sh
activity-heatmap https://www.youtube.com/@hololive --refresh
```

### `--output <dir>`

Optional.

Base output directory.

Default:

```text
outputs
```

Example:

```sh
activity-heatmap https://www.youtube.com/@hololive --output reports
```

This writes:

```text
reports/hololive/
  channel.json
  videos.jsonl
  data.json
  heatmap.html
```

### `--help`

Print usage information and exit.

### `--version`

Print version information and exit.

## Output Directory

The final output directory is:

```text
<output>/<channel-slug>/
```

The channel slug is resolved from YouTube channel metadata:

1. Use the channel handle without the leading `@`.
2. If no handle is available, use the channel ID.

Example:

```text
https://www.youtube.com/@hololive
→ outputs/hololive/
```

## Generated Files

The tool writes the following files:

```text
channel.json
videos.jsonl
data.json
heatmap.html
```

### `channel.json`

Contains resolved channel metadata.

Example shape:

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

### `videos.jsonl`

Contains raw video records.

Each line is one video record.

The file stores all fetched videos that have `liveStreamingDetails.actualEndTime`.

The file is append-only during normal execution.

Example line:

```json
{"schemaVersion":1,"videoId":"abc123","title":"Example Live","publishedAt":"2026-01-01T12:00:00Z","liveStreamingDetails":{"scheduledStartTime":"2026-01-01T12:00:00Z","actualStartTime":"2026-01-01T12:03:00Z","actualEndTime":"2026-01-01T14:10:00Z"},"fetchedAt":"2026-05-19T10:30:00+09:00"}
```

### `data.json`

Contains aggregated heatmap data.

This file is regenerated on every run and may be overwritten.

Example shape:

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

### `heatmap.html`

Static HTML report.

Uses Apache ECharts loaded from a CDN.

This file is regenerated on every run and may be overwritten.

## Default Execution Behavior

When `videos.jsonl` does not exist:

1. Resolve the channel.
2. Fetch the uploads playlist.
3. Fetch video details.
4. Save eligible video records to `videos.jsonl`.
5. Generate `channel.json`.
6. Generate `data.json`.
7. Generate `heatmap.html`.

When `videos.jsonl` already exists:

1. Resolve the channel.
2. Read existing `videoId` values from `videos.jsonl`.
3. Fetch uploads in newest-first order.
4. Stop when an existing `videoId` is reached.
5. Append only newly discovered eligible records to `videos.jsonl`.
6. Regenerate `channel.json`.
7. Regenerate `data.json`.
8. Regenerate `heatmap.html`.

Line order in `videos.jsonl` has no semantic meaning.

## Refresh Behavior

When `--refresh` is specified:

1. Resolve the channel.
2. Fetch the uploads playlist from the beginning.
3. Fetch video details.
4. Write the new raw data to a temporary file.
5. Replace `videos.jsonl` only after the fetch succeeds.
6. Regenerate `channel.json`.
7. Regenerate `data.json`.
8. Regenerate `heatmap.html`.

If refresh fails, the existing `videos.jsonl` must not be destroyed.

## Video Save Rules

The tool saves a video record only when:

```text
liveStreamingDetails.actualEndTime exists
```

Videos without `actualEndTime` are skipped.

This excludes currently live, scheduled, incomplete, or otherwise unfinished streams.

## Live Stream Aggregation Rules

A saved video is included in the heatmap when:

```text
liveStreamingDetails.actualStartTime exists
liveStreamingDetails.actualEndTime exists
```

Aggregation uses `actualStartTime`.

For each included video:

1. Convert `actualStartTime` to the system timezone.
2. Extract weekday.
3. Extract hour of day.
4. Increment that weekday/hour cell by `1`.

Duration does not affect the heatmap.

## Timezone

The default timezone is the system timezone.

There is no timezone option in the MVP.

The resolved timezone name should be written to `data.json` when available.

## Visualization Layout

`heatmap.html` must render:

```text
X axis: hour of day, 0-23
Y axis: weekday
Cell value: number of live streams started in that weekday/hour slot
```

The page should include:

```text
Channel title
Channel handle or ID
Timezone
Total live count
Generated timestamp
Heatmap chart
```

## Exit Codes

### `0`

Success.

### `1`

General runtime error.

Examples:

```text
Missing API key
Invalid channel URL
Channel not found
API request failed without usable local data
Invalid local data file
Failed to write output files
```

### `2`

Invalid command-line usage.

Examples:

```text
Missing channel URL
Unknown option
Too many arguments
Unsupported URL format
```

## Error Messages

Error messages should be concise and actionable.

Examples:

```text
error: YOUTUBE_API_KEY is required
error: unsupported YouTube channel URL: https://www.youtube.com/c/example
error: failed to resolve channel
error: failed to fetch videos and no cached videos.jsonl exists
error: invalid videos.jsonl: line 12
```

## Console Output

Normal output should be brief.

Example first run:

```text
resolved channel: hololive ホロライブ - VTuber Group (@hololive)
fetching videos from YouTube Data API...
saved raw data: outputs/hololive/videos.jsonl
generated: outputs/hololive/data.json
generated: outputs/hololive/heatmap.html
```

Example cached/incremental run:

```text
resolved channel: hololive ホロライブ - VTuber Group (@hololive)
loaded raw data: outputs/hololive/videos.jsonl
fetching new videos...
appended 12 new records
generated: outputs/hololive/data.json
generated: outputs/hololive/heatmap.html
```

Example no new videos:

```text
resolved channel: hololive ホロライブ - VTuber Group (@hololive)
loaded raw data: outputs/hololive/videos.jsonl
no new videos found
generated: outputs/hololive/data.json
generated: outputs/hololive/heatmap.html
```

## Non-Goals

The MVP CLI does not support:

```text
Multiple channels in one command
OAuth authentication
API key command-line option
Explicit timezone option
Duration-based aggregation
Channel comparison
Scraping or crawling YouTube HTML
Self-contained offline HTML
Old YouTube URL formats such as /c/... or /user/...
```
