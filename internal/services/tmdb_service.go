package services

import (
	"TinderDosFilmes/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
)

type DiscoverResponse struct {
	Results    []models.Filme `json:"results"`
	TotalPages int            `json:"total_pages"`
}

type TMDBService struct {
	ApiKey string
}

func (s *TMDBService) BuscarFilmes(generos []int, streamings []int) ([]models.Filme, error) {
	const LIMITE_TOTAL = 100

	type combinacao struct {
		genero    int
		streaming int
	}

	combinacoes := []combinacao{}
	for _, g := range generos {
		for _, st := range streamings {
			combinacoes = append(combinacoes, combinacao{g, st})
		}
	}

	if len(combinacoes) == 0 {
		return []models.Filme{}, nil
	}

	porCombinacao := LIMITE_TOTAL / len(combinacoes)
	if porCombinacao < 1 {
		porCombinacao = 1
	}

	type resultado struct {
		filmes []models.Filme
		err    error
	}

	// dispara todas as requests em paralelo
	ch := make(chan resultado, len(combinacoes))
	for _, c := range combinacoes {
		go func(g, st int) {
			filmes, err := s.buscarComLimite(g, st, porCombinacao)
			ch <- resultado{filmes, err}
		}(c.genero, c.streaming)
	}

	// coleta resultados
	vistos := map[int]bool{}
	todos := []models.Filme{}

	for range combinacoes {
		res := <-ch
		if res.err != nil {
			log.Printf("Erro ao buscar filmes: %v", res.err)
			continue
		}
		for _, f := range res.filmes {
			if !vistos[f.ID] {
				vistos[f.ID] = true
				todos = append(todos, f)
			}
		}
	}

	rand.Shuffle(len(todos), func(i, j int) {
		todos[i], todos[j] = todos[j], todos[i]
	})

	if len(todos) > LIMITE_TOTAL {
		todos = todos[:LIMITE_TOTAL]
	}

	return todos, nil
}

func (s *TMDBService) buscarComLimite(genero, streaming, limite int) ([]models.Filme, error) {
	if s.ApiKey == "" {
		return nil, errors.New("TMDB API key não configurada")
	}

	filmes := []models.Filme{}
	pagina := 1

	for len(filmes) < limite {
		url := fmt.Sprintf(
			"https://api.themoviedb.org/3/discover/movie?api_key=%s&language=pt-BR&with_genres=%d&with_watch_providers=%d&watch_region=BR&page=%d",
			s.ApiKey, genero, streaming, pagina,
		)

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("TMDB erro status %d: %s", resp.StatusCode, string(body))
		}

		var discover DiscoverResponse
		if err := json.NewDecoder(resp.Body).Decode(&discover); err != nil {
			return nil, err
		}

		if len(discover.Results) == 0 {
			break
		}

		filmes = append(filmes, discover.Results...)
		pagina++

		if pagina > discover.TotalPages {
			break
		}
	}

	if len(filmes) > limite {
		filmes = filmes[:limite]
	}

	return filmes, nil
}
