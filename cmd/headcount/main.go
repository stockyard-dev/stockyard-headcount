package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-headcount/internal/server";"github.com/stockyard-dev/stockyard-headcount/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8690"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./headcount-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("headcount: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Headcount — Self-hosted user analytics\n  ─────────────────────────────────\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Track:      POST http://localhost:%s/api/track\n  Data:       %s\n  ─────────────────────────────────\n\n",port,port,port,dataDir)
log.Printf("headcount: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
