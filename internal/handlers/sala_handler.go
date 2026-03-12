package handlers

import (
	"TinderDosFilmes/internal/models"
	"TinderDosFilmes/internal/services"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type SalaHandler struct {
	Service *services.TMDBService
}

var salas = map[string]models.Sala{}

type CriarSalaRequest struct {
	Genero    int `json:"genero"`
	Streaming int `json:"streaming"`
}

type CriarSalaResponse struct {
	SalaID string `json:"salaId"`
}

func gerarCodigoSala() string {
	const letras = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codigo := make([]byte, 6)
	for i := range codigo {
		codigo[i] = letras[rand.Intn(len(letras))]
	}
	return string(codigo)
}

func (h *SalaHandler) CriarSala(w http.ResponseWriter, r *http.Request) {
	var req CriarSalaRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	filmes, err := h.Service.BuscarFilmes(req.Genero, req.Streaming)
	if err != nil {
		http.Error(w, "Erro ao buscar filmes", http.StatusInternalServerError)
		return
	}

	codigo := gerarCodigoSala()

	sala := models.Sala{
		ID:        codigo,
		Genero:    req.Genero,
		Streaming: req.Streaming,
		Filmes:    filmes,
		CriadaEm:  time.Now(),
		ExpiraEm:  time.Now().Add(2 * time.Hour),
	}

	salas[codigo] = sala

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CriarSalaResponse{SalaID: codigo})
}

func (h *SalaHandler) Sala(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/sala/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	sala, ok := salas[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 1 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": sala.ID,
				"filtros": map[string]interface{}{
					"generos":    []int{sala.Genero},
					"streamings": []int{sala.Streaming},
				},
				"participantes": []interface{}{},
				"status":        "lobby",
			})
			return
		}
		if len(parts) == 2 && parts[1] == "filmes" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sala.Filmes)
			return
		}
	case http.MethodPost:
		if len(parts) == 2 && parts[1] == "entrar" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"userId": "1",
				"token":  "fake-token",
			})
			return
		}
	}

	http.NotFound(w, r)
}
