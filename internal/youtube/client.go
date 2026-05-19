// Package youtube is a minimal client for the parts of the YouTube Data API v3
// used by activity-heatmap.
package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/takai/activity-heatmap/internal/channelurl"
)

const defaultBaseURL = "https://www.googleapis.com/youtube/v3"

type Client struct {
	APIKey  string
	HTTP    *http.Client
	BaseURL string
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		BaseURL: defaultBaseURL,
	}
}

type ResolvedChannel struct {
	ID                string
	Title             string
	Handle            *string
	UploadsPlaylistID string
}

type apiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) do(ctx context.Context, endpoint string, params url.Values, out any) error {
	params.Set("key", c.APIKey)
	u := c.BaseURL + "/" + endpoint + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		var apiErr apiError
		if jerr := json.Unmarshal(body, &apiErr); jerr == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("youtube api: %s (%d)", apiErr.Error.Message, apiErr.Error.Code)
		}
		return fmt.Errorf("youtube api: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

type channelsResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title     string `json:"title"`
			CustomURL string `json:"customUrl"`
		} `json:"snippet"`
		ContentDetails struct {
			RelatedPlaylists struct {
				Uploads string `json:"uploads"`
			} `json:"relatedPlaylists"`
		} `json:"contentDetails"`
	} `json:"items"`
}

func (c *Client) ResolveChannel(ctx context.Context, ref channelurl.Ref) (ResolvedChannel, error) {
	params := url.Values{}
	params.Set("part", "snippet,contentDetails")
	switch ref.Kind {
	case channelurl.KindHandle:
		params.Set("forHandle", "@"+ref.Value)
	case channelurl.KindChannelID:
		params.Set("id", ref.Value)
	default:
		return ResolvedChannel{}, fmt.Errorf("unsupported channel ref")
	}
	var resp channelsResponse
	if err := c.do(ctx, "channels", params, &resp); err != nil {
		return ResolvedChannel{}, err
	}
	if len(resp.Items) == 0 {
		return ResolvedChannel{}, fmt.Errorf("channel not found")
	}
	it := resp.Items[0]
	rc := ResolvedChannel{
		ID:                it.ID,
		Title:             it.Snippet.Title,
		UploadsPlaylistID: it.ContentDetails.RelatedPlaylists.Uploads,
	}
	if h := strings.TrimSpace(it.Snippet.CustomURL); h != "" {
		if !strings.HasPrefix(h, "@") {
			h = "@" + h
		}
		rc.Handle = &h
	} else if ref.Kind == channelurl.KindHandle {
		h := "@" + ref.Value
		rc.Handle = &h
	}
	if rc.UploadsPlaylistID == "" {
		return ResolvedChannel{}, fmt.Errorf("channel has no uploads playlist")
	}
	return rc, nil
}

type playlistItemsResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Items         []struct {
		ContentDetails struct {
			VideoID string `json:"videoId"`
		} `json:"contentDetails"`
	} `json:"items"`
}

// UploadsItem is a single item from a channel's uploads playlist, in
// newest-first order.
type UploadsItem struct {
	VideoID string
}

// IterateUploads invokes fn for each upload, newest first. Returning stop=true
// halts pagination cleanly.
func (c *Client) IterateUploads(ctx context.Context, playlistID string, fn func(UploadsItem) (stop bool, err error)) error {
	pageToken := ""
	for {
		params := url.Values{}
		params.Set("part", "contentDetails")
		params.Set("playlistId", playlistID)
		params.Set("maxResults", "50")
		if pageToken != "" {
			params.Set("pageToken", pageToken)
		}
		var resp playlistItemsResponse
		if err := c.do(ctx, "playlistItems", params, &resp); err != nil {
			return err
		}
		for _, it := range resp.Items {
			if it.ContentDetails.VideoID == "" {
				continue
			}
			stop, err := fn(UploadsItem{VideoID: it.ContentDetails.VideoID})
			if err != nil {
				return err
			}
			if stop {
				return nil
			}
		}
		if resp.NextPageToken == "" {
			return nil
		}
		pageToken = resp.NextPageToken
	}
}

type VideoDetail struct {
	ID                 string
	Title              string
	PublishedAt        time.Time
	ScheduledStartTime *time.Time
	ActualStartTime    *time.Time
	ActualEndTime      *time.Time
}

type videosResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title       string    `json:"title"`
			PublishedAt time.Time `json:"publishedAt"`
		} `json:"snippet"`
		LiveStreamingDetails *struct {
			ScheduledStartTime *time.Time `json:"scheduledStartTime"`
			ActualStartTime    *time.Time `json:"actualStartTime"`
			ActualEndTime      *time.Time `json:"actualEndTime"`
		} `json:"liveStreamingDetails"`
	} `json:"items"`
}

// FetchVideos fetches details for up to 50 video IDs.
func (c *Client) FetchVideos(ctx context.Context, ids []string) ([]VideoDetail, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 50 {
		return nil, fmt.Errorf("FetchVideos: too many ids (%d), max 50", len(ids))
	}
	params := url.Values{}
	params.Set("part", "snippet,liveStreamingDetails")
	params.Set("id", strings.Join(ids, ","))
	params.Set("maxResults", "50")
	var resp videosResponse
	if err := c.do(ctx, "videos", params, &resp); err != nil {
		return nil, err
	}
	out := make([]VideoDetail, 0, len(resp.Items))
	for _, it := range resp.Items {
		vd := VideoDetail{
			ID:          it.ID,
			Title:       it.Snippet.Title,
			PublishedAt: it.Snippet.PublishedAt,
		}
		if it.LiveStreamingDetails != nil {
			vd.ScheduledStartTime = it.LiveStreamingDetails.ScheduledStartTime
			vd.ActualStartTime = it.LiveStreamingDetails.ActualStartTime
			vd.ActualEndTime = it.LiveStreamingDetails.ActualEndTime
		}
		out = append(out, vd)
	}
	return out, nil
}
