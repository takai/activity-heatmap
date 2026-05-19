package channelurl

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Ref
		wantErr bool
	}{
		{"www handle", "https://www.youtube.com/@hololive", Ref{Kind: KindHandle, Value: "hololive"}, false},
		{"bare handle", "https://youtube.com/@hololive", Ref{Kind: KindHandle, Value: "hololive"}, false},
		{"channel id", "https://www.youtube.com/channel/UCabc123", Ref{Kind: KindChannelID, Value: "UCabc123"}, false},
		{"handle with trailing slash", "https://www.youtube.com/@hololive/", Ref{Kind: KindHandle, Value: "hololive"}, false},
		{"handle with extra path", "https://www.youtube.com/@hololive/videos", Ref{Kind: KindHandle, Value: "hololive"}, false},
		{"channel id trailing slash", "https://www.youtube.com/channel/UCabc123/", Ref{Kind: KindChannelID, Value: "UCabc123"}, false},
		{"reject /c/", "https://www.youtube.com/c/example", Ref{}, true},
		{"reject /user/", "https://www.youtube.com/user/example", Ref{}, true},
		{"reject empty", "", Ref{}, true},
		{"reject non-youtube host", "https://example.com/@foo", Ref{}, true},
		{"reject empty handle", "https://www.youtube.com/@", Ref{}, true},
		{"reject empty channel id", "https://www.youtube.com/channel/", Ref{}, true},
		{"reject non-UC channel id", "https://www.youtube.com/channel/foo", Ref{}, true},
		{"reject not a URL", "not a url", Ref{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q): want error, got %+v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q): unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("Parse(%q): got %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
