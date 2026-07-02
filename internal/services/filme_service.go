package services

import (
	"TinderDosFilmes/internal/models"
	"database/sql"
	"log"
	"strings"

	"github.com/lib/pq"
)

type FilmeService struct {
	DB *sql.DB
}

type FiltrosFilme struct {
	Generos    []int
	Streamings []int
	AnoInicio  *int
	AnoFim     *int
	NotaMinima float64
	Diretor    string
}

func (s *FilmeService) BuscarFilmes(filtros FiltrosFilme) ([]models.Filme, error) {
	var anoInicio interface{}
	if filtros.AnoInicio != nil {
		anoInicio = *filtros.AnoInicio
	}

	var anoFim interface{}
	if filtros.AnoFim != nil {
		anoFim = *filtros.AnoFim
	}

	var notaMinima interface{}
	if filtros.NotaMinima > 0 {
		notaMinima = filtros.NotaMinima
	}

	var diretor interface{}
	if strings.TrimSpace(filtros.Diretor) != "" {
		diretor = strings.TrimSpace(filtros.Diretor)
	}

	rows, err := s.DB.Query(`
		SELECT tmdb_id, titulo, visao_geral, poster_path, nota_media, data_lancamento, generos, streamings, COALESCE(diretor, '')
		FROM filmes
		WHERE generos && $1
		  AND streamings && $2
		  AND ($3::INTEGER IS NULL OR (data_lancamento ~ '^[0-9]{4}' AND SUBSTRING(data_lancamento FROM 1 FOR 4)::INTEGER >= $3))
		  AND ($4::INTEGER IS NULL OR (data_lancamento ~ '^[0-9]{4}' AND SUBSTRING(data_lancamento FROM 1 FOR 4)::INTEGER <= $4))
		  AND ($5::NUMERIC IS NULL OR nota_media >= $5)
		  AND ($6::TEXT IS NULL OR diretor ILIKE '%' || $6 || '%')
		ORDER BY RANDOM()
		LIMIT 100`,
		pq.Array(filtros.Generos),
		pq.Array(filtros.Streamings),
		anoInicio,
		anoFim,
		notaMinima,
		diretor,
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
			&f.Diretor,
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

// BuscarDiretores retorna nomes únicos cadastrados nos filmes.
// Um filme pode ter mais de um diretor separado por vírgula.
func (s *FilmeService) BuscarDiretores(termo string) ([]string, error) {
	termo = strings.TrimSpace(termo)
	if len([]rune(termo)) < 2 {
		return []string{}, nil
	}

	rows, err := s.DB.Query(`
		SELECT DISTINCT TRIM(nome) AS diretor
		FROM filmes
		CROSS JOIN LATERAL unnest(string_to_array(COALESCE(diretor, ''), ',')) AS nome
		WHERE TRIM(nome) <> ''
		  AND TRIM(nome) ILIKE '%' || $1 || '%'
		ORDER BY diretor
		LIMIT 8`, termo)
	if err != nil {
		log.Printf("Erro ao buscar diretores: %v", err)
		return nil, err
	}
	defer rows.Close()

	diretores := []string{}
	for rows.Next() {
		var diretor string
		if err := rows.Scan(&diretor); err != nil {
			return nil, err
		}
		diretores = append(diretores, diretor)
	}

	return diretores, rows.Err()
}
