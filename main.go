package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatbotgw/chirpy/internal/database"
	"github.com/google/uuid"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func main () {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	platform := os.Getenv("PLATFORM")

	dbQueries := database.New(db)

	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		db: dbQueries,
		platform: platform,
	}

	httpServerMux := http.NewServeMux()
	httpServerMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	httpServerMux.HandleFunc("GET /api/healthz", health)
	httpServerMux.HandleFunc("GET /admin/metrics", apiCfg.hits)
	httpServerMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	httpServerMux.HandleFunc("POST /api/users", apiCfg.newUser)
	httpServerMux.HandleFunc("POST /api/chirps", apiCfg.handlerChirp)
	httpServerMux.HandleFunc("GET /api/chirps", apiCfg.handlerChirps)


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
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden")
	}
	cfg.fileserverHits.Store(0)
	cfg.db.ResetUsers(r.Context())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	// create a chirp
	type Chirp struct {
	    ID        uuid.UUID `json:"id"`
	    CreatedAt time.Time `json:"created_at"`
	    UpdatedAt time.Time `json:"updated_at"`
	    Body      string    `json:"body"`
	    UserID    uuid.UUID `json:"user_id"`
	}

	// also, move validate code here
    type parameters struct {
        Body string `json:"body"`
        UserID uuid.UUID `json:"user_id"`
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

	newChirp := database.CreateChirpParams{
		ID: 		uuid.New(),
		CreatedAt: 	time.Now(),
		UpdatedAt: 	time.Now(),
		Body:		profaneWordCheck(params.Body),
		UserID: 	params.UserID,
	}
	chirpRow, err := cfg.db.CreateChirp(context.Background(), newChirp)
	if err != nil {
		log.Printf("Error creating db entry: %s", err)
		return
	}
	chirpResponse := Chirp {
		ID: chirpRow.ID,
		CreatedAt: chirpRow.CreatedAt,
		UpdatedAt: chirpRow.UpdatedAt,
		Body: chirpRow.Body,
		UserID: chirpRow.UserID,
	}
    
	respondWithJSON(w, 201, chirpResponse)
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

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
	    ID        uuid.UUID `json:"id"`
	    CreatedAt time.Time `json:"created_at"`
	    UpdatedAt time.Time `json:"updated_at"`
	    Body      string    `json:"body"`
	    UserID    uuid.UUID `json:"user_id"`
	}

	chirpArray, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error reading db entry: %s", err)
		return
	}

	var responseChirps []Chirp
	for _, chirp := range chirpArray {
	    responseChirps = append(responseChirps, Chirp{
	        ID: chirp.ID,
	        CreatedAt: chirp.CreatedAt,
	        UpdatedAt: chirp.UpdatedAt,
	        Body: chirp.Body,
	        UserID: chirp.UserID,
	    })
	}

	respondWithJSON(w, 200, responseChirps)
}

func (cfg *apiConfig) newUser(w http.ResponseWriter, r *http.Request) {
	// accepts an email as JSON in the request body and returns 
	// the user's ID, email, and timestamps in the response body
    type parameters struct {
        Email string `json:"email"`
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Error decoding parameters")
		return
    }

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error retrieving user from database: %s", err)
		respondWithError(w, 500, "Error retrieving user")
		return
	}

	// maps the database package user to the main package user
    respondWithJSON(w, 201, User{
	    ID:        user.ID,
	    CreatedAt: user.CreatedAt,
	    UpdatedAt: user.UpdatedAt,
	    Email: user.Email,
	})

}
