package navidrome

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	apiVersion = "1.16.1"
	clientName = "lazyspotify"
)

type Credentials struct {
	ServerURL          string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

type Client struct {
	creds Credentials
	http  *http.Client
}

func NewClient(creds Credentials) (*Client, error) {
	if strings.TrimSpace(creds.ServerURL) == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if strings.TrimSpace(creds.Username) == "" {
		return nil, fmt.Errorf("username is required")
	}
	if strings.TrimSpace(creds.Password) == "" {
		return nil, fmt.Errorf("password is required")
	}
	transport := &http.Transport{}
	if creds.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		creds: creds,
		http:  &http.Client{Timeout: 30 * time.Second, Transport: transport},
	}, nil
}

func (c *Client) authQuery() (url.Values, error) {
	saltBytes := make([]byte, 8)
	if _, err := rand.Read(saltBytes); err != nil {
		return nil, err
	}
	salt := hex.EncodeToString(saltBytes)
	hash := md5.Sum([]byte(c.creds.Password + salt))
	token := hex.EncodeToString(hash[:])

	q := url.Values{}
	q.Set("u", c.creds.Username)
	q.Set("t", token)
	q.Set("s", salt)
	q.Set("v", apiVersion)
	q.Set("c", clientName)
	q.Set("f", "json")
	return q, nil
}

func (c *Client) buildURL(endpoint string, extra url.Values) (string, error) {
	base, err := url.Parse(strings.TrimRight(c.creds.ServerURL, "/"))
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/rest/" + endpoint
	q, err := c.authQuery()
	if err != nil {
		return "", err
	}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	base.RawQuery = q.Encode()
	return base.String(), nil
}

// StreamURL returns a ready-to-fetch URL with auth params baked in.
func (c *Client) StreamURL(id string, extra url.Values) (string, error) {
	params := url.Values{}
	params.Set("id", id)
	for k, vs := range extra {
		for _, v := range vs {
			params.Add(k, v)
		}
	}
	return c.buildURL("stream", params)
}

// CoverArtURL returns an auth-bearing URL for a cover art id.
func (c *Client) CoverArtURL(id string) string {
	if strings.TrimSpace(id) == "" {
		return ""
	}
	u, err := c.buildURL("getCoverArt", url.Values{"id": []string{id}})
	if err != nil {
		return ""
	}
	return u
}

type errorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type envelope[T any] struct {
	Response T `json:"subsonic-response"`
}

type baseResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Type    string    `json:"type"`
	Error   *errorObj `json:"error,omitempty"`
}

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("subsonic api error %d: %s", e.Code, e.Message)
}

func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// 40: wrong creds, 41: token auth not supported, 10: required param missing
		return apiErr.Code == 40 || apiErr.Code == 41 || apiErr.Code == 44 || apiErr.Code == 45 || apiErr.Code == 50
	}
	return false
}

