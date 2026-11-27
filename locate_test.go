package fnmusic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/egoist/mygo/ui"
	backend "github.com/waygu-music-plugin/fnmusic/internal/client"
	"github.com/waygu-music-plugin/sdk/pluginapi"
)

type locateTestHost struct {
	nopHost
	posted chan func()
	toast  string
}

func (h *locateTestHost) Post(fn func()) { h.posted <- fn }
func (h *locateTestHost) Toast(s string) { h.toast = s }

// Locating a song opens its parent folder and scrolls to the matching track.
func TestLocateTrackOpensAndScrolls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data any = map[string]any{"userToken": "token"}
		if r.URL.Path == "/music/api/v1/track/metadata" {
			if r.URL.Query().Get("guid") != "target" {
				t.Error("incorrect track GUID")
			}
			data = map[string]any{"audioSpec": map[string]any{"path": "/music/album/target.flac"}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
	}))
	defer srv.Close()
	client, err := backend.New(backend.Config{Address: srv.URL, Username: "u", PassHash: "p"})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	if err := state.Attach(client); err != nil {
		t.Fatal(err)
	}
	state.capabilities, state.capabilitiesKnown = backend.FolderCapabilities{FolderView: true}, true
	host := &locateTestHost{posted: make(chan func(), 1)}
	h := NewHome(state, host)
	f := h.syncFolders()
	f.roots = []backend.SharedLibrary{{Name: "music", Path: "/music"}}
	page := backend.FolderPage{Name: "album", Path: "/music/album", Folders: []backend.FolderNode{{Name: "sub", Path: "/music/album/sub"}}}
	for i := 0; i < 80; i++ {
		tr := pluginapi.Track{Ref: fmt.Sprintf("id-%d", i), Name: fmt.Sprintf("song-%d", i)}
		if i == 65 {
			tr.Ref, tr.Name = "target", "target-song"
		}
		page.Tracks = append(page.Tracks, backend.FolderTrack{Track: tr})
	}
	f.pages[page.Path] = page
	h.locateTrack(&pluginapi.Track{Ref: "target"})
	(<-host.posted)()
	if got := h.Route(); got.Page != PageFolders || got.Query != page.Path {
		t.Fatalf("route = %+v", got)
	}
	tt := ui.NewTester(h.Content, 1000, 500)
	tt.Frame()
	tt.Frame()
	if err := tt.Click("target-song"); err != nil {
		t.Fatalf("target track was not scrolled into view: %v", err)
	}
	if host.toast != "" {
		t.Fatalf("unexpected error: %s", host.toast)
	}
}
