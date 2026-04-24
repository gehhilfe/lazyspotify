package navidrome

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dubeyKartikay/lazyspotify/core/logger"
)

type mpvCommand struct {
	Command   []any  `json:"command"`
	RequestID uint64 `json:"request_id"`
}

type mpvReply struct {
	RequestID uint64          `json:"request_id"`
	Error     string          `json:"error"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type mpvEvent struct {
	Event string          `json:"event"`
	Name  string          `json:"name,omitempty"`
	ID    int             `json:"id,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
	Reason string         `json:"reason,omitempty"`
}

type ipcConn struct {
	mu      sync.Mutex
	conn    net.Conn
	nextID  uint64
	pending sync.Map // uint64 -> chan *mpvReply
	events  chan *mpvEvent
	closed  atomic.Bool
}

func dialMpvSocket(path string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		c, err := net.Dial("unix", path)
		if err == nil {
			return c, nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timeout")
	}
	return nil, fmt.Errorf("connect mpv socket %q: %w", path, lastErr)
}

func newIPCConn(conn net.Conn) *ipcConn {
	c := &ipcConn{conn: conn, events: make(chan *mpvEvent, 64)}
	go c.readLoop()
	return c
}

func (c *ipcConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	err := c.conn.Close()
	close(c.events)
	return err
}

func (c *ipcConn) Events() <-chan *mpvEvent {
	return c.events
}

func (c *ipcConn) readLoop() {
	defer func() {
		if !c.closed.Load() {
			c.closed.Store(true)
			close(c.events)
		}
	}()
	r := bufio.NewReader(c.conn)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			logger.Log.Debug().Err(err).Msg("mpv ipc read end")
			return
		}
		if len(line) == 0 {
			continue
		}

		// Try as reply first (has request_id).
		var probe struct {
			RequestID *uint64 `json:"request_id"`
			Event     string  `json:"event"`
		}
		if err := json.Unmarshal(line, &probe); err != nil {
			logger.Log.Debug().Err(err).Bytes("raw", line).Msg("mpv ipc parse error")
			continue
		}

		if probe.RequestID != nil {
			var reply mpvReply
			if err := json.Unmarshal(line, &reply); err != nil {
				logger.Log.Debug().Err(err).Msg("mpv reply decode")
				continue
			}
			if v, ok := c.pending.LoadAndDelete(reply.RequestID); ok {
				ch := v.(chan *mpvReply)
				select {
				case ch <- &reply:
				default:
				}
			}
			continue
		}

		if probe.Event == "" {
			continue
		}
		var ev mpvEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			logger.Log.Debug().Err(err).Msg("mpv event decode")
			continue
		}
		select {
		case c.events <- &ev:
		default:
			// Drop events if the consumer falls behind.
			logger.Log.Warn().Str("event", ev.Event).Msg("dropping mpv event, channel full")
		}
	}
}

func (c *ipcConn) send(args []any) (*mpvReply, error) {
	if c.closed.Load() {
		return nil, fmt.Errorf("mpv ipc closed")
	}
	id := atomic.AddUint64(&c.nextID, 1)
	cmd := mpvCommand{Command: args, RequestID: id}
	buf, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	buf = append(buf, '\n')

	ch := make(chan *mpvReply, 1)
	c.pending.Store(id, ch)
	defer c.pending.Delete(id)

	c.mu.Lock()
	_, err = c.conn.Write(buf)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case reply := <-ch:
		if reply.Error != "" && reply.Error != "success" {
			return reply, fmt.Errorf("mpv command %v failed: %s", args, reply.Error)
		}
		return reply, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("mpv command %v timed out", args)
	}
}

func (c *ipcConn) observeProperty(id int, name string) error {
	_, err := c.send([]any{"observe_property", id, name})
	return err
}
