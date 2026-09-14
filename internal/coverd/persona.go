// Package coverd implements the R6 official-style cover site: a loopback-only
// Go service rendered through Caddy for every non-panel path, so each node's
// main domain serves a distinct, natural, official-style Persian site.
//
// Red lines (docs/CAMO_ACCESS_UI_SPEC_FA.md):
//   - personas are ORIGINAL "official-style" designs; they never clone a
//     specific real organization's identity;
//   - syndicated content keeps source attribution and links out — videos are
//     embedded from official players, never re-hosted;
//   - zero panel artifacts in markup, headers, or copy.
package coverd

import (
	"crypto/sha256"
	"encoding/binary"
)

// PersonaID identifies one persona pack.
type PersonaID string

const (
	PersonaNewsPortal   PersonaID = "news-portal"
	PersonaCultural     PersonaID = "cultural"
	PersonaCityServices PersonaID = "city-services"
	PersonaNewsAgency   PersonaID = "news-agency"
	PersonaResearch     PersonaID = "research"
	PersonaCharity      PersonaID = "charity"
)

// Persona describes one original persona pack. Templates and class
// vocabularies differ structurally per persona (two nodes must never render
// structurally identical pages — CAMO-003).
type Persona struct {
	ID          PersonaID
	Title       string
	Tagline     string
	Accent      string // CSS accent color
	Background  string // page background
	Ink         string // body text color
	LayoutClass string // top-level layout class (structural marker)
}

// Personas is the shipped pack set (>= 6, per spec).
var Personas = []Persona{
	{PersonaNewsPortal, "پایگاه خبری شبکه آگاهی", "روایت دقیق رویدادها", "#1d3a6e", "#f5f7fb", "#1c2430", "layout-portal"},
	{PersonaCultural, "بنیاد فرهنگی میراث فردا", "پژوهش و نشر فرهنگ", "#7a5c2e", "#faf6ee", "#332a1c", "layout-cultural"},
	{PersonaCityServices, "سامانه خدمات شهری نیک‌بوم", "خدمات روزمره شهروندی", "#14665a", "#f2faf8", "#16302b", "layout-services"},
	{PersonaNewsAgency, "خبرگزاری منطقه‌ای پارس‌ران", "خبر فوری، مستند", "#5e1f24", "#fbf5f5", "#2c1a1b", "layout-agency"},
	{PersonaResearch, "مؤسسه مطالعات راهبردی کیان", "تحلیل و سنجش", "#274046", "#f4f8f9", "#1b2b30", "layout-research"},
	{PersonaCharity, "بنیاد خیریه دست‌های مهربان", "همت برای همدلی", "#2f5d34", "#f6faf5", "#1d2e1f", "layout-charity"},
}

// PersonaByID returns the persona with the given id (ok=false when absent).
func PersonaByID(id PersonaID) (Persona, bool) {
	for _, p := range Personas {
		if p.ID == id {
			return p, true
		}
	}
	return Persona{}, false
}

// PersonaForNode is the default assignment: stable_hash(node_id) mod N,
// stable across restarts and releases (never re-bucketed by pack count
// changes beyond an operator override).
func PersonaForNode(nodeID string) Persona {
	if len(Personas) == 0 {
		panic("coverd: no personas shipped")
	}
	h := sha256.Sum256([]byte("pvnaive/coverd/persona/v1:" + nodeID))
	idx := binary.BigEndian.Uint64(h[:8]) % uint64(len(Personas))
	return Personas[idx]
}
