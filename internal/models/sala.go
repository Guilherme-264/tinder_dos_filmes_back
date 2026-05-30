package models

import "time"

type Sala struct {
	ID         string    `json:"id"`
	Generos    []int     `json:"generos"`
	Streamings []int     `json:"streamings"`
	Filmes     []Filme   `json:"filmes"`
	CriadaEm   time.Time `json:"criada_em"`
	ExpiraEm   time.Time `json:"expira_em"`
	Status     string    `json:"status"`
}
