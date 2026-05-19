# activity-heatmap PRD

## Overview

`activity-heatmap` is a command-line tool that generates a weekday/hour heatmap from a public YouTube channel’s live stream archives.

A user provides a YouTube channel URL, and the tool outputs a static HTML file showing when that channel tends to start live streams.

## Problem

YouTube channels often have recurring live stream patterns, but YouTube does not provide a simple way to see those patterns across a channel’s history.

Users who follow, analyze, or manage YouTube channels may want to answer questions such as:

```text
What day of the week does this channel usually stream?
What time of day does this channel usually go live?
Does this channel have a consistent streaming pattern?
```

Today, answering these questions requires manually checking many archived live streams or writing custom scripts.

## Goal

Provide a simple CLI that turns a YouTube channel URL into a reusable local heatmap report.

The first version should focus on one clear job:

```text
Given one public YouTube channel, generate a weekday/hour heatmap of its live stream start times.
```

## Target Users

Primary users:

```text
Fans who want to understand a channel’s live schedule patterns
Researchers or analysts studying YouTube activity patterns
Creators who want to inspect their own public archive
Developers who prefer local, file-based tools
```

Secondary users:

```text
People building datasets from YouTube channel activity
People comparing streaming habits across channels manually
```

## User Story

As a user, I want to run:

```sh
activity-heatmap https://www.youtube.com/@hololive
```

and get:

```text
outputs/hololive/heatmap.html
```

so that I can open the file in a browser and quickly see which weekdays and hours the channel tends to start live streams.

## Core Requirements

The tool must accept a single YouTube channel URL.

The tool must use the YouTube Data API with an API key provided by the `YOUTUBE_API_KEY` environment variable.

The tool must fetch public video metadata for the target channel.

The tool must save raw video records as JSON Lines.

The tool must reuse existing raw data when available.

The tool must append newly discovered videos to the raw data file.

The tool must generate a static HTML heatmap.

The tool must generate a machine-readable aggregated data file.

The tool must support full refresh through an explicit `--refresh` option.

## MVP Behavior

Default command:

```sh
activity-heatmap https://www.youtube.com/@hololive
```

Expected output:

```text
outputs/
  hololive/
    channel.json
    videos.jsonl
    data.json
    heatmap.html
```

If `videos.jsonl` does not exist, the tool fetches available public videos and creates it.

If `videos.jsonl` already exists, the tool fetches from newest to oldest and stops when it reaches an existing `videoId`.

Newly discovered records are appended to `videos.jsonl`.

`data.json` and `heatmap.html` are regenerated on every run.

`videos.jsonl` is replaced only when `--refresh` is specified.

## Live Stream Definition

For the MVP, a live stream is a saved video record with both:

```text
liveStreamingDetails.actualStartTime
liveStreamingDetails.actualEndTime
```

Videos without `actualEndTime` are not saved.

The heatmap uses `actualStartTime` only.

## Aggregation

The tool converts each live stream’s `actualStartTime` to the system timezone.

It then groups live streams by:

```text
weekday
hour of day
```

Each live stream contributes `1` to the corresponding weekday/hour cell.

Duration-based aggregation is not part of the MVP.

## Visualization

The generated HTML must show a heatmap with:

```text
X axis: hour of day, 0-23
Y axis: weekday
Cell value: number of live streams started in that slot
```

The MVP uses Apache ECharts from a CDN.

The HTML output should be readable without requiring a local web server.

## Supported Inputs

The MVP supports:

```text
https://www.youtube.com/@handle
https://youtube.com/@handle
https://www.youtube.com/channel/UC...
```

Other URL formats are out of scope.

## Output Directory Naming

The output directory is based on resolved channel metadata.

Use the channel handle without `@` when available.

If no handle is available, use the channel ID.

Example:

```text
https://www.youtube.com/@hololive
→ outputs/hololive/
```

## Success Criteria

The MVP is successful if a user can:

```text
Install or build the CLI
Set YOUTUBE_API_KEY
Run the command with a supported YouTube channel URL
Open heatmap.html
Understand the channel’s live start-time pattern
Run the command again and reuse existing data
Run with --refresh to rebuild the raw dataset
```

## Non-Goals

The MVP does not support:

```text
Multiple channels in one command
Channel comparison
Duration-based heatmap aggregation
Private, deleted, members-only, or unavailable videos
OAuth authentication
YouTube scraping
Offline bundled chart assets
Self-contained HTML output
Old YouTube URL formats such as /c/... or /user/...
GUI application
Hosted web service
```

## Future Opportunities

Potential future features include:

```text
Multiple channel input
Channel comparison reports
Explicit --timezone option
--no-fetch mode for regenerating reports from local data only
Duration-based heatmap mode
CSV export
Markdown report output
Self-contained offline HTML
Support for /c/... and /user/... URL formats
Config file support
```

## Product Principle

The tool should stay local, predictable, and file-based.

Raw data should be preserved in a form that is easy to inspect, diff, and reuse. HTML and aggregate files should be treated as generated outputs.
