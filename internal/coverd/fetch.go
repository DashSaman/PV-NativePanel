package coverd

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Item is one syndicated content entry cached for a node. Only
// titles/summaries/thumbnails/links are cached — media is never re-hosted
// (spec red line: embed or link out with attribution).
type Item struct {
	Source      string     `json:"source"`
	Kind        string     `json:"kind"` // news | video | message
	Title       string     `json:"title"`
	Summary     string     `json:"summary,omitempty"`
	URL         string     `json:"url"`
	ThumbURL    string     `json:"thumb_url,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// FetchStatus describes the outcome of one polite fetch attempt.
type FetchStatus string

const (
	FetchOK          FetchStatus = "ok"
	FetchNotModified FetchStatus = "not_modified"
	FetchStale       FetchStatus = "stale" // source failed; keep last snapshot
)

// FetchResult is the outcome of a single source fetch.
type FetchResult struct {
	Status       FetchStatus
	Items        []Item
	ETag         string
	LastModified string
}

// Fetcher retrieves and parses public RSS/Atom feeds politely: conditional
// GET (ETag/If-Modified-Since), a bounded body size, and a conservative
// client timeout. Scheduling (6h ± jitter, backoff) lives in the service.
type Fetcher struct {
	Client    *http.Client
	UserAgent string
	MaxBody   int64
}

// NewFetcher builds a fetcher with conservative defaults.
func NewFetcher(userAgent string) *Fetcher {
	return &Fetcher{
		Client:    &http.Client{Timeout: 20 * time.Second},
		UserAgent: userAgent,
		MaxBody:   2 << 20, // 2 MiB cap: feeds, never media
	}
}

// rssTypes: minimal RSS 2.0 + Atom tolerant structs.
type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
	Entries []atomEntry `xml:"entry"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	MediaURL    string `xml:"media>content"`
	Enclosure   struct {
		URL string `xml:"url,attr"`
	} `xml:"enclosure"`
	PubDate string `xml:"pubDate"`
}

type atomEntry struct {
	Title string `xml:"title"`
	Links []struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Summary   string `xml:"summary"`
	Content   string `xml:"content"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
}

// Fetch performs one conditional GET against src and parses RSS/Atom.
// Any parse or network failure returns FetchStale with no items — the caller
// keeps the last snapshot and marks it stale (site never renders empty).
func (f *Fetcher) Fetch(source string, srcURL string, etag string, lastModified string, kind string) (FetchResult, error) {
	if strings.TrimSpace(srcURL) == "" {
		return FetchResult{Status: FetchStale}, errors.New("coverd: empty source url")
	}
	req, err := http.NewRequest(http.MethodGet, srcURL, nil)
	if err != nil {
		return FetchResult{Status: FetchStale}, fmt.Errorf("coverd: request: %w", err)
	}
	req.Header.Set("User-Agent", f.UserAgent)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return FetchResult{Status: FetchStale}, fmt.Errorf("coverd: fetch %s: %w", source, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return FetchResult{Status: FetchNotModified, ETag: etag, LastModified: lastModified}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return FetchResult{Status: FetchStale}, fmt.Errorf("coverd: fetch %s: status %d", source, resp.StatusCode)
	}

	body := http.MaxBytesReader(nil, resp.Body, f.MaxBody)
	var feed rssFeed
	dec := xml.NewDecoder(body)
	dec.Strict = false
	if err := dec.Decode(&feed); err != nil {
		return FetchResult{Status: FetchStale}, fmt.Errorf("coverd: parse %s: %w", source, err)
	}

	items := make([]Item, 0, len(feed.Channel.Items)+len(feed.Entries))
	for _, it := range feed.Channel.Items {
		thumb := it.MediaURL
		if thumb == "" {
			thumb = it.Enclosure.URL
		}
		items = append(items, Item{
			Source: source, Kind: kind,
			Title: sanitizeText(it.Title), Summary: sanitizeText(it.Description),
			URL:         strings.TrimSpace(it.Link),
			ThumbURL:    strings.TrimSpace(thumb),
			PublishedAt: parseTime(it.PubDate),
		})
	}
	for _, e := range feed.Entries {
		link := ""
		for _, l := range e.Links {
			if l.Href != "" {
				link = l.Href
				break
			}
		}
		published := parseTime(e.Published)
		if published == nil {
			published = parseTime(e.Updated)
		}
		items = append(items, Item{
			Source: source, Kind: kind,
			Title: sanitizeText(e.Title), Summary: sanitizeText(e.Summary),
			URL: link, PublishedAt: published,
		})
	}
	if len(items) == 0 {
		return FetchResult{Status: FetchStale}, errors.New("coverd: feed contained no items")
	}
	return FetchResult{
		Status:       FetchOK,
		Items:        items,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}, nil
}

// sanitizeText strips control characters and clamps absurd lengths so fuzzed
// or hostile feeds can never crash or overwhelm a page.
func sanitizeText(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || (r >= 32 && r != 127) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > 600 {
		out = out[:600]
	}
	return out
}

func parseTime(raw string) *time.Time {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	layouts := []string{
		time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822,
		"2006-01-02T15:04:05Z07:00", "2006-01-02T15:04:05", "2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, strings.TrimSpace(raw)); err == nil {
			t = t.UTC()
			return &t
		}
	}
	return nil
}
