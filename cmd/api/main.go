package main

import (
	"TinderDosFilmes/internal/config"
	"TinderDosFilmes/internal/database"
	"TinderDosFilmes/internal/handlers"
	"TinderDosFilmes/internal/services"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Banco inacessível:", err)
	} else {
		log.Println("Banco conectado com sucesso!")
	}
	defer db.Close()

	tmdbService := &services.TMDBService{
		ApiKey: cfg.TMDBApiKey,
	}

	movieHandler := &handlers.MovieHandler{
		Service: tmdbService,
	}
	salaHandler := &handlers.SalaHandler{
		Service: tmdbService,
		DB:      db,
	}

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
	go func() {
		for {
			time.Sleep(10 * time.Second)
			salaHandler.ApagarSalas()
		}
	}()
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
