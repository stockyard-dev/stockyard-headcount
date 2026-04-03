package server
import ("encoding/json";"log";"net/http";"github.com/stockyard-dev/stockyard-headcount/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("POST /api/track",s.track)
s.mux.HandleFunc("GET /api/events",s.events)
s.mux.HandleFunc("GET /api/top/pages",s.topPages);s.mux.HandleFunc("GET /api/top/referrers",s.topReferrers);s.mux.HandleFunc("GET /api/top/events",s.topEvents)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/headcount/"})})
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);return s}

func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)track(w http.ResponseWriter,r *http.Request){var e store.Event;json.NewDecoder(r.Body).Decode(&e);if e.IP==""{e.IP=r.RemoteAddr};if e.UA==""{e.UA=r.UserAgent()};s.db.Track(&e);w.Header().Set("Access-Control-Allow-Origin","*");wj(w,200,map[string]string{"tracked":"ok"})}
func(s *Server)events(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"events":oe(s.db.RecentEvents(100))})}
func(s *Server)topPages(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"pages":oe(s.db.TopPages(30))})}
func(s *Server)topReferrers(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"referrers":oe(s.db.TopReferrers(30))})}
func(s *Server)topEvents(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"events":oe(s.db.TopEvents(30))})}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"headcount","events":st.TotalEvents,"users":st.UniqueUsers})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
