package navidrome

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(Credentials{ServerURL: srv.URL, Username: "alice", Password: "secret"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func assertAuthParams(t *testing.T, r *http.Request) {
	t.Helper()
	q := r.URL.Query()
	if q.Get("u") != "alice" {
		t.Errorf("u = %q, want alice", q.Get("u"))
	}
	if q.Get("v") != apiVersion {
		t.Errorf("v = %q, want %q", q.Get("v"), apiVersion)
	}
	if q.Get("c") != clientName {
		t.Errorf("c = %q, want %q", q.Get("c"), clientName)
	}
	if q.Get("f") != "json" {
		t.Errorf("f = %q, want json", q.Get("f"))
	}
	salt := q.Get("s")
	token := q.Get("t")
	if salt == "" || token == "" {
		t.Fatalf("missing salt/token: s=%q t=%q", salt, token)
	}
	want := md5.Sum([]byte("secret" + salt))
	if hex.EncodeToString(want[:]) != token {
		t.Errorf("token mismatch: got %q", token)
	}
}

func TestPingOK(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/rest/ping") {
			t.Errorf("path = %q", r.URL.Path)
		}
		assertAuthParams(t, r)
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","version":"1.16.1","type":"navidrome"}}`)
	})
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestPingWrongPassword(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"failed","error":{"code":40,"message":"Wrong username or password"}}}`)
	})
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthError(err) {
		t.Errorf("IsAuthError(%v) = false", err)
	}
}

func TestGetPlaylists(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","playlists":{"playlist":[{"id":"p1","name":"Mix","songCount":12}]}}}`)
	})
	pls, err := c.GetPlaylists(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pls) != 1 || pls[0].ID != "p1" || pls[0].Name != "Mix" || pls[0].SongCount != 12 {
		t.Errorf("unexpected: %+v", pls)
	}
}

func TestGetPlaylist(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "p1" {
			t.Errorf("id = %q", r.URL.Query().Get("id"))
		}
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","playlist":{"id":"p1","name":"Mix","entry":[{"id":"s1","title":"Song","artist":"A","album":"B","duration":180}]}}}`)
	})
	pl, err := c.GetPlaylist(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if pl.ID != "p1" || len(pl.Entry) != 1 || pl.Entry[0].Title != "Song" {
		t.Errorf("unexpected: %+v", pl)
	}
}

func TestGetAlbum(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","album":{"id":"a1","name":"Al","artist":"Ar","song":[{"id":"s1","title":"T","duration":100}]}}}`)
	})
	al, err := c.GetAlbum(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}
	if al.ID != "a1" || len(al.Song) != 1 {
		t.Errorf("unexpected: %+v", al)
	}
}

func TestGetArtist(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","artist":{"id":"r1","name":"Band","albumCount":2,"album":[{"id":"a1","name":"Al"}]}}}`)
	})
	ar, err := c.GetArtist(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	if ar.ID != "r1" || len(ar.Album) != 1 {
		t.Errorf("unexpected: %+v", ar)
	}
}

func TestGetStarred2(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","starred2":{"song":[{"id":"s1","title":"T"}],"album":[{"id":"a1","name":"Al"}],"artist":[{"id":"r1","name":"Band"}]}}}`)
	})
	songs, albums, artists, err := c.GetStarred2(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 1 || len(albums) != 1 || len(artists) != 1 {
		t.Errorf("unexpected sizes: %d %d %d", len(songs), len(albums), len(artists))
	}
}

func TestGetAlbumList2(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("type") != "alphabeticalByName" || q.Get("size") != "25" || q.Get("offset") != "50" {
			t.Errorf("params: %v", q)
		}
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","albumList2":{"album":[{"id":"a1","name":"Al"}]}}}`)
	})
	albums, err := c.GetAlbumList2(context.Background(), "alphabeticalByName", 50, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(albums) != 1 {
		t.Errorf("unexpected: %+v", albums)
	}
}

func TestGetArtists(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","artists":{"index":[
			{"name":"A","artist":[{"id":"r1","name":"Alpha"},{"id":"r2","name":"Aviator"}]},
			{"name":"B","artist":[{"id":"r3","name":"Beta"}]}
		]}}}`)
	})
	artists, err := c.GetArtists(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(artists) != 3 {
		t.Errorf("expected 3 artists flattened, got %d: %+v", len(artists), artists)
	}
	if artists[0].ID != "r1" || artists[2].ID != "r3" {
		t.Errorf("unexpected order: %+v", artists)
	}
}

func TestSearch3(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("query") != "abc" || q.Get("songCount") != "5" || q.Get("albumCount") != "6" || q.Get("artistCount") != "7" {
			t.Errorf("params: %v", q)
		}
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","searchResult3":{"song":[{"id":"s"}],"album":[{"id":"a"}],"artist":[{"id":"r"}]}}}`)
	})
	res, err := c.Search3(context.Background(), "abc", 5, 0, 6, 0, 7, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Songs) != 1 || len(res.Albums) != 1 || len(res.Artists) != 1 {
		t.Errorf("unexpected: %+v", res)
	}
}

func TestStreamURL(t *testing.T) {
	c, err := NewClient(Credentials{ServerURL: "https://music.example.com", Username: "alice", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	u, err := c.StreamURL("s1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "/rest/stream") || !strings.Contains(u, "id=s1") || !strings.Contains(u, "u=alice") {
		t.Errorf("url = %q", u)
	}
}

func TestCoverArtURL(t *testing.T) {
	c, err := NewClient(Credentials{ServerURL: "https://music.example.com", Username: "alice", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	u := c.CoverArtURL("c1")
	if !strings.Contains(u, "/rest/getCoverArt") || !strings.Contains(u, "id=c1") {
		t.Errorf("url = %q", u)
	}
	if c.CoverArtURL("") != "" {
		t.Error("empty id should yield empty url")
	}
}

func TestHTTPErrorSurfaced(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	err := c.Ping(context.Background())
	if err == nil || !strings.Contains(err.Error(), "http 500") {
		t.Errorf("expected http 500, got %v", err)
	}
}
