// Package channelurl parses the YouTube channel URL formats supported by
// activity-heatmap.
package channelurl

import (
	"fmt"
	"net/url"
	"strings"
)

type Kind int

const (
	KindHandle Kind = iota + 1
	KindChannelID
)

type Ref struct {
	Kind  Kind
	Value string
}

func Parse(raw string) (Ref, error) {
	if raw == "" {
		return Ref{}, fmt.Errorf("empty URL")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Ref{}, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Ref{}, fmt.Errorf("unsupported scheme: %q", u.Scheme)
	}
	host := strings.ToLower(u.Host)
	if host != "www.youtube.com" && host != "youtube.com" && host != "m.youtube.com" {
		return Ref{}, fmt.Errorf("not a youtube.com URL: %q", u.Host)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return Ref{}, fmt.Errorf("missing channel path")
	}
	first := parts[0]
	switch {
	case strings.HasPrefix(first, "@"):
		handle := strings.TrimPrefix(first, "@")
		if handle == "" {
			return Ref{}, fmt.Errorf("empty channel handle")
		}
		return Ref{Kind: KindHandle, Value: handle}, nil
	case first == "channel":
		if len(parts) < 2 || parts[1] == "" {
			return Ref{}, fmt.Errorf("missing channel ID")
		}
		id := parts[1]
		if !strings.HasPrefix(id, "UC") {
			return Ref{}, fmt.Errorf("invalid channel ID: %q", id)
		}
		return Ref{Kind: KindChannelID, Value: id}, nil
	default:
		return Ref{}, fmt.Errorf("unsupported channel URL path: %q", u.Path)
	}
}
