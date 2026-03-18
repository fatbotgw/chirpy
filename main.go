package main

import (
	"log"
	"net/http"
)

func main () {

	const filepathRoot = "."
	const port = "8080"

	httpServerMux := http.NewServeMux()
	httpServerMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	httpServerMux.HandleFunc("/healthz", health)


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
