package coverd

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// DBStore is the durable ContentStore backing the flip: it reads the
// node's cached snapshot through migration 0029's SECURITY DEFINER read
// projection (pvnaive.cover_latest) and persists persona choice through
// cover_set_persona / cover_persona. Direct table access is forbidden by
// RLS; these functions are the only path.
type DBStore struct {
	DB *sql.DB
}

// Latest implements ContentStore. It returns the stored snapshot rows
// newest-first; an empty node yields an empty slice (never an error for
// "no content" — the renderer falls back to evergreen pages).
func (s *DBStore) Latest(nodeID string, limit int) ([]Item, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("coverd: db store is unavailable")
	}
	if strings.TrimSpace(nodeID) == "" {
		return nil, errors.New("coverd: node id is required")
	}
	if limit < 1 {
		limit = 40
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.DB.Query(
		`SELECT source, kind, title, summary, url, thumb_url, published_at
         FROM pvnaive.cover_latest($1, $2)`,
		nodeID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("coverd: read latest snapshot: %w", err)
	}
	defer rows.Close()
	items := make([]Item, 0, limit)
	for rows.Next() {
		var item Item
		var source, kind, title, url string
		var summary, thumbURL sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&source, &kind, &title, &summary, &url, &thumbURL, &publishedAt); err != nil {
			return nil, fmt.Errorf("coverd: scan snapshot row: %w", err)
		}
		item = Item{
			Source:   source,
			Kind:     kind,
			Title:    title,
			URL:      url,
			ThumbURL: thumbURL.String,
		}
		if summary.Valid {
			item.Summary = summary.String
		}
		if publishedAt.Valid {
			published := publishedAt.Time
			item.PublishedAt = &published
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("coverd: iterate snapshot rows: %w", err)
	}
	return items, nil
}

// StoredPersona reads the operator-pinned persona for a node (empty/unknown
// -> ok=false so callers fall back to deterministic assignment).
func (s *DBStore) StoredPersona(nodeID string) (PersonaID, bool, error) {
	if s == nil || s.DB == nil {
		return "", false, errors.New("coverd: db store is unavailable")
	}
	if strings.TrimSpace(nodeID) == "" {
		return "", false, errors.New("coverd: node id is required")
	}
	var raw sql.NullString
	if err := s.DB.QueryRow(`SELECT pvnaive.cover_persona($1)`, nodeID).Scan(&raw); err != nil {
		return "", false, fmt.Errorf("coverd: read stored persona: %w", err)
	}
	if !raw.Valid {
		return "", false, nil
	}
	id := PersonaID(strings.TrimSpace(raw.String))
	if _, ok := PersonaByID(id); !ok {
		// A persona id that no longer exists in the shipped pack set must
		// never crash the cover; fall back instead.
		return "", false, nil
	}
	return id, true, nil
}

// SetPersona pins a persona for a node through the SECURITY DEFINER
// mutator. Unknown personas are refused client-side (fail-closed input
// contract; the database would reject them too).
func (s *DBStore) SetPersona(nodeID string, id PersonaID) error {
	if s == nil || s.DB == nil {
		return errors.New("coverd: db store is unavailable")
	}
	if strings.TrimSpace(nodeID) == "" {
		return errors.New("coverd: node id is required")
	}
	if _, ok := PersonaByID(id); !ok {
		return fmt.Errorf("coverd: unknown persona %q", id)
	}
	if _, err := s.DB.Exec(`SELECT pvnaive.cover_set_persona($1, $2)`, nodeID, string(id)); err != nil {
		return fmt.Errorf("coverd: persist persona: %w", err)
	}
	return nil
}
