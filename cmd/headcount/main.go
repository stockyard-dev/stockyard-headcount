package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-headcount/internal/server"
	"github.com/stockyard-dev/stockyard-headcount/internal/store"
	"github.com/stockyard-dev/stockyard/bus"
)

var version = "dev"

func main() {
	portFlag := flag.String("port", "", "HTTP port (overrides PORT env var)")
	dataFlag := flag.String("data", "", "Data directory (overrides DATA_DIR env var)")
	flag.Parse()

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "9700"
	}

	dataDir := *dataFlag
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		dataDir = "./headcount-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("headcount: %v", err)
	}
	defer db.Close()

	// Bus lives one level up from the per-tool data dir so all tools in
	// a bundle share a single _bus.db. Bus failures are non-fatal.
	if b, berr := bus.Open(filepath.Dir(dataDir), "headcount"); berr != nil {
		log.Printf("headcount: bus disabled: %v", berr)
	} else {
		defer b.Close()
		subscribeToBus(b, db)
	}

	srv := server.New(db, server.DefaultLimits(), dataDir)

	fmt.Printf("\n  Headcount v%s — Self-hosted privacy-friendly analytics\n", version)
	fmt.Printf("  Dashboard:  http://localhost:%s/ui\n", port)
	fmt.Printf("  API:        http://localhost:%s/api\n", port)
	fmt.Printf("  Data:       %s\n", dataDir)
	fmt.Printf("  Questions?  hello@stockyard.dev — I read every message\n\n")

	log.Printf("headcount: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
