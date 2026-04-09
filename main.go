package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"strings"
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
	httpServerMux.HandleFunc("GET /api/healthz", health)
	httpServerMux.HandleFunc("GET /admin/metrics", apiCfg.hits)
	httpServerMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	httpServerMux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)


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
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)  // 200 OK status code
	w.Write([]byte(fmt.Sprintf(`<html>
								  <body>
								    <h1>Welcome, Chirpy Admin</h1>
								    <p>Chirpy has been visited %d times!</p>
								  </body>
								</html>`, cfg.fileserverHits.Load())))
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

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Body string `json:"body"`
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Error decoding parameters")
		return
    }

	if len(params.Body) > 140 {
		// throw error
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	profaneWordCheck(params.Body)
    
    type validResponse struct {
	    Valid string `json:"cleaned_body"`
	}
    respondWithJSON(w, 200, validResponse{
	    Valid: profaneWordCheck(params.Body),
	})

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	// error
	type errorResponse struct {
	    Error string `json:"error"`
	}
	respBody := errorResponse{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}


func profaneWordCheck(body string) string {
	// look for:
	// kerfuffle
	// sharbert
	// fornax

	// replace with:
	// ****		<- specifically 4 asterisks
	words := strings.Split(body, " ")
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	tempWords := []string{}

	for _, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				word = "****"
			}			
		}
		tempWords = append(tempWords, word)
	}
	
	return strings.Join(tempWords, " ")
}
