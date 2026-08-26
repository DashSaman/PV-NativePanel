package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DashSaman/PV-NativePanel/internal/httpapi"
)

func main() {
	addr := os.Getenv("PVNATIVE_HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewServer(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("PVNative API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
