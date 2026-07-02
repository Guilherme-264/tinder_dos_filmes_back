package main

import (
	"TinderDosFilmes/internal/config"
	"TinderDosFilmes/internal/database"
	"TinderDosFilmes/internal/services"
	"log"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT tmdb_id
		FROM filmes
		WHERE diretor IS NULL OR TRIM(diretor) = ''
		ORDER BY tmdb_id`)
	if err != nil {
		log.Fatal("Erro ao buscar filmes sem diretor: ", err)
	}

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			log.Fatal("Erro ao ler filme: ", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		log.Fatal("Erro ao listar filmes: ", err)
	}

	tmdb := &services.TMDBService{ApiKey: cfg.TMDBApiKey}
	atualizados := 0
	for _, id := range ids {
		diretor, err := tmdb.BuscarDiretores(id)
		if err != nil {
			log.Printf("Filme %d: %v", id, err)
			continue
		}
		if diretor == "" {
			continue
		}

		if _, err := db.Exec(`UPDATE filmes SET diretor = $1 WHERE tmdb_id = $2`, diretor, id); err != nil {
			log.Printf("Erro ao atualizar filme %d: %v", id, err)
			continue
		}
		atualizados++
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Backfill concluído: %d de %d filmes atualizados", atualizados, len(ids))
}
