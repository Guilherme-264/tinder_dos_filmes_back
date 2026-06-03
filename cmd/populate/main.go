package main

import (
	"TinderDosFilmes/internal/config"
	"TinderDosFilmes/internal/database"
	"TinderDosFilmes/internal/services"
	"log"
	"time"

	"github.com/lib/pq"
)

var generos = []int{28, 35, 10749, 27, 878, 18, 53, 16, 99, 14}
var streamings = []int{8, 337, 119, 531, 1899, 2, 307}

func main() {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	defer db.Close()
	_, err = db.Exec(`TRUNCATE TABLE filmes`)
	if err != nil {
		log.Fatal("Erro ao limpar tabela:", err)
	}
	log.Println("Tabela filmes limpa!")

	svc := &services.TMDBService{ApiKey: cfg.TMDBApiKey}

	log.Println("Iniciando população do banco...")

	for _, genero := range generos {
		for _, streaming := range streamings {
			for pagina := 1; pagina <= 10; pagina++ {
				filmes, err := svc.BuscarPorGeneroStreaming(genero, streaming, pagina)
				if err != nil {
					log.Printf("Erro genero %d streaming %d pagina %d: %v", genero, streaming, pagina, err)
					continue
				}
				if len(filmes) == 0 {
					break
				}

				for _, f := range filmes {
					_, err := db.Exec(`
						INSERT INTO filmes (tmdb_id, titulo, visao_geral, poster_path, nota_media, data_lancamento, generos, streamings)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
						ON CONFLICT (tmdb_id) DO UPDATE SET
							generos    = (SELECT ARRAY(SELECT DISTINCT unnest(filmes.generos || EXCLUDED.generos))),
							streamings = (SELECT ARRAY(SELECT DISTINCT unnest(filmes.streamings || EXCLUDED.streamings)))`,
						f.ID, f.Title, f.Overview, f.PosterPath,
						f.VoteAverage, f.ReleaseDate,
						pq.Array([]int{genero}),
						pq.Array([]int{streaming}),
					)
					if err != nil {
						log.Printf("Erro ao salvar filme %d: %v", f.ID, err)
					}
				}

				log.Printf("✓ Gênero %d | Streaming %d | Página %d | %d filmes", genero, streaming, pagina, len(filmes))
				time.Sleep(250 * time.Millisecond)
			}
		}
	}

	log.Println("✅ População concluída!")
}
