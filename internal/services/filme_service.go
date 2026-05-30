package services

import (
	"TinderDosFilmes/internal/models"
	"database/sql"
	"log"

	"github.com/lib/pq"
)

type FilmeService struct {
	DB *sql.DB
}

func (s *FilmeService) BuscarFilmes(generos, streamings []int) ([]models.Filme, error) {
	rows, err := s.DB.Query(`
		SELECT tmdb_id, titulo, visao_geral, poster_path, nota_media, data_lancamento, generos, streamings
		FROM filmes
		WHERE generos && $1
		  AND streamings && $2
		ORDER BY RANDOM()
		LIMIT 100`,
		pq.Array(generos),
		pq.Array(streamings),
	)
	if err != nil {
		log.Printf("Erro na query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var filmes []models.Filme
	for rows.Next() {
		var f models.Filme
		var generos pq.Int64Array
		var streamings pq.Int64Array

		err := rows.Scan(
			&f.ID, &f.Title, &f.Overview, &f.PosterPath,
			&f.VoteAverage, &f.ReleaseDate,
			&generos,
			&streamings,
		)
		if err != nil {
			log.Printf("Erro no scan: %v", err)
			continue
		}

		// Converte Int64Array para []int
		for _, g := range generos {
			f.GenreIDs = append(f.GenreIDs, int(g))
		}
		for _, s := range streamings {
			f.Streaming = append(f.Streaming, int(s))
		}

		filmes = append(filmes, f)
	}

	log.Printf("Filmes encontrados: %d", len(filmes))
	return filmes, nil
}
