package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

// Event is a single tracked analytics event. Append-only — no update or
// delete in this tool. Common types are 'pageview' and custom event names.
// Props is a JSON blob for arbitrary custom properties.
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
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL DEFAULT 'pageview',
		page TEXT DEFAULT '',
		referrer TEXT DEFAULT '',
		user_agent TEXT DEFAULT '',
		session_id TEXT DEFAULT '',
		country TEXT DEFAULT '',
		device TEXT DEFAULT '',
		browser TEXT DEFAULT '',
		props TEXT DEFAULT '{}',
		created_at TEXT DEFAULT(datetime('now'))
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_page ON events(page)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_name ON events(name)`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string   { return time.Now().UTC().Format(time.RFC3339) }

// Track inserts an event into the log. Auto-fills ID, CreatedAt, Name
// (default 'pageview'), Props (default '{}'), and detects Device/Browser
// from the User-Agent string if not already set.
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
	_, err := d.db.Exec(
		`INSERT INTO events(id, name, page, referrer, user_agent, session_id, country, device, browser, props, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Name, e.Page, e.Referrer, e.UserAgent, e.SessionID, e.Country, e.Device, e.Browser, e.Props, e.CreatedAt,
	)
	return err
}

func (d *DB) TopPages(since string, limit int) []PageStats {
	if limit <= 0 {
		limit = 20
	}
	rows, _ := d.db.Query(
		`SELECT page, COUNT(*) AS cnt FROM events
		 WHERE name='pageview' AND created_at>=?
		 GROUP BY page ORDER BY cnt DESC LIMIT ?`,
		since, limit,
	)
	if rows == nil {
		return []PageStats{}
	}
	defer rows.Close()
	out := []PageStats{}
	for rows.Next() {
		var p PageStats
		rows.Scan(&p.Page, &p.Views)
		out = append(out, p)
	}
	return out
}

func (d *DB) TopReferrers(since string, limit int) []ReferrerStats {
	if limit <= 0 {
		limit = 20
	}
	rows, _ := d.db.Query(
		`SELECT referrer, COUNT(*) AS cnt FROM events
		 WHERE name='pageview' AND referrer != '' AND created_at>=?
		 GROUP BY referrer ORDER BY cnt DESC LIMIT ?`,
		since, limit,
	)
	if rows == nil {
		return []ReferrerStats{}
	}
	defer rows.Close()
	out := []ReferrerStats{}
	for rows.Next() {
		var r ReferrerStats
		rows.Scan(&r.Referrer, &r.Count)
		out = append(out, r)
	}
	return out
}

func (d *DB) PageviewsByDay(since string) []TimeSeries {
	rows, _ := d.db.Query(
		`SELECT DATE(created_at) AS d, COUNT(*) FROM events
		 WHERE name='pageview' AND created_at>=?
		 GROUP BY d ORDER BY d ASC`,
		since,
	)
	if rows == nil {
		return []TimeSeries{}
	}
	defer rows.Close()
	out := []TimeSeries{}
	for rows.Next() {
		var t TimeSeries
		rows.Scan(&t.Date, &t.Count)
		out = append(out, t)
	}
	return out
}

func (d *DB) DeviceBreakdown(since string) map[string]int {
	rows, _ := d.db.Query(
		`SELECT device, COUNT(*) FROM events
		 WHERE name='pageview' AND created_at>=?
		 GROUP BY device`,
		since,
	)
	if rows == nil {
		return map[string]int{}
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var dev string
		var c int
		rows.Scan(&dev, &c)
		if dev == "" {
			dev = "unknown"
		}
		out[dev] = c
	}
	return out
}

func (d *DB) BrowserBreakdown(since string) map[string]int {
	rows, _ := d.db.Query(
		`SELECT browser, COUNT(*) FROM events
		 WHERE name='pageview' AND created_at>=?
		 GROUP BY browser`,
		since,
	)
	if rows == nil {
		return map[string]int{}
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var b string
		var c int
		rows.Scan(&b, &c)
		if b == "" {
			b = "unknown"
		}
		out[b] = c
	}
	return out
}

func (d *DB) CountryBreakdown(since string) map[string]int {
	rows, _ := d.db.Query(
		`SELECT country, COUNT(*) FROM events
		 WHERE name='pageview' AND country != '' AND created_at>=?
		 GROUP BY country ORDER BY COUNT(*) DESC LIMIT 20`,
		since,
	)
	if rows == nil {
		return map[string]int{}
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var c string
		var n int
		rows.Scan(&c, &n)
		out[c] = n
	}
	return out
}

// LiveVisitors counts distinct sessions that emitted any event within the
// last `minutes` window.
func (d *DB) LiveVisitors(minutes int) int {
	if minutes <= 0 {
		minutes = 5
	}
	since := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute).Format(time.RFC3339)
	var n int
	d.db.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM events WHERE created_at>=?`, since).Scan(&n)
	return n
}

// Stats returns top-line analytics numbers for a given since timestamp.
// bounce_rate is the % of single-pageview sessions in the window.
func (d *DB) Stats(since string) map[string]any {
	var pageviews, sessions, events int
	d.db.QueryRow(`SELECT COUNT(*) FROM events WHERE name='pageview' AND created_at>=?`, since).Scan(&pageviews)
	d.db.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM events WHERE created_at>=?`, since).Scan(&sessions)
	d.db.QueryRow(`SELECT COUNT(*) FROM events WHERE created_at>=?`, since).Scan(&events)

	bounceRate := 0.0
	if sessions > 0 {
		var single int
		d.db.QueryRow(
			`SELECT COUNT(*) FROM (
				SELECT session_id FROM events
				WHERE name='pageview' AND created_at>=?
				GROUP BY session_id HAVING COUNT(*)=1
			)`,
			since,
		).Scan(&single)
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
	if limit <= 0 {
		limit = 50
	}
	rows, _ := d.db.Query(
		`SELECT id, name, page, referrer, user_agent, session_id, country, device, browser, props, created_at
		 FROM events ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if rows == nil {
		return []Event{}
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		rows.Scan(&e.ID, &e.Name, &e.Page, &e.Referrer, &e.UserAgent, &e.SessionID, &e.Country, &e.Device, &e.Browser, &e.Props, &e.CreatedAt)
		out = append(out, e)
	}
	return out
}

// detectDevice returns 'mobile', 'tablet', or 'desktop' from a User-Agent.
func detectDevice(ua string) string {
	if ua == "" {
		return "unknown"
	}
	for _, k := range []string{"iPhone", "Android", "Mobile"} {
		if strings.Contains(ua, k) {
			return "mobile"
		}
	}
	for _, k := range []string{"iPad", "Tablet"} {
		if strings.Contains(ua, k) {
			return "tablet"
		}
	}
	return "desktop"
}

// detectBrowser returns a coarse browser name from a User-Agent. Order
// matters because Edge UAs contain 'Chrome' as a substring.
func detectBrowser(ua string) string {
	if ua == "" {
		return "unknown"
	}
	if strings.Contains(ua, "Firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "Edg/") {
		return "Edge"
	}
	if strings.Contains(ua, "Chrome") {
		return "Chrome"
	}
	if strings.Contains(ua, "Safari") {
		return "Safari"
	}
	if strings.Contains(ua, "curl") {
		return "curl"
	}
	return "other"
}
