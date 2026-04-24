package navidrome

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dubeyKartikay/lazyspotify/core/daemon"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
	"github.com/dubeyKartikay/lazyspotify/librespot/models"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

const (
	volumeMax       = 100
	propPause       = "pause"
	propVolume      = "volume"
	propDuration    = "duration"
	propPlaylistPos = "playlist-pos"
	propIdleActive  = "idle-active"
)

type Player struct {
	client     *Client
	daemon     daemon.DaemonManager
	socketPath string
	ipc        *ipcConn

	events chan models.PlayerEvent
	ready  chan error
	done   chan struct{}

	mu            sync.Mutex
	currentTracks []Song
	currentIdx    int
	lastVolume    int
	lastDuration  int
	lastPosition  int
	isPaused      bool
}

func NewPlayer(client *Client) (*Player, error) {
	cfg := utils.GetConfig()
	cmdTemplate := cfg.Player.Mpv.Cmd
	if len(cmdTemplate) == 0 {
		cmdTemplate = []string{"mpv"}
	}
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("lazyspotify-mpv-%d.sock", os.Getpid()))
	// Clean up any pre-existing socket so mpv can bind a fresh one.
	_ = os.Remove(socketPath)

	args := append([]string{}, cmdTemplate...)
	args = append(args,
		"--idle",
		"--no-video",
		"--audio-display=no",
		"--really-quiet",
		"--input-ipc-server="+socketPath,
	)

	mgr, err := daemon.NewDaemonManager(args)
	if err != nil {
		return nil, fmt.Errorf("init mpv daemon: %w", err)
	}

	return &Player{
		client:     client,
		daemon:     mgr,
		socketPath: socketPath,
		events:     make(chan models.PlayerEvent, 32),
		ready:      make(chan error, 1),
		done:       make(chan struct{}),
		lastVolume: volumeMax,
	}, nil
}

func (p *Player) Start(ctx context.Context) error {
	if err := p.daemon.StartDaemon(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}
	go p.connectAndObserve()
	return nil
}

func (p *Player) connectAndObserve() {
	conn, err := dialMpvSocket(p.socketPath, 30*time.Second)
	if err != nil {
		p.ready <- err
		return
	}
	ipc := newIPCConn(conn)
	p.ipc = ipc

	// Observe the properties we actually care about.
	if err := ipc.observeProperty(1, propPause); err != nil {
		p.ready <- err
		return
	}
	_ = ipc.observeProperty(2, propVolume)
	_ = ipc.observeProperty(3, propDuration)
	_ = ipc.observeProperty(4, propPlaylistPos)
	_ = ipc.observeProperty(5, propIdleActive)

	p.ready <- nil

	go p.consumeEvents()
}

func (p *Player) consumeEvents() {
	for ev := range p.ipc.Events() {
		p.handleMpvEvent(ev)
	}
	close(p.events)
}

func (p *Player) handleMpvEvent(ev *mpvEvent) {
	switch ev.Event {
	case "property-change":
		p.handlePropertyChange(ev)
	case "start-file":
		// fires when a new entry begins loading; wait for playlist-pos/duration updates.
	case "file-loaded":
		p.emitCurrentMetadata()
	case "end-file":
		if ev.Reason == "eof" || ev.Reason == "stop" {
			// let idle-active or playlist-pos handle stop emission
		}
	}
}

