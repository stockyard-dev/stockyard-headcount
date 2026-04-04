package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stockyard-dev/stockyard-headcount/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}

	// Tracking (public — no auth, called from websites)
	s.mux.HandleFunc("POST /api/event", s.trackEvent)

	// Analytics (dashboard)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/pages", s.topPages)
	s.mux.HandleFunc("GET /api/referrers", s.topReferrers)
	s.mux.HandleFunc("GET /api/timeseries", s.timeseries)
	s.mux.HandleFunc("GET /api/devices", s.devices)
	s.mux.HandleFunc("GET /api/browsers", s.browsers)
	s.mux.HandleFunc("GET /api/countries", s.countries)
	s.mux.HandleFunc("GET /api/live", s.live)
	s.mux.HandleFunc("GET /api/events", s.recentEvents)

	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/headcount/"})
	})
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(v)
}
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { http.NotFound(w, r); return }
	http.Redirect(w, r, "/ui", 302)
}

func since(r *http.Request) string {
	p := r.URL.Query().Get("period")
	switch p {
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

func (s *Server) trackEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var e store.Event
	json.NewDecoder(r.Body).Decode(&e)
	if e.Page == "" && e.Name == "" {
		we(w, 400, "page or name required")
		return
	}
	if e.UserAgent == "" {
		e.UserAgent = r.UserAgent()
	}
	s.db.Track(&e)
	wj(w, 201, map[string]string{"status": "tracked"})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request)      { wj(w, 200, s.db.Stats(since(r))) }
func (s *Server) topPages(w http.ResponseWriter, r *http.Request)   { wj(w, 200, map[string]any{"pages": s.db.TopPages(since(r), 20)}) }
func (s *Server) topReferrers(w http.ResponseWriter, r *http.Request) { wj(w, 200, map[string]any{"referrers": s.db.TopReferrers(since(r), 20)}) }
func (s *Server) timeseries(w http.ResponseWriter, r *http.Request) { wj(w, 200, map[string]any{"timeseries": s.db.PageviewsByDay(since(r))}) }
func (s *Server) devices(w http.ResponseWriter, r *http.Request)    { wj(w, 200, s.db.DeviceBreakdown(since(r))) }
func (s *Server) browsers(w http.ResponseWriter, r *http.Request)   { wj(w, 200, s.db.BrowserBreakdown(since(r))) }
func (s *Server) countries(w http.ResponseWriter, r *http.Request)  { wj(w, 200, s.db.CountryBreakdown(since(r))) }
func (s *Server) live(w http.ResponseWriter, r *http.Request)       { wj(w, 200, map[string]int{"live": s.db.LiveVisitors(5)}) }
func (s *Server) recentEvents(w http.ResponseWriter, r *http.Request) { wj(w, 200, map[string]any{"events": s.db.RecentEvents(50)}) }

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats(time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339))
	wj(w, 200, map[string]any{"service": "headcount", "status": "ok", "pageviews": st["pageviews"], "live": st["live_visitors"]})
}
