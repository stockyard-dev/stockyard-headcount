package server
import("encoding/json";"net/http";"github.com/stockyard-dev/stockyard-headcount/internal/store")
type Server struct{db *store.DB;limits Limits;mux *http.ServeMux}
func New(db *store.DB,tier string)*Server{s:=&Server{db:db,limits:LimitsFor(tier),mux:http.NewServeMux()};s.routes();return s}
func(s *Server)ListenAndServe(addr string)error{return(&http.Server{Addr:addr,Handler:s.mux}).ListenAndServe()}
func(s *Server)routes(){
    s.mux.HandleFunc("GET /health",s.handleHealth)
    s.mux.HandleFunc("GET /api/version",s.handleVersion)
    s.mux.HandleFunc("GET /api/stats",s.handleStats)
    s.mux.HandleFunc("GET /api/departments",s.handleListDepts)
    s.mux.HandleFunc("POST /api/departments",s.handleCreateDept)
    s.mux.HandleFunc("DELETE /api/departments/{id}",s.handleDeleteDept)
    s.mux.HandleFunc("GET /api/employees",s.handleListEmployees)
    s.mux.HandleFunc("POST /api/employees",s.handleCreateEmployee)
    s.mux.HandleFunc("PATCH /api/employees/{id}",s.handleUpdateEmployee)
    s.mux.HandleFunc("DELETE /api/employees/{id}",s.handleDeleteEmployee)
    s.mux.HandleFunc("GET /api/leave",s.handleListLeave)
    s.mux.HandleFunc("POST /api/leave",s.handleCreateLeave)
    s.mux.HandleFunc("PATCH /api/leave/{id}",s.handleUpdateLeave)
    s.mux.HandleFunc("GET /",s.handleUI)
}
func(s *Server)handleHealth(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]string{"status":"ok","service":"stockyard-headcount"})}
func(s *Server)handleVersion(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]string{"version":"0.1.0","service":"stockyard-headcount"})}
func writeJSON(w http.ResponseWriter,status int,v interface{}){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);json.NewEncoder(w).Encode(v)}
func writeError(w http.ResponseWriter,status int,msg string){writeJSON(w,status,map[string]string{"error":msg})}
func(s *Server)handleUI(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html");w.Write(dashboardHTML)}
