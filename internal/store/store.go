package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type Event struct {
	ID        string `json:"id"`
	Name      string `json:"name"` // pageview, click, custom
	Page      string `json:"page"`
	Referrer  string `json:"referrer"`
	UserAgent string `json:"user_agent"`
	SessionID string `json:"session_id"`
	Country   string `json:"country"`
	Device    string `json:"device"` // desktop, mobile, tablet
	Browser   string `json:"browser"`
	Props     string `json:"props"` // JSON custom properties
	CreatedAt string `json:"created_at"`
}

type PageStats struct {
	Page  string `json:"page"`
	Views int    `json:"views"`
}

type ReferrerStats struct {
	Referrer string `json:"referrer"`
	Count    int    `json:"count"`
}

type TimeSeries struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "headcount.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS events(
		id TEXT PRIMARY KEY, name TEXT NOT NULL DEFAULT 'pageview',
		page TEXT DEFAULT '', referrer TEXT DEFAULT '',
		user_agent TEXT DEFAULT '', session_id TEXT DEFAULT '',
		country TEXT DEFAULT '', device TEXT DEFAULT '',
		browser TEXT DEFAULT '', props TEXT DEFAULT '{}',
		created_at TEXT DEFAULT(datetime('now')))`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_page ON events(page)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id)`)

	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) Track(e *Event) error {
	e.ID = genID()
	if e.CreatedAt == "" {
		e.CreatedAt = now()
	}
	if e.Name == "" {
		e.Name = "pageview"
	}
	if e.Props == "" {
		e.Props = "{}"
	}
	if e.Device == "" {
		e.Device = detectDevice(e.UserAgent)
	}
	if e.Browser == "" {
		e.Browser = detectBrowser(e.UserAgent)
	}
	_, err := d.db.Exec(`INSERT INTO events(id,name,page,referrer,user_agent,session_id,country,device,browser,props,created_at)VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.Name, e.Page, e.Referrer, e.UserAgent, e.SessionID, e.Country, e.Device, e.Browser, e.Props, e.CreatedAt)
	return err
}

func (d *DB) TopPages(since string, limit int) []PageStats {
	if limit <= 0 { limit = 20 }
	rows, _ := d.db.Query(`SELECT page, COUNT(*) as cnt FROM events WHERE name='pageview' AND created_at>=? GROUP BY page ORDER BY cnt DESC LIMIT ?`, since, limit)
	if rows == nil { return []PageStats{} }
	defer rows.Close()
	var out []PageStats
	for rows.Next() {
		var p PageStats
		rows.Scan(&p.Page, &p.Views)
		out = append(out, p)
	}
	if out == nil { return []PageStats{} }
	return out
}

func (d *DB) TopReferrers(since string, limit int) []ReferrerStats {
	if limit <= 0 { limit = 20 }
	rows, _ := d.db.Query(`SELECT referrer, COUNT(*) as cnt FROM events WHERE name='pageview' AND referrer!='' AND created_at>=? GROUP BY referrer ORDER BY cnt DESC LIMIT ?`, since, limit)
	if rows == nil { return []ReferrerStats{} }
	defer rows.Close()
	var out []ReferrerStats
	for rows.Next() {
		var r ReferrerStats
		rows.Scan(&r.Referrer, &r.Count)
		out = append(out, r)
	}
	if out == nil { return []ReferrerStats{} }
	return out
}

func (d *DB) PageviewsByDay(since string) []TimeSeries {
	rows, _ := d.db.Query(`SELECT DATE(created_at) as d, COUNT(*) FROM events WHERE name='pageview' AND created_at>=? GROUP BY d ORDER BY d ASC`, since)
	if rows == nil { return []TimeSeries{} }
	defer rows.Close()
	var out []TimeSeries
	for rows.Next() {
		var t TimeSeries
		rows.Scan(&t.Date, &t.Count)
		out = append(out, t)
	}
	if out == nil { return []TimeSeries{} }
	return out
}

func (d *DB) DeviceBreakdown(since string) map[string]int {
	rows, _ := d.db.Query(`SELECT device, COUNT(*) FROM events WHERE name='pageview' AND created_at>=? GROUP BY device`, since)
	if rows == nil { return map[string]int{} }
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var dev string; var c int
		rows.Scan(&dev, &c)
		if dev == "" { dev = "unknown" }
		out[dev] = c
	}
	return out
}

func (d *DB) BrowserBreakdown(since string) map[string]int {
	rows, _ := d.db.Query(`SELECT browser, COUNT(*) FROM events WHERE name='pageview' AND created_at>=? GROUP BY browser`, since)
	if rows == nil { return map[string]int{} }
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var b string; var c int
		rows.Scan(&b, &c)
		if b == "" { b = "unknown" }
		out[b] = c
	}
	return out
}

func (d *DB) CountryBreakdown(since string) map[string]int {
	rows, _ := d.db.Query(`SELECT country, COUNT(*) FROM events WHERE name='pageview' AND country!='' AND created_at>=? GROUP BY country ORDER BY COUNT(*) DESC LIMIT 20`, since)
	if rows == nil { return map[string]int{} }
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var c string; var n int
		rows.Scan(&c, &n)
		out[c] = n
	}
	return out
}

func (d *DB) LiveVisitors(minutes int) int {
	if minutes <= 0 { minutes = 5 }
	since := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute).Format(time.RFC3339)
	var n int
	d.db.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM events WHERE created_at>=?`, since).Scan(&n)
	return n
}

