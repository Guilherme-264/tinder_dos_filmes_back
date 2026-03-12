package models

import "time"

type Sala struct {
	ID        string    `json:"id"`
	Genero    int       `json:"genero"`
	Streaming int       `json:"streaming"`
	Filmes    []Filme   `json:"filmes"`
	CriadaEm  time.Time `json:"criada_em"`
	ExpiraEm  time.Time `json:"expira_em"`
}
