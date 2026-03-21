package services

import (
	"TinderDosFilmes/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type DiscoverResponse struct {
	Results    []models.Filme `json:"results"`
	TotalPages int            `json:"total_pages"`
}

type TMDBService struct {
	ApiKey string
}

func (s *TMDBService) BuscarFilmes(generos, streamings []int) ([]models.Filme, error) {
	log.Printf("generos: %v, streamings: %v", generos, streamings)

	if s.ApiKey == "" {
		return nil, errors.New("TMDB API key não configurada")
	}

	var todosFilmes []models.Filme
	vistos := map[int]bool{}

	for _, streaming := range streamings {
		for _, genero := range generos {
			for pagina := 1; pagina <= 5; pagina++ {
				url := "https://api.themoviedb.org/3/discover/movie?api_key=" + s.ApiKey +
					"&language=pt-BR" +
					"&with_genres=" + strconv.Itoa(genero) +
					"&with_watch_providers=" + strconv.Itoa(streaming) +
					"&watch_region=BR" +
					"&page=" + strconv.Itoa(pagina)

				resp, err := http.Get(url)
				if err != nil {
					return nil, err
				}
				log.Printf("Buscando: %s", url)

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					return nil, fmt.Errorf("TMDB erro status %d: %s", resp.StatusCode, string(body))
				}

				body, _ := io.ReadAll(resp.Body)
				log.Printf("Resposta TMDB: %s", string(body))
				resp.Body.Close()

				var discover DiscoverResponse
				if err := json.Unmarshal(body, &discover); err != nil {
					return nil, err
				}
				resp.Body.Close() // fecha aqui, não com defer

				for _, f := range discover.Results {
					if !vistos[f.ID] {
						vistos[f.ID] = true
						f.Streaming = []int{streaming}
						todosFilmes = append(todosFilmes, f)
					}
				}

				if pagina >= discover.TotalPages {
					break
				}
				log.Printf("Página %d: %d filmes encontrados", pagina, len(discover.Results))

			}
		}
	}

	if todosFilmes == nil {
		todosFilmes = []models.Filme{}
	}
	log.Printf("Total de filmes: %d", len(todosFilmes))

	return todosFilmes, nil
}
