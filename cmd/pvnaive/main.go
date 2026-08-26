package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/httpapi"
)

func main() {
	addr := os.Getenv("PVNAIVE_LISTEN")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewServer(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("PVNaive API listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
