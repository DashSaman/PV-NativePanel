package coverd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel>
<title>منبع رسمی</title>
<item>
  <title>نشست هم‌اندیشی مسئولان برگزار شد</title>
  <link>https://example.org/news/1</link>
  <description>خلاصهٔ خبر در چند سطر.</description>
  <pubDate>Mon, 14 Sep 2026 08:30:00 +0000</pubDate>
</item>
<item>
  <title>گزارش تصویری از مراسم</title>
  <link>https://example.org/news/2</link>
  <enclosure url="https://example.org/img/2.jpg" type="image/jpeg" length="12000"/>
  <pubDate>Mon, 14 Sep 2026 06:00:00 +0000</pubDate>
</item>
</channel></rss>`

func TestFetcherParsesRSS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	f := NewFetcher("pvnaive-coverd-test")
	res, err := f.Fetch("منبع", srv.URL, "", "", "news")
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != FetchOK || len(res.Items) != 2 {
		t.Fatalf("status=%s items=%d", res.Status, len(res.Items))
	}
	if res.Items[0].Title != "نشست هم‌اندیشی مسئولان برگزار شد" || res.Items[0].URL != "https://example.org/news/1" {
		t.Fatalf("first item wrong: %+v", res.Items[0])
	}
	if res.Items[1].ThumbURL != "https://example.org/img/2.jpg" {
		t.Fatalf("enclosure thumb missing: %+v", res.Items[1])
	}
	if res.Items[0].PublishedAt == nil || res.Items[0].PublishedAt.UTC().Hour() != 8 {
		t.Fatalf("pubDate parse wrong: %+v", res.Items[0].PublishedAt)
	}
}

func TestFetcherConditionalGet(t *testing.T) {
	var seenETag string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"v1"` {
			seenETag = r.Header.Get("If-None-Match")
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	f := NewFetcher("pvnaive-coverd-test")
	first, err := f.Fetch("منبع", srv.URL, "", "", "news")
	if err != nil || first.ETag != `"v1"` {
		t.Fatalf("first fetch = %s %q %v", first.Status, first.ETag, err)
	}
	second, err := f.Fetch("منبع", srv.URL, first.ETag, "", "news")
	if err != nil {
		t.Fatal(err)
	}
	if second.Status != FetchNotModified {
		t.Fatalf("second fetch status = %s, want not_modified", second.Status)
	}
	if seenETag != `"v1"` {
		t.Fatal("conditional header must be sent on revalidation")
	}
}

func TestFetcherFailureIsStale(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	f := NewFetcher("pvnaive-coverd-test")
	res, err := f.Fetch("منبع", srv.URL, "", "", "news")
	if res.Status != FetchStale || err == nil {
		t.Fatalf("5xx must be stale with error: %s %v", res.Status, err)
	}
	// Network-level failure too.
	res, err = f.Fetch("منبع", "http://127.0.0.1:1/nope", "", "", "news")
	if res.Status != FetchStale || err == nil {
		t.Fatalf("network failure must be stale: %s %v", res.Status, err)
	}
	// Empty URL.
	if _, err := f.Fetch("منبع", "  ", "", "", "news"); err == nil {
		t.Fatal("empty url must error")
	}
}

func TestFetcherGarbageFeedNeverCrashes(t *testing.T) {
	garbage := []string{
		"",
		"<not xml at all",
		"<rss><channel><item><title>\x01\x02broken</title>",
		strings.Repeat("<a>]</a>", 10000),
		"<rss><channel>" + strings.Repeat("<item><title>x</title></item>", 1) + "</channel></rss>",
	}
	for i, body := range garbage {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(body))
		}))
		f := NewFetcher("pvnaive-coverd-test")
		res, err := f.Fetch("منبع", srv.URL, "", "", "news")
		srv.Close()
		_ = res
		if err == nil && res.Status == FetchOK && i == 0 {
			t.Fatal("empty feed must not be ok")
		}
	}
}

func TestFetcherClampsHostileLengths(t *testing.T) {
	long := strings.Repeat("طولانی ", 2000) // ~14000 chars
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<rss><channel><item><title>" + long + "</title><link>https://e.org/x</link></item></channel></rss>"))
	}))
	defer srv.Close()
	f := NewFetcher("pvnaive-coverd-test")
	res, err := f.Fetch("منبع", srv.URL, "", "", "news")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(res.Items[0].Title); got > 600 {
		t.Fatalf("title length %d must be clamped to 600", got)
	}
	if strings.ContainsRune(res.Items[0].Title, '\x01') {
		t.Fatal("control characters must be stripped")
	}
}