func (p *Player) handlePropertyChange(ev *mpvEvent) {
	switch ev.Name {
	case propPause:
		var paused bool
		_ = json.Unmarshal(ev.Data, &paused)
		p.mu.Lock()
		p.isPaused = paused
		p.mu.Unlock()
		if paused {
			p.emit(models.PlayerEvent{Type: models.EventTypePaused, Paused: &models.PausedEventData{}})
		} else {
			p.emit(models.PlayerEvent{Type: models.EventTypePlaying, Playing: &models.PlayingEventData{}})
		}
	case propVolume:
		var v float64
		if err := json.Unmarshal(ev.Data, &v); err != nil {
			return
		}
		iv := int(math.Round(v))
		p.mu.Lock()
		p.lastVolume = iv
		p.mu.Unlock()
		p.emit(models.PlayerEvent{Type: models.EventTypeVolume, Volume: &models.VolumeEventData{Value: iv, Max: volumeMax}})
	case propDuration:
		var d float64
		if err := json.Unmarshal(ev.Data, &d); err != nil {
			return
		}
		ms := int(math.Round(d * 1000))
		p.mu.Lock()
		p.lastDuration = ms
		p.mu.Unlock()
		p.emitCurrentMetadata()
	case propPlaylistPos:
		var pos float64
		if err := json.Unmarshal(ev.Data, &pos); err != nil {
			return
		}
		idx := int(pos)
		if idx < 0 {
			return
		}
		p.mu.Lock()
		p.currentIdx = idx
		p.mu.Unlock()
		p.emitCurrentMetadata()
	case propIdleActive:
		var idle bool
		if err := json.Unmarshal(ev.Data, &idle); err != nil {
			return
		}
		if idle {
			p.emit(models.PlayerEvent{Type: models.EventTypeStopped, Stopped: &models.StoppedEventData{}})
		}
	}
}

func (p *Player) emit(ev models.PlayerEvent) {
	select {
	case p.events <- ev:
	default:
		logger.Log.Warn().Str("type", string(ev.Type)).Msg("dropping player event, channel full")
	}
}

func (p *Player) emitCurrentMetadata() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.currentIdx < 0 || p.currentIdx >= len(p.currentTracks) {
		return
	}
	song := p.currentTracks[p.currentIdx]
	coverURL := p.client.CoverArtURL(song.CoverArt)
	var coverPtr *string
	if coverURL != "" {
		coverPtr = &coverURL
	}
	md := &models.MetadataEventData{
		URI:           TrackURI(song.ID),
		Name:          song.Title,
		ArtistNames:   []string{song.Artist},
		AlbumName:     song.Album,
		AlbumCoverURL: coverPtr,
		Position:      p.lastPosition,
		Duration:      p.lastDuration,
	}
	select {
	case p.events <- models.PlayerEvent{Type: models.EventTypeMetadata, Metadata: md}:
	default:
	}
}

func (p *Player) PlayPause(ctx context.Context) error {
	if p.ipc == nil {
		return fmt.Errorf("player not ready")
	}
	_, err := p.ipc.send([]any{"cycle", propPause})
	return err
}

func (p *Player) Seek(ctx context.Context, positionMs int, relative bool) error {
	if p.ipc == nil {
		return fmt.Errorf("player not ready")
	}
	seconds := float64(positionMs) / 1000.0
	mode := "relative"
	if !relative {
		mode = "absolute"
	}
	_, err := p.ipc.send([]any{"seek", seconds, mode})
	return err
}

func (p *Player) Next(ctx context.Context) error {
	if p.ipc == nil {
		return fmt.Errorf("player not ready")
	}
	_, err := p.ipc.send([]any{"playlist-next"})
	return err
}

func (p *Player) Previous(ctx context.Context) error {
	if p.ipc == nil {
		return fmt.Errorf("player not ready")
	}
	_, err := p.ipc.send([]any{"playlist-prev"})
	return err
}

func (p *Player) SetVolume(ctx context.Context, volume int, relative bool) error {
	if p.ipc == nil {
		return fmt.Errorf("player not ready")
	}
	p.mu.Lock()
	current := p.lastVolume
	p.mu.Unlock()
	target := volume
	if relative {
		target = current + volume
	}
	if target < 0 {
		target = 0
	}
	if target > volumeMax {
		target = volumeMax
	}
	_, err := p.ipc.send([]any{"set_property", propVolume, target})
	if err == nil {
		p.mu.Lock()
		p.lastVolume = target
		p.mu.Unlock()
	}
	return err
}

