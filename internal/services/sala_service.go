package services

import "TinderDosFilmes/internal/models"

type Sala struct {
	ID        string
	Genero    int
	Streaming int
	Filmes    []models.Filme
}
