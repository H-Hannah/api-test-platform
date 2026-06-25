package main

import (
	"log"
	"net/http"
	"path/filepath"

	"api-test-platform/internal/api"
	"api-test-platform/internal/config"
	"api-test-platform/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	if err := st.MigrateAll(filepath.Join("migrations")); err != nil {
		log.Fatal("migrate:", err)
	}

	srv := api.NewServer(cfg, st)
	log.Printf("api-test-platform listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
