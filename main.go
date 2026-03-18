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

	httpServer := http.Server {
		Addr: ":" + port,
		Handler: httpServerMux,
	}
	

	log.Printf("\nServing files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(httpServer.ListenAndServe())
}