func (c *Client) do(ctx context.Context, endpoint string, extra url.Values, out any) error {
	u, err := c.buildURL(endpoint, extra)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var base envelope[baseResponse]
	if err := json.Unmarshal(body, &base); err != nil {
		return fmt.Errorf("decode subsonic envelope: %w", err)
	}
	if base.Response.Status != "ok" {
		if base.Response.Error != nil {
			return &APIError{Code: base.Response.Error.Code, Message: base.Response.Error.Message}
		}
		return &APIError{Code: 0, Message: "subsonic returned failed status without error body"}
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}

// --- Response payloads ---

type Song struct {
	ID          string `json:"id"`
	Parent      string `json:"parent"`
	Title       string `json:"title"`
	Album       string `json:"album"`
	AlbumID     string `json:"albumId"`
	Artist      string `json:"artist"`
	ArtistID    string `json:"artistId"`
	Duration    int    `json:"duration"`
	CoverArt    string `json:"coverArt"`
	Track       int    `json:"track"`
	Year        int    `json:"year"`
	Genre       string `json:"genre"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	Suffix      string `json:"suffix"`
}

type Album struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	ArtistID  string `json:"artistId"`
	SongCount int    `json:"songCount"`
	Duration  int    `json:"duration"`
	CoverArt  string `json:"coverArt"`
	Year      int    `json:"year"`
	Genre     string `json:"genre"`
	Song      []Song `json:"song,omitempty"`
}

type Artist struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CoverArt   string  `json:"coverArt"`
	AlbumCount int     `json:"albumCount"`
	Album      []Album `json:"album,omitempty"`
}

type Playlist struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Owner     string `json:"owner"`
	SongCount int    `json:"songCount"`
	Duration  int    `json:"duration"`
	CoverArt  string `json:"coverArt"`
	Public    bool   `json:"public"`
	Entry     []Song `json:"entry,omitempty"`
}

type pingResp struct {
	baseResponse
}

type playlistsResp struct {
	baseResponse
	Playlists struct {
		Playlist []Playlist `json:"playlist"`
	} `json:"playlists"`
}

type playlistResp struct {
	baseResponse
	Playlist Playlist `json:"playlist"`
}

type albumResp struct {
	baseResponse
	Album Album `json:"album"`
}

type artistResp struct {
	baseResponse
	Artist Artist `json:"artist"`
}

type starredResp struct {
	baseResponse
	Starred2 struct {
		Song   []Song   `json:"song"`
		Album  []Album  `json:"album"`
		Artist []Artist `json:"artist"`
	} `json:"starred2"`
}

type albumListResp struct {
	baseResponse
	AlbumList2 struct {
		Album []Album `json:"album"`
	} `json:"albumList2"`
}

type artistsResp struct {
	baseResponse
	Artists struct {
		Index []struct {
			Name   string   `json:"name"`
			Artist []Artist `json:"artist"`
		} `json:"index"`
	} `json:"artists"`
}

type searchResp struct {
	baseResponse
	SearchResult3 struct {
		Song   []Song   `json:"song"`
		Album  []Album  `json:"album"`
		Artist []Artist `json:"artist"`
	} `json:"searchResult3"`
}

// --- Public methods ---

func (c *Client) Ping(ctx context.Context) error {
	var r envelope[pingResp]
	return c.do(ctx, "ping", nil, &r)
}

func (c *Client) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	var r envelope[playlistsResp]
	if err := c.do(ctx, "getPlaylists", nil, &r); err != nil {
		return nil, err
	}
	return r.Response.Playlists.Playlist, nil
}

func (c *Client) GetPlaylist(ctx context.Context, id string) (Playlist, error) {
	var r envelope[playlistResp]
	err := c.do(ctx, "getPlaylist", url.Values{"id": []string{id}}, &r)
	return r.Response.Playlist, err
}

func (c *Client) GetAlbum(ctx context.Context, id string) (Album, error) {
	var r envelope[albumResp]
	err := c.do(ctx, "getAlbum", url.Values{"id": []string{id}}, &r)
	return r.Response.Album, err
}

func (c *Client) GetArtist(ctx context.Context, id string) (Artist, error) {
	var r envelope[artistResp]
	err := c.do(ctx, "getArtist", url.Values{"id": []string{id}}, &r)
	return r.Response.Artist, err
}

func (c *Client) GetStarred2(ctx context.Context) (songs []Song, albums []Album, artists []Artist, err error) {
	var r envelope[starredResp]
	if e := c.do(ctx, "getStarred2", nil, &r); e != nil {
		return nil, nil, nil, e
	}
	return r.Response.Starred2.Song, r.Response.Starred2.Album, r.Response.Starred2.Artist, nil
}

func (c *Client) GetAlbumList2(ctx context.Context, listType string, offset, size int) ([]Album, error) {
	if listType == "" {
		listType = "alphabeticalByName"
	}
	if size <= 0 {
		size = 10
	}
	if offset < 0 {
		offset = 0
	}
	q := url.Values{}
	q.Set("type", listType)
	q.Set("size", strconv.Itoa(size))
	q.Set("offset", strconv.Itoa(offset))
	var r envelope[albumListResp]
	err := c.do(ctx, "getAlbumList2", q, &r)
	return r.Response.AlbumList2.Album, err
}

// GetArtists returns every indexed artist in the library, flattened.
func (c *Client) GetArtists(ctx context.Context) ([]Artist, error) {
	var r envelope[artistsResp]
	if err := c.do(ctx, "getArtists", nil, &r); err != nil {
		return nil, err
	}
	var out []Artist
	for _, idx := range r.Response.Artists.Index {
		out = append(out, idx.Artist...)
	}
	return out, nil
}

type SearchResult struct {
	Songs   []Song
	Albums  []Album
	Artists []Artist
}

func (c *Client) Search3(ctx context.Context, query string, songCount, songOffset, albumCount, albumOffset, artistCount, artistOffset int) (SearchResult, error) {
	q := url.Values{}
	q.Set("query", query)
	if songCount > 0 {
		q.Set("songCount", strconv.Itoa(songCount))
	}
	if songOffset > 0 {
		q.Set("songOffset", strconv.Itoa(songOffset))
	}
	if albumCount > 0 {
		q.Set("albumCount", strconv.Itoa(albumCount))
	}
	if albumOffset > 0 {
		q.Set("albumOffset", strconv.Itoa(albumOffset))
	}
	if artistCount > 0 {
		q.Set("artistCount", strconv.Itoa(artistCount))
	}
	if artistOffset > 0 {
		q.Set("artistOffset", strconv.Itoa(artistOffset))
	}
	var r envelope[searchResp]
	if err := c.do(ctx, "search3", q, &r); err != nil {
		return SearchResult{}, err
	}
	return SearchResult{
		Songs:   r.Response.SearchResult3.Song,
		Albums:  r.Response.SearchResult3.Album,
		Artists: r.Response.SearchResult3.Artist,
	}, nil
}
