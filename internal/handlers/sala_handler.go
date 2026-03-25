package handlers

import (
	"TinderDosFilmes/internal/models"
	"TinderDosFilmes/internal/services"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
)

type SalaHandler struct {
	Service *services.TMDBService
	DB      *sql.DB
}

var (
	salas   = map[string]models.Sala{}
	salasMu sync.RWMutex
)

type CriarSalaRequest struct {
	Generos    []int `json:"generos"`
	Streamings []int `json:"streamings"`
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	filmes, err := h.Service.BuscarFilmes(req.Generos, req.Streamings)
	if err != nil {
		http.Error(w, "Erro ao buscar filmes", http.StatusInternalServerError)
		return
	}

	codigo := gerarCodigoSala()
	sala := models.Sala{
		ID:         codigo,
		Generos:    req.Generos,
		Streamings: req.Streamings,
		Filmes:     filmes,
		CriadaEm:   time.Now(),
		ExpiraEm:   time.Now().Add(2 * time.Hour),
		Status:     "lobby",
	}

	_, err = h.DB.Exec(
		`INSERT INTO salas (id, generos, streamings, criado_em, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		sala.ID,
		pq.Array(sala.Generos),
		pq.Array(sala.Streamings),
		sala.CriadaEm,
		sala.Status,
	)
	if err != nil {
		log.Printf("Erro ao salvar sala: %v", err)
		http.Error(w, "Erro ao salvar sala", http.StatusInternalServerError)
		return
	}

	salasMu.Lock()
	salas[codigo] = sala
	salasMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CriarSalaResponse{SalaID: codigo})
}

func (h *SalaHandler) ApagarSalas() {
	agora := time.Now()
	salasMu.Lock()
	defer salasMu.Unlock()

	for id, sala := range salas {
		if agora.After(sala.ExpiraEm) {
			delete(salas, id)

			_, err := h.DB.Exec(`DELETE FROM salas WHERE id = $1`, id)
			if err != nil {
				log.Printf("Erro ao apagar sala %s do banco: %v", id, err)
			} else {
				log.Printf("Sala %s expirada e removida", id)
			}

		}

	}

}

func (h *SalaHandler) Sala(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/sala/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	salasMu.RLock()
	sala, ok := salas[id]
	salasMu.RUnlock()

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
					"generos":    sala.Generos,
					"streamings": sala.Streamings,
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
