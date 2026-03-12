package main

import (
	"TinderDosFilmes/internal/config"
	"TinderDosFilmes/internal/handlers"
	"TinderDosFilmes/internal/services"
	"log"
	"net/http"
	"os"
)

func main() {

	cfg := config.LoadConfig()

	tmdbService := &services.TMDBService{
		ApiKey: cfg.TMDBApiKey,
	}

	movieHandler := &handlers.MovieHandler{
		Service: tmdbService,
	}
	salaHandler := &handlers.SalaHandler{
		Service: tmdbService,
	}

	// lightweight CORS wrapper for development. if you need more control,
	// consider using a proper middleware library.
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			h(w, r)
		}
	}

	http.HandleFunc("/discover", withCORS(movieHandler.Discover))
	http.HandleFunc("/sala", withCORS(salaHandler.CriarSala))
	http.HandleFunc("/sala/", withCORS(salaHandler.Sala))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor rodando na porta %s 🚀\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
