package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Event struct{ID string `json:"id"`;Name string `json:"name"`;UserID string `json:"user_id,omitempty"`;SessionID string `json:"session_id,omitempty"`;Page string `json:"page,omitempty"`;Referrer string `json:"referrer,omitempty"`;UA string `json:"user_agent,omitempty"`;IP string `json:"ip,omitempty"`;Country string `json:"country,omitempty"`;Props string `json:"properties,omitempty"`;CreatedAt string `json:"created_at"`}
type TopItem struct{Name string `json:"name"`;Count int `json:"count"`}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"headcount.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS events(id TEXT PRIMARY KEY,name TEXT NOT NULL,user_id TEXT DEFAULT '',session_id TEXT DEFAULT '',page TEXT DEFAULT '',referrer TEXT DEFAULT '',user_agent TEXT DEFAULT '',ip TEXT DEFAULT '',country TEXT DEFAULT '',props TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_name ON events(name)`)
db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_page ON events(page)`)
db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_date ON events(created_at)`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Track(e *Event)error{e.ID=genID();e.CreatedAt=now();if e.Name==""{e.Name="pageview"}
_,err:=d.db.Exec(`INSERT INTO events(id,name,user_id,session_id,page,referrer,user_agent,ip,country,props,created_at)VALUES(?,?,?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.UserID,e.SessionID,e.Page,e.Referrer,e.UA,e.IP,e.Country,e.Props,e.CreatedAt);return err}
func(d *DB)TopPages(days int)[]TopItem{if days<=0{days=30};since:=time.Now().AddDate(0,0,-days).UTC().Format(time.RFC3339)
rows,_:=d.db.Query(`SELECT page,COUNT(*) c FROM events WHERE page!='' AND created_at>=? GROUP BY page ORDER BY c DESC LIMIT 20`,since);if rows==nil{return nil};defer rows.Close()
var o []TopItem;for rows.Next(){var t TopItem;rows.Scan(&t.Name,&t.Count);o=append(o,t)};return o}
func(d *DB)TopReferrers(days int)[]TopItem{if days<=0{days=30};since:=time.Now().AddDate(0,0,-days).UTC().Format(time.RFC3339)
rows,_:=d.db.Query(`SELECT referrer,COUNT(*) c FROM events WHERE referrer!='' AND created_at>=? GROUP BY referrer ORDER BY c DESC LIMIT 20`,since);if rows==nil{return nil};defer rows.Close()
var o []TopItem;for rows.Next(){var t TopItem;rows.Scan(&t.Name,&t.Count);o=append(o,t)};return o}
func(d *DB)TopEvents(days int)[]TopItem{if days<=0{days=30};since:=time.Now().AddDate(0,0,-days).UTC().Format(time.RFC3339)
rows,_:=d.db.Query(`SELECT name,COUNT(*) c FROM events WHERE created_at>=? GROUP BY name ORDER BY c DESC LIMIT 20`,since);if rows==nil{return nil};defer rows.Close()
var o []TopItem;for rows.Next(){var t TopItem;rows.Scan(&t.Name,&t.Count);o=append(o,t)};return o}
func(d *DB)UniqueUsers(days int)int{if days<=0{days=30};since:=time.Now().AddDate(0,0,-days).UTC().Format(time.RFC3339);var n int;d.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM events WHERE user_id!='' AND created_at>=?`,since).Scan(&n);return n}
func(d *DB)UniqueSessions(days int)int{if days<=0{days=30};since:=time.Now().AddDate(0,0,-days).UTC().Format(time.RFC3339);var n int;d.db.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM events WHERE session_id!='' AND created_at>=?`,since).Scan(&n);return n}
func(d *DB)RecentEvents(limit int)[]Event{if limit<=0{limit=50};rows,_:=d.db.Query(`SELECT id,name,user_id,session_id,page,referrer,user_agent,ip,country,props,created_at FROM events ORDER BY created_at DESC LIMIT ?`,limit);if rows==nil{return nil};defer rows.Close()
var o []Event;for rows.Next(){var e Event;rows.Scan(&e.ID,&e.Name,&e.UserID,&e.SessionID,&e.Page,&e.Referrer,&e.UA,&e.IP,&e.Country,&e.Props,&e.CreatedAt);o=append(o,e)};return o}
type Stats struct{TotalEvents int `json:"total_events"`;UniqueUsers int `json:"unique_users"`;UniqueSessions int `json:"unique_sessions"`;Today int `json:"today"`}
func(d *DB)Stats()Stats{var s Stats;d.db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&s.TotalEvents);s.UniqueUsers=d.UniqueUsers(30);s.UniqueSessions=d.UniqueSessions(30)
today:=time.Now().Format("2006-01-02");d.db.QueryRow(`SELECT COUNT(*) FROM events WHERE created_at>=?`,today+"T00:00:00Z").Scan(&s.Today);return s}
