package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/stockyard-dev/stockyard-headcount/internal/store"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{
		db:      db,
		mux:     http.NewServeMux(),
		limits:  limits,
		dataDir: dataDir,
	}
	s.loadPersonalConfig()

	// Tracking endpoint (public — no auth, called from instrumented websites)
	s.mux.HandleFunc("POST /api/event", s.trackEvent)
	s.mux.HandleFunc("OPTIONS /api/event", s.trackOptions)

	// Analytics endpoints (dashboard)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/pages", s.topPages)
	s.mux.HandleFunc("GET /api/referrers", s.topReferrers)
	s.mux.HandleFunc("GET /api/timeseries", s.timeseries)
	s.mux.HandleFunc("GET /api/devices", s.devices)
	s.mux.HandleFunc("GET /api/browsers", s.browsers)
	s.mux.HandleFunc("GET /api/countries", s.countries)
	s.mux.HandleFunc("GET /api/live", s.live)
	s.mux.HandleFunc("GET /api/events", s.recentEvents)

	// Personalization
	s.mux.HandleFunc("GET /api/config", s.configHandler)

	// Tier / health
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{
			"tier":        s.limits.Tier,
			"upgrade_url": "https://stockyard.dev/headcount/",
		})
	})

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ─── helpers ──────────────────────────────────────────────────────

func wj(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func we(w http.ResponseWriter, code int, msg string) {
	wj(w, code, map[string]string{"error": msg})
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}

// since parses the period query parameter into a since timestamp.
// Defaults to 30 days. Supported: today, 7d, 30d, 90d.
func since(r *http.Request) string {
	switch r.URL.Query().Get("period") {
	case "today":
		return time.Now().UTC().Truncate(24 * time.Hour).Format(time.RFC3339)
	case "7d":
		return time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339)
	case "30d":
		return time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
	case "90d":
		return time.Now().UTC().AddDate(0, 0, -90).Format(time.RFC3339)
	default:
		return time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
	}
}

// ─── personalization ──────────────────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("headcount: warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("headcount: loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

// ─── tracking ─────────────────────────────────────────────────────

// trackEvent is the public endpoint instrumented websites POST to. CORS
// headers allow it to be called from any origin (which is the whole point).
func (s *Server) trackEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	var e store.Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if e.Page == "" && e.Name == "" {
		we(w, 400, "page or name required")
		return
	}
	if e.UserAgent == "" {
		e.UserAgent = r.UserAgent()
	}
	if err := s.db.Track(&e); err != nil {
		we(w, 500, "track failed")
		return
	}
	wj(w, 201, map[string]string{"status": "tracked"})
}

func (s *Server) trackOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(204)
}

// ─── analytics ────────────────────────────────────────────────────

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.Stats(since(r)))
}

func (s *Server) topPages(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"pages": s.db.TopPages(since(r), 20)})
}

func (s *Server) topReferrers(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"referrers": s.db.TopReferrers(since(r), 20)})
}

func (s *Server) timeseries(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"timeseries": s.db.PageviewsByDay(since(r))})
}

func (s *Server) devices(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.DeviceBreakdown(since(r)))
}

func (s *Server) browsers(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.BrowserBreakdown(since(r)))
}

func (s *Server) countries(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.CountryBreakdown(since(r)))
}

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]int{"live": s.db.LiveVisitors(5)})
}

func (s *Server) recentEvents(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"events": s.db.RecentEvents(50)})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats(time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339))
	wj(w, 200, map[string]any{
		"service":   "headcount",
		"status":    "ok",
		"pageviews": st["pageviews"],
		"live":      st["live_visitors"],
	})
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
