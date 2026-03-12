package services

import (
	"TinderDosFilmes/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type DiscoverResponse struct {
	Results []models.Filme `json:"results"`
}

type TMDBService struct {
	ApiKey string
}

func (s *TMDBService) BuscarFilmes(genero, streaming int) ([]models.Filme, error) {
	if s.ApiKey == "" {
		return nil, errors.New("TMDB API key não configurada")
	}

	url := "https://api.themoviedb.org/3/discover/movie?api_key=" + s.ApiKey +
		"&language=pt-BR" +
		"&with_genres=" + strconv.Itoa(genero) +
		"&with_watch_providers=" + strconv.Itoa(streaming) +
		"&watch_region=BR"

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
	err = json.NewDecoder(resp.Body).Decode(&discover)
	if err != nil {
		return nil, err
	}

	if discover.Results == nil {
		discover.Results = []models.Filme{}
	}

	return discover.Results, nil
}
