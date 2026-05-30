package handlers

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Cliente struct {
	ID      string
	Apelido string
	Conn    *websocket.Conn
	Send    chan []byte
}

type Jogador struct {
	ID      string `json:"id"`
	Apelido string `json:"apelido"`
}

type Voto struct {
	UserID  string `json:"userId"`
	FilmeID int    `json:"filmeId"`
	Voto    string `json:"voto"`
}

type Hub struct {
	SalaID   string
	Clientes map[string]*Cliente
	Votos    map[int][]Voto
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

func getHub(salaID string) (*Hub, bool) {
	hubsMu.RLock()
	defer hubsMu.RUnlock()

	h, ok := hubs[salaID]
	return h, ok
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, cliente := range h.Clientes {
		cliente.Send <- msg
	}
}

func (h *Hub) Jogadores() []Jogador {
	jogadores := make([]Jogador, 0, len(h.Clientes))
	for _, cliente := range h.Clientes {
		jogadores = append(jogadores, Jogador{
			ID:      cliente.ID,
			Apelido: cliente.Apelido,
		})
	}
	return jogadores
}

func (h *Hub) AtualizarApelido(userID, apelido string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if cliente, ok := h.Clientes[userID]; ok {
		cliente.Apelido = apelido
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
	for _, votoRegistrado := range h.Votos[voto.FilmeID] {
		if votoRegistrado.Voto == "like" {
			likes++
		}
	}

	return likes >= totalJogadores
}
