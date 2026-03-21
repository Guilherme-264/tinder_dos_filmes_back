package handlers

import (
	"TinderDosFilmes/internal/services"
	"encoding/json"
	"net/http"
	"strconv"
)

type MovieHandler struct {
	Service *services.TMDBService
}

func (h *MovieHandler) Discover(w http.ResponseWriter, r *http.Request) {
	generoStr := r.URL.Query().Get("genero")
	streamingStr := r.URL.Query().Get("streaming")

	genero, err := strconv.Atoi(generoStr)
	if err != nil {
		http.Error(w, "genero inválido", http.StatusBadRequest)
		return
	}

	streaming, err := strconv.Atoi(streamingStr)
	if err != nil {
		http.Error(w, "streaming inválido", http.StatusBadRequest)
		return
	}

	movies, err := h.Service.BuscarFilmes([]int{genero}, []int{streaming})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}
