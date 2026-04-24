package navidrome

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newLibraryWithHandler(t *testing.T, handler http.HandlerFunc) *Library {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(Credentials{ServerURL: srv.URL, Username: "alice", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	return NewLibrary(c)
}

func TestLibrary_GetUserPlaylists_Pagination(t *testing.T) {
	lib := newLibraryWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","playlists":{"playlist":[
			{"id":"1","name":"A"},{"id":"2","name":"B"},{"id":"3","name":"C"},
			{"id":"4","name":"D"},{"id":"5","name":"E"},{"id":"6","name":"F"},
			{"id":"7","name":"G"},{"id":"8","name":"H"},{"id":"9","name":"I"},
			{"id":"10","name":"J"},{"id":"11","name":"K"},{"id":"12","name":"L"}
		]}}}`)
	})
	entities, pag, err := lib.GetUserPlaylists(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 10 {
		t.Errorf("len = %d, want 10", len(entities))
	}
	if entities[0].ID != PlaylistURI("1") || entities[0].Name != "A" {
		t.Errorf("first entity: %+v", entities[0])
	}
	if !pag.HasNext || pag.NextCursor != "10" {
		t.Errorf("pagination: %+v", pag)
	}

	// Page 2
	entities, pag, err = lib.GetUserPlaylists(context.Background(), "10")
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 2 || pag.HasNext {
		t.Errorf("page 2: entities=%d pag=%+v", len(entities), pag)
	}
}

func TestLibrary_GetSavedAlbums_NativePagination(t *testing.T) {
	gotOffset := ""
	lib := newLibraryWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("type") != "alphabeticalByName" {
			t.Errorf("type = %q", q.Get("type"))
		}
		gotOffset = q.Get("offset")
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","albumList2":{"album":[
			{"id":"a1","name":"One","artist":"X"}
		]}}}`)
	})
	entities, pag, err := lib.GetSavedAlbums(context.Background(), "20")
	if err != nil {
		t.Fatal(err)
	}
	if gotOffset != "20" {
		t.Errorf("offset = %q", gotOffset)
	}
	if len(entities) != 1 {
		t.Errorf("entities: %+v", entities)
	}
	if pag.HasNext {
		t.Errorf("expected no more: %+v", pag)
	}
	if entities[0].ID != AlbumURI("a1") {
		t.Errorf("bad id: %q", entities[0].ID)
	}
}

func TestLibrary_GetPlaylistTracks_URIParse(t *testing.T) {
	lib := newLibraryWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "p1" {
			t.Errorf("id = %q", r.URL.Query().Get("id"))
		}
		fmt.Fprint(w, `{"subsonic-response":{"status":"ok","playlist":{"id":"p1","name":"Mix","entry":[
			{"id":"s1","title":"Song","artist":"Ar","album":"Al","duration":180}
		]}}}`)
	})
	entities, _, err := lib.GetPlaylistTracks(context.Background(), PlaylistURI("p1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 1 || entities[0].ID != TrackURI("s1") {
		t.Errorf("entities: %+v", entities)
	}
	if !strings.Contains(entities[0].Desc, "Ar") || !strings.Contains(entities[0].Desc, "Al") {
		t.Errorf("desc = %q", entities[0].Desc)
	}
}

func TestLibrary_GetPlaylistTracks_BadURI(t *testing.T) {
	lib := newLibraryWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not have been called")
	})
	_, _, err := lib.GetPlaylistTracks(context.Background(), "spotify:playlist:abc", "")
	if err == nil {
		t.Fatal("expected error for non-nd uri")
	}
}

func TestLibrary_SearchTracks_Pagination(t *testing.T) {
	gotOffset := ""
	lib := newLibraryWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		gotOffset = r.URL.Query().Get("songOffset")
		// Return exactly one full page (10 items) to signal HasNext.
		b := &strings.Builder{}
		b.WriteString(`{"subsonic-response":{"status":"ok","searchResult3":{"song":[`)
		for i := range 10 {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(b, `{"id":"s%d","title":"T%d","artist":"A"}`, i, i)
		}
		b.WriteString(`]}}}`)
		fmt.Fprint(w, b.String())
	})
	_, pag, err := lib.SearchTracks(context.Background(), "hello", "10")
	if err != nil {
		t.Fatal(err)
	}
	if gotOffset != "10" {
		t.Errorf("offset = %q", gotOffset)
	}
	if !pag.HasNext || pag.NextCursor != "20" {
		t.Errorf("pagination: %+v", pag)
	}
}

func TestParseURI(t *testing.T) {
	cases := []struct {
		uri, kind, id string
	}{
		{"nd:track:s1", "track", "s1"},
		{"nd:album:a1", "album", "a1"},
		{"nd:playlist:p1", "playlist", "p1"},
		{"nd:artist:r1", "artist", "r1"},
		{"spotify:track:abc", "", ""},
		{"nd:no-colon", "", ""},
	}
	for _, c := range cases {
		k, id := ParseURI(c.uri)
		if k != c.kind || id != c.id {
			t.Errorf("ParseURI(%q) = (%q, %q), want (%q, %q)", c.uri, k, id, c.kind, c.id)
		}
	}
}
