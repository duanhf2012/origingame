package player

import (
	"reflect"
	"testing"
	"time"
)

type recordingProxy struct {
	name  string
	calls *[]string
}

func (proxy *recordingProxy) OnInit(*Player) error {
	*proxy.calls = append(*proxy.calls, "init:"+proxy.name)
	return nil
}
func (proxy *recordingProxy) OnLoaded(PlayerLoadContext) error {
	*proxy.calls = append(*proxy.calls, "loaded:"+proxy.name)
	return nil
}
func (proxy *recordingProxy) OnAllLoaded(PlayerLoadContext) error {
	*proxy.calls = append(*proxy.calls, "all:"+proxy.name)
	return nil
}
func (proxy *recordingProxy) OnOnline(PlayerOnlineContext) {
	*proxy.calls = append(*proxy.calls, "online:"+proxy.name)
}
func (proxy *recordingProxy) OnOffline(PlayerOfflineContext) {
	*proxy.calls = append(*proxy.calls, "offline:"+proxy.name)
}
func (proxy *recordingProxy) OnRelease() {
	*proxy.calls = append(*proxy.calls, "release:"+proxy.name)
}

func TestProxyLifecycleUsesForwardEntryAndReverseExitOrder(t *testing.T) {
	var calls []string
	current := &Player{dataInfo: DataInfo{State: StateLoading}}
	current.registerProxy(&recordingProxy{name: "a", calls: &calls})
	current.registerProxy(&recordingProxy{name: "b", calls: &calls})
	if err := current.initialize(); err != nil {
		t.Fatal(err)
	}
	if err := current.FinishLoad(false); err != nil {
		t.Fatal(err)
	}
	if err := current.Online("gateway-pub-1", "connection-1", time.Now()); err != nil {
		t.Fatal(err)
	}
	current.Offline(time.Now(), 15*time.Minute)
	current.Release()
	want := []string{
		"init:a", "init:b", "loaded:a", "loaded:b", "all:a", "all:b",
		"online:a", "online:b", "offline:b", "offline:a", "release:b", "release:a",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}
