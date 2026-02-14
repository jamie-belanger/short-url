package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

/*
	application stores things the application needs in a struct that's
	easy to inject into any dependency method
*/
type application struct {
	logger *slog.Logger
	links  map[string]string
}

func main() {
	// Command line params
	port := flag.Int("port", 4000, "HTTP port to listen on")
	flag.Parse()

	// And sanity check params
	switch {
	case *port < 0 || *port > 65535:
		fmt.Printf("ERROR: Port value out of range: %d\n", *port)
		os.Exit(1)
	case *port < 1024:
		fmt.Println("WARNING: Port in reserved range may fail if user is not root")
	}

	a := &application{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions {
			Level: slog.LevelInfo,
		})),
		links:  make(map[string]string),
	}
	a.logger.Info("Application starting", slog.Int("port", *port))

	s := http.NewServeMux()
	s.HandleFunc("POST /shorten", a.shortLinkCreate)
	s.HandleFunc("GET /{id}", a.shortLinkGet)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), s)
	a.logger.Error(err.Error())
	os.Exit(1)
}
