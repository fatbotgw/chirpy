package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main () {

	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{}

	httpServerMux := http.NewServeMux()
	httpServerMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	httpServerMux.HandleFunc("/healthz", health)
	httpServerMux.HandleFunc("/metrics", apiCfg.hits)
	httpServerMux.HandleFunc("/reset", apiCfg.reset)


	httpServer := http.Server {
		Addr: ":" + port,
		Handler: httpServerMux,
	}

	log.Printf("\nServing files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(httpServer.ListenAndServe())
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)  // 200 OK status code
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) hits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)  // 200 OK status code
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
    })
}
