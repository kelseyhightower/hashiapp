package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"github.com/braintree/manners"
	"github.com/kelseyhightower/hashiapp/handlers"
	"github.com/kelseyhightower/hashiapp/health"
	"github.com/kelseyhightower/hashiapp/user"
)

const version = "1.0.0"

func main() {
	log.Println("Starting Hashiapp...")

	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		log.Fatal("VAULT_TOKEN must be set and non-empty")
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		log.Fatal("VAULT_ADDR must be set and non-empty")
	}

	vc, err := newVaultClient(vaultAddr, vaultToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Getting JWT shared secret...")
	secret, err := vc.getJWTSecret("secret/hashiapp")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Getting database credentials...")
	username, password, err := vc.getDatabaseCredentials("mysql/creds/hashiapp")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing database connection pool...")
	dbAddr := os.Getenv("HASHIAPP_DB_HOST")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/hashiapp", username, password, dbAddr)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

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
	mux.HandleFunc("/healthz/status", health.HealthzStatusHandler)

	httpServer := manners.NewServer()
	httpServer.Addr = httpAddr
	httpServer.Handler = handlers.LoggingHandler(mux)

	errChan := make(chan error, 10)

	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	go func() {
		errChan <- vc.renewDatabaseCredentials()
	}()

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
			httpServer.BlockingClose()
			os.Exit(0)
		}
	}
}
