package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

/*
application stores things the application needs in a struct that's
easy to inject into any dependency method
*/
type application struct {
	logger   *slog.Logger
	mu	     sync.RWMutex
	dbtype   DbType
	dbconn   *sql.DB
	links    map[string]string
}

func main() {
	// Command line params
	port := flag.Int("port", 4000, "HTTP port to listen on")
	database := flag.String("database", "memory", "Database type (memory or sqlite)")
	flag.Parse()

	// And sanity check params
	switch {
	case *port < 0 || *port > 65535:
		fmt.Printf("ERROR: Port value out of range: %d\n", *port)
		os.Exit(1)
	case *port < 1024:
		fmt.Println("WARNING: Port in reserved range may fail if user is not root")
	}

	var databaseType DbType
	switch *database {
	case "memory":
		databaseType = DatabaseMemory
	case "sqlite":
		databaseType = DatabaseSqlite
	default:
		fmt.Printf("ERROR: Database value not recognized: %v\n", *database)
		fmt.Println("--> supported drivers: 'memory' (default) and 'sqlite'")
		os.Exit(1)
	}

	a := &application{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		dbtype: databaseType,
	}
	a.logger.Info("Application starting", slog.Int("port", *port))
	if !a.ConnectDatabase() {
		os.Exit(1)
	}
	defer a.DisconnectDatabase()

	s := http.NewServeMux()
	s.HandleFunc("POST    /shorten", a.shortLinkCreate)
	s.HandleFunc("GET     /{id}", a.shortLinkGet)
	s.HandleFunc("DELETE  /{id}", a.shortLinkDelete)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), s)
	a.logger.Error(err.Error())
	os.Exit(1)
}
