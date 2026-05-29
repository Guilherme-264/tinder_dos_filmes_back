package handlers

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Cliente struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

type Voto struct {
	UserID  string `json:"userId"`
	FilmeID int    `json:"filmeId"`
	Voto    string `json:"voto"` // "like" ou "dislike"
}

type Hub struct {
	SalaID   string
	Clientes map[string]*Cliente
	Votos    map[int][]Voto // filmeId → votos
	mu       sync.RWMutex
}

var (
	hubs   = map[string]*Hub{}
	hubsMu sync.RWMutex
)

func getOrCreateHub(salaID string) *Hub {
	hubsMu.Lock()
	defer hubsMu.Unlock()

	if h, ok := hubs[salaID]; ok {
		return h
	}

	h := &Hub{
		SalaID:   salaID,
		Clientes: map[string]*Cliente{},
		Votos:    map[int][]Voto{},
	}
	hubs[salaID] = h
	return h
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.Clientes {
		c.Send <- msg
	}
}

func (h *Hub) RegistrarVoto(voto Voto, totalJogadores int) (match bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Votos[voto.FilmeID] = append(h.Votos[voto.FilmeID], voto)

	if voto.Voto != "like" {
		return false
	}

	likes := 0
	for _, v := range h.Votos[voto.FilmeID] {
		if v.Voto == "like" {
			likes++
		}
	}

	return likes >= totalJogadores
}
