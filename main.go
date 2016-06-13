package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/braintree/manners"
	"github.com/kelseyhightower/hashiapp/handlers"
	"github.com/kelseyhightower/hashiapp/health"
	"github.com/kelseyhightower/hashiapp/user"
)

const version = "1.0.0"

func main() {
	secret, err := getJWTSecret()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(secret)

	log.Println("Starting Hashiapp...")
	httpAddr := os.Getenv("NOMAD_ADDR_http")
	if httpAddr == "" {
		log.Fatal("NOMAD_ADDR_http must be set and non-empty")
	}
	log.Printf("HTTP service listening on %s", httpAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HelloHandler)
	mux.Handle("/login", handlers.LoginHandler(secret, user.DB))
	mux.Handle("/secure", handlers.JWTAuthHandler(handlers.HelloHandler))
	mux.Handle("/version", handlers.VersionHandler(version))
	mux.HandleFunc("/healthz", health.HealthzHandler)
	mux.HandleFunc("/readiness", health.ReadinessHandler)
	mux.HandleFunc("/healthz/status", health.HealthzStatusHandler)
	mux.HandleFunc("/readiness/status", health.ReadinessStatusHandler)

	httpServer := manners.NewServer()
	httpServer.Addr = httpAddr
	httpServer.Handler = handlers.LoggingHandler(mux)

	errChan := make(chan error, 10)
	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	// Manage Database Connection
	dbm, err := NewDBManager("mysql", os.Getenv("HASHIAPP_DB_HOST"))
	if err != nil {
		log.Fatal(err)
	}
	if err := dbm.DB.Ping(); err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(err)
			}
		case s := <-signalChan:
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			health.SetReadinessStatus(http.StatusServiceUnavailable)
			httpServer.BlockingClose()
			os.Exit(0)
		}
	}
}
