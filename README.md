# activity-heatmap

Generate a weekday/hour heatmap from a public YouTube channel's live stream archives.

Point it at a channel URL and it writes a self-contained HTML report you can open in a browser to see when that channel tends to go live.

## What you get

For every run, the tool writes four files under `outputs/<channel>/`:

```
outputs/
  hololive/
    channel.json   resolved channel metadata
    videos.jsonl   raw video records (preserved across runs)
    data.json      aggregated heatmap data
    heatmap.html   the report you open in a browser
```

`heatmap.html` shows the heatmap and includes an interactive date-range slider so you can focus on recent activity (e.g., last 1 / 3 / 6 / 12 months) without re-running the command.

## Prerequisites

- A YouTube Data API v3 key, exported as `YOUTUBE_API_KEY`.
- The `activity-heatmap` binary (build with `mise run build`; the binary lands at `bin/activity-heatmap`).

## Usage

```sh
export YOUTUBE_API_KEY=your-key-here
activity-heatmap https://www.youtube.com/@hololive
```

Then open `outputs/hololive/heatmap.html` in your browser.

### Supported URL formats

```
https://www.youtube.com/@handle
https://youtube.com/@handle
https://www.youtube.com/channel/UC...
```

Other formats (`/c/...`, `/user/...`, etc.) are not supported.

### Options

| Option | Description |
| --- | --- |
| `--refresh` | Rebuild raw data from scratch. The existing `videos.jsonl` is replaced only after the new fetch succeeds. |
| `--output <dir>` | Base output directory (default `outputs`). |
| `--version` | Print version and exit. |
| `--help` | Print usage and exit. |

### First run

The first time you run the tool for a channel it fetches every public video and saves the ones with completed live streams:

```sh
$ activity-heatmap https://www.youtube.com/@hololive
resolved channel: hololive ホロライブ - VTuber Group (@hololive)
fetching videos from YouTube Data API...
saved raw data: outputs/hololive/videos.jsonl
generated: outputs/hololive/data.json
generated: outputs/hololive/heatmap.html
```

### Subsequent runs

Later runs reuse the cached `videos.jsonl` and only fetch videos that are newer than what's already saved:

```sh
$ activity-heatmap https://www.youtube.com/@hololive
resolved channel: hololive ホロライブ - VTuber Group (@hololive)
loaded raw data: outputs/hololive/videos.jsonl
fetching new videos...
appended 12 new records
generated: outputs/hololive/data.json
generated: outputs/hololive/heatmap.html
```

If nothing new is found, the report is still regenerated:

```sh
$ activity-heatmap https://www.youtube.com/@hololive
resolved channel: hololive ホロライブ - VTuber Group (@hololive)
loaded raw data: outputs/hololive/videos.jsonl
no new videos found
generated: outputs/hololive/data.json
generated: outputs/hololive/heatmap.html
```

### Rebuilding from scratch

Use `--refresh` to discard the cached raw data and re-fetch the full history:

```sh
activity-heatmap https://www.youtube.com/@hololive --refresh
```

If the refresh fails partway, your existing `videos.jsonl` is left untouched.

### Custom output directory

```sh
activity-heatmap https://www.youtube.com/@hololive --output reports
# → reports/hololive/{channel.json,videos.jsonl,data.json,heatmap.html}
```

## Reading the heatmap

- **X axis**: hour of day (0–23).
- **Y axis**: weekday (Mon–Sun).
- **Cell value**: number of live streams the channel started in that weekday/hour slot.
- **Timezone**: your system timezone. The resolved name is shown in the report header and saved in `data.json`.

The date-range slider above the chart filters the heatmap to a subset of the channel's history. Preset buttons jump to common windows; drag the slider handles for any range you like.

Only completed live streams are counted. Currently-live, scheduled, members-only, private, or unavailable videos are excluded.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `1` | Runtime error (missing API key, channel not found, write failure, …). |
| `2` | Invalid command-line usage (missing URL, unknown option, unsupported URL format). |

## Limitations

The MVP intentionally does one thing well:

- One channel per command (no comparison, no batching).
- System timezone only (no `--timezone` option yet).
- Counts stream starts, not durations.
- `heatmap.html` loads ECharts from a CDN, so an internet connection is needed to view the chart.
- URL formats `/c/...` and `/user/...` are not supported.