func (d *DB) Stats(since string) map[string]any {
	var pageviews, sessions, events int
	d.db.QueryRow(`SELECT COUNT(*) FROM events WHERE name='pageview' AND created_at>=?`, since).Scan(&pageviews)
	d.db.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM events WHERE created_at>=?`, since).Scan(&sessions)
	d.db.QueryRow(`SELECT COUNT(*) FROM events WHERE created_at>=?`, since).Scan(&events)
	bounceRate := 0.0
	if sessions > 0 {
		var single int
		d.db.QueryRow(`SELECT COUNT(*) FROM (SELECT session_id FROM events WHERE name='pageview' AND created_at>=? GROUP BY session_id HAVING COUNT(*)=1)`, since).Scan(&single)
		bounceRate = float64(single) / float64(sessions) * 100
	}
	return map[string]any{
		"pageviews":     pageviews,
		"sessions":      sessions,
		"events":        events,
		"bounce_rate":   fmt.Sprintf("%.1f", bounceRate),
		"live_visitors": d.LiveVisitors(5),
	}
}

func (d *DB) RecentEvents(limit int) []Event {
	if limit <= 0 { limit = 50 }
	rows, _ := d.db.Query(`SELECT id,name,page,referrer,user_agent,session_id,country,device,browser,props,created_at FROM events ORDER BY created_at DESC LIMIT ?`, limit)
	if rows == nil { return []Event{} }
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		rows.Scan(&e.ID, &e.Name, &e.Page, &e.Referrer, &e.UserAgent, &e.SessionID, &e.Country, &e.Device, &e.Browser, &e.Props, &e.CreatedAt)
		out = append(out, e)
	}
	if out == nil { return []Event{} }
	return out
}

func detectDevice(ua string) string {
	if ua == "" { return "unknown" }
	for _, k := range []string{"iPhone", "Android", "Mobile"} {
		if contains(ua, k) { return "mobile" }
	}
	for _, k := range []string{"iPad", "Tablet"} {
		if contains(ua, k) { return "tablet" }
	}
	return "desktop"
}

func detectBrowser(ua string) string {
	if ua == "" { return "unknown" }
	if contains(ua, "Firefox") { return "Firefox" }
	if contains(ua, "Edg/") { return "Edge" }
	if contains(ua, "Chrome") { return "Chrome" }
	if contains(ua, "Safari") { return "Safari" }
	if contains(ua, "curl") { return "curl" }
	return "other"
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsImpl(s, sub))
}
func containsImpl(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub { return true }
	}
	return false
}
