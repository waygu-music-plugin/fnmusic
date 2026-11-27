package fnmusic

import (
	"github.com/egoist/mygo/ui"
	"github.com/waygu-music-plugin/sdk/pluginapi"
	"github.com/waygu-music-plugin/sdk/plugintest"
	"testing"
)

type gainTestHost struct {
	*plugintest.Host
	updated []pluginapi.AudioPreferences
}

func (h *gainTestHost) AudioPreferencesChanged(p pluginapi.AudioPreferences) {
	h.updated = append(h.updated, p)
}

// Changing the switch persists ReplayGain and sends the active mode to the host.
func TestReplayGainSwitch(t *testing.T) {
	host := &gainTestHost{Host: plugintest.NewHost(ID)}
	prefs := &platformPreferences{}
	saves := 0
	h := NewHome(NewState(), host, NewRuntime(NewState(), prefs, func() { saves++ }))
	tt := ui.NewTester(func(c *ui.Context) { h.replayGainSettings(c, *c.Theme()) }, 800, 100)
	tt.Frame()
	if err := tt.Click("启用音量均衡"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !prefs.ReplayGain || saves != 1 || len(host.updated) != 1 || host.updated[0].ReplayGainMode != "album" {
		t.Fatalf("enabled: %+v saves=%d updates=%+v", prefs, saves, host.updated)
	}
	if err := tt.Click("启用音量均衡"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if prefs.ReplayGain || saves != 2 || len(host.updated) != 2 || host.updated[1].ReplayGainMode != "no" {
		t.Fatalf("disabled: %+v saves=%d updates=%+v", prefs, saves, host.updated)
	}
}
