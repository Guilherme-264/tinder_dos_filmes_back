package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSMessage struct {
	Tipo    string          `json:"tipo"`
	Payload json.RawMessage `json:"payload"`
}

type apelidoPayload struct {
	Apelido string `json:"apelido"`
}

func limparApelido(apelido string) string {
	apelido = strings.TrimSpace(apelido)
	if apelido == "" {
		return "Jogador"
	}

	runes := []rune(apelido)
	if len(runes) > 24 {
		return string(runes[:24])
	}

	return apelido
}

func montarMensagem(tipo string, payload interface{}) []byte {
	payloadJSON, _ := json.Marshal(payload)
	msg, _ := json.Marshal(WSMessage{
		Tipo:    tipo,
		Payload: payloadJSON,
	})
	return msg
}

func (h *SalaHandler) WSHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/ws/sala/")
	salaID := strings.Split(path, "/")[0]

	salasMu.RLock()
	sala, ok := salas[salaID]
	salasMu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao fazer upgrade WebSocket: %v", err)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = gerarCodigoSala()
	}
	apelido := limparApelido(r.URL.Query().Get("apelido"))

	cliente := &Cliente{
		ID:      userID,
		Apelido: apelido,
		Conn:    conn,
		Send:    make(chan []byte, 64),
	}

	hub := getOrCreateHub(salaID)

	hub.mu.Lock()
	if clienteAntigo, existe := hub.Clientes[userID]; existe {
		close(clienteAntigo.Send)
	}
	hub.Clientes[userID] = cliente
	totalAtual := len(hub.Clientes)
	jogadores := hub.Jogadores()
	hub.mu.Unlock()

	cliente.Send <- montarMensagem("sala_atual", map[string]interface{}{
		"userId":    userID,
		"total":     totalAtual,
		"jogadores": jogadores,
	})

	hub.Broadcast(montarMensagem("jogador_entrou", map[string]string{
		"userId":  userID,
		"apelido": apelido,
		"status":  "entrou",
	}))

	go func() {
		defer conn.Close()
		for msg := range cliente.Send {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
	}()

	defer func() {
		hub.mu.Lock()
		delete(hub.Clientes, userID)
		hub.mu.Unlock()
		close(cliente.Send)
		conn.Close()

		hub.Broadcast(montarMensagem("jogador_saiu", map[string]string{
			"userId": userID,
			"status": "saiu",
		}))
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Tipo {
		case "apelido":
			var payload apelidoPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}

			apelido = limparApelido(payload.Apelido)
			hub.AtualizarApelido(userID, apelido)
			hub.Broadcast(montarMensagem("jogador_atualizado", map[string]string{
				"userId":  userID,
				"apelido": apelido,
			}))
		case "voto":
			var voto Voto
			if err := json.Unmarshal(msg.Payload, &voto); err != nil {
				continue
			}
			voto.UserID = userID

			hub.mu.RLock()
			totalJogadores := len(hub.Clientes)
			hub.mu.RUnlock()

			isMatch := hub.RegistrarVoto(voto, totalJogadores)
			hub.Broadcast(montarMensagem("voto_registrado", voto))

			if isMatch {
				var filmeMatch interface{} = nil
				salasMu.RLock()
				for _, f := range sala.Filmes {
					if f.ID == voto.FilmeID {
						filmeMatch = f
						break
					}
				}
				salasMu.RUnlock()

				hub.Broadcast(montarMensagem("match", filmeMatch))
			}
		case "iniciar_sala":
			hub.Broadcast(montarMensagem("sala_iniciada", map[string]interface{}{}))
		}
	}
}
