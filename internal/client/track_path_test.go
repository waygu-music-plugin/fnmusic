package fnmusic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The metadata request resolves a song's file path and rejects missing paths.
func TestTrackPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data any = map[string]any{}
		switch r.URL.Path {
		case "/music/api/v1/user/password-login":
			data = map[string]any{"userToken": "token"}
		case "/music/api/v1/track/metadata":
			switch r.URL.Query().Get("guid") {
			case "top":
				data = map[string]any{"audioSpec": map[string]any{"path": "/music/专辑/歌曲.flac"}}
			case "nested":
				data = map[string]any{"track": map[string]any{"audioSpec": map[string]any{"path": "/music/b.flac"}}}
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
	}))
	defer ts.Close()
	c, err := New(Config{Address: ts.URL, Username: "me", PassHash: HashPassword("secret")})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ guid, want string }{{"top", "/music/专辑/歌曲.flac"}, {"nested", "/music/b.flac"}} {
		got, err := c.TrackPath(tc.guid)
		if err != nil || got != tc.want {
			t.Fatalf("%s: got %q, %v", tc.guid, got, err)
		}
	}
	if _, err := c.TrackPath("missing"); err == nil {
		t.Fatal("missing path accepted")
	}
}