func (p *Player) GetVolume(ctx context.Context) (common.VolumeInfo, error) {
	if p.ipc == nil {
		return common.VolumeInfo{}, fmt.Errorf("player not ready")
	}
	reply, err := p.ipc.send([]any{"get_property", propVolume})
	if err != nil {
		return common.VolumeInfo{}, err
	}
	var v float64
	if len(reply.Data) > 0 {
		_ = json.Unmarshal(reply.Data, &v)
	}
	iv := int(math.Round(v))
	p.mu.Lock()
	p.lastVolume = iv
	p.mu.Unlock()
	return common.VolumeInfo{Volume: iv, Max: volumeMax}, nil
}

func (p *Player) Events() <-chan models.PlayerEvent {
	return p.events
}

func (p *Player) WaitTillReady() error {
	return <-p.ready
}

func (p *Player) WaitForDaemonFailure() error {
	select {
	case err := <-p.daemon.RestartFailErrorChannel:
		return err
	case <-p.done:
		return nil
	}
}

func (p *Player) Destroy(ctx context.Context) {
	close(p.done)
	if p.ipc != nil {
		_ = p.ipc.Close()
	}
	p.daemon.StopDaemon()
	_ = os.Remove(p.socketPath)
}

// PlayTrack plays the given track. If contextURI is an nd:playlist or nd:album
// the full tracklist is loaded so next/previous work, and playback starts at
// the requested track.
func (p *Player) PlayTrack(ctx context.Context, uri, contextURI string) error {
	if p.ipc == nil {
		return fmt.Errorf("player not ready")
	}

	kind, id := ParseURI(uri)
	if kind != "track" || id == "" {
		return fmt.Errorf("invalid track uri %q", uri)
	}

	tracks, startIdx, err := p.resolveContext(ctx, id, contextURI)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.currentTracks = tracks
	p.currentIdx = startIdx
	p.mu.Unlock()

	// Clear the current playlist before loading new tracks.
	if _, err := p.ipc.send([]any{"playlist-clear"}); err != nil {
		logger.Log.Debug().Err(err).Msg("playlist-clear failed (likely empty)")
	}

	for i, t := range tracks {
		streamURL, err := p.client.StreamURL(t.ID, nil)
		if err != nil {
			return fmt.Errorf("build stream url: %w", err)
		}
		mode := "append"
		if i == 0 {
			mode = "replace"
		}
		if _, err := p.ipc.send([]any{"loadfile", streamURL, mode}); err != nil {
			return fmt.Errorf("loadfile: %w", err)
		}
	}

	if startIdx > 0 {
		if _, err := p.ipc.send([]any{"set_property", propPlaylistPos, startIdx}); err != nil {
			return fmt.Errorf("set playlist-pos: %w", err)
		}
	}

	if _, err := p.ipc.send([]any{"set_property", propPause, false}); err != nil {
		return fmt.Errorf("unpause: %w", err)
	}
	return nil
}

func (p *Player) resolveContext(ctx context.Context, trackID, contextURI string) ([]Song, int, error) {
	kind, id := ParseURI(contextURI)
	switch kind {
	case "playlist":
		pl, err := p.client.GetPlaylist(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		idx := indexOfTrack(pl.Entry, trackID)
		if idx < 0 {
			return []Song{findSong(pl.Entry, trackID)}, 0, nil
		}
		return pl.Entry, idx, nil
	case "album":
		al, err := p.client.GetAlbum(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		idx := indexOfTrack(al.Song, trackID)
		if idx < 0 {
			return []Song{findSong(al.Song, trackID)}, 0, nil
		}
		return al.Song, idx, nil
	default:
		// No context — play the single track.
		// We can't easily pull a single-song metadata endpoint cheaply, so just
		// construct a minimal Song. Metadata will be updated once mpv reports
		// duration and playlist-pos.
		return []Song{{ID: trackID}}, 0, nil
	}
}

func indexOfTrack(songs []Song, id string) int {
	for i, s := range songs {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func findSong(songs []Song, id string) Song {
	for _, s := range songs {
		if s.ID == id {
			return s
		}
	}
	return Song{ID: id}
}

// compile-time check that Player satisfies the expected method set used by
// core/backend.Player (avoids import cycle).
var _ interface {
	PlayPause(context.Context) error
} = (*Player)(nil)
