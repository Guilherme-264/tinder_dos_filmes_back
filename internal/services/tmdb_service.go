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
	"strconv"
	"time"
)

type DiscoverResponse struct {
	Results    []models.Filme `json:"results"`
	TotalPages int            `json:"total_pages"`
}

type TMDBService struct {
	ApiKey string
}

// Busca filmes de um gênero + streaming específico
func (s *TMDBService) buscarPorGeneroStreaming(genero, streaming, pagina int) ([]models.Filme, error) {
	url := "https://api.themoviedb.org/3/discover/movie?api_key=" + s.ApiKey +
		"&language=pt-BR" +
		"&with_genres=" + strconv.Itoa(genero) +
		"&with_watch_providers=" + strconv.Itoa(streaming) +
		"&watch_region=BR" +
		"&sort_by=popularity.desc" +
		"&page=" + strconv.Itoa(pagina)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB erro status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)

	var discover DiscoverResponse
	if err := json.Unmarshal(body, &discover); err != nil {
		return nil, err
	}

	return discover.Results, nil
}

func (s *TMDBService) BuscarFilmes(generos, streamings []int) ([]models.Filme, error) {
	log.Printf("generos: %v, streamings: %v", generos, streamings)

	if s.ApiKey == "" {
		return nil, errors.New("TMDB API key não configurada")
	}

	const totalAlvo = 100
	qtdGeneros := len(generos)
	if qtdGeneros == 0 {
		return []models.Filme{}, nil
	}

	// Calcula quantos filmes por gênero proporcionalmente
	filmesPorGenero := totalAlvo / qtdGeneros
	resto := totalAlvo % qtdGeneros

	vistos := map[int]bool{}
	var resultado []models.Filme

	for i, genero := range generos {
		// Último gênero absorve o resto da divisão
		qtdAlvo := filmesPorGenero
		if i == qtdGeneros-1 {
			qtdAlvo += resto
		}

		var filmesDoGenero []models.Filme

		// Busca páginas até atingir a quantidade alvo
		for pagina := 1; pagina <= 5 && len(filmesDoGenero) < qtdAlvo; pagina++ {
			for _, streaming := range streamings {
				filmes, err := s.buscarPorGeneroStreaming(genero, streaming, pagina)
				if err != nil {
					log.Printf("Erro ao buscar genero %d streaming %d: %v", genero, streaming, err)
					continue
				}
				for _, f := range filmes {
					if !vistos[f.ID] {
						vistos[f.ID] = true
						f.Streaming = []int{streaming}
						filmesDoGenero = append(filmesDoGenero, f)
					}
				}
			}
		}

		// Limita à quantidade proporcional
		if len(filmesDoGenero) > qtdAlvo {
			filmesDoGenero = filmesDoGenero[:qtdAlvo]
		}

		log.Printf("Gênero %d: %d/%d filmes", genero, len(filmesDoGenero), qtdAlvo)
		resultado = append(resultado, filmesDoGenero...)
	}

	// Embaralha no backend antes de retornar
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(resultado), func(i, j int) {
		resultado[i], resultado[j] = resultado[j], resultado[i]
	})

	log.Printf("Total final: %d filmes embaralhados", len(resultado))
	return resultado, nil
}
