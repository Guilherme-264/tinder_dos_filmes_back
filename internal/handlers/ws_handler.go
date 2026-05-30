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

	cliente := &Cliente{
		ID:   userID,
		Conn: conn,
		Send: make(chan []byte, 64),
	}

	hub := getOrCreateHub(salaID)

	hub.mu.Lock()
	if clienteAntigo, existe := hub.Clientes[userID]; existe {
		close(clienteAntigo.Send)
	}
	hub.Clientes[userID] = cliente

	// monta lista de IDs
	ids := make([]string, 0, len(hub.Clientes))
	for id := range hub.Clientes {
		ids = append(ids, id)
	}
	totalAtual := len(hub.Clientes)
	hub.mu.Unlock()

	listaPayload, _ := json.Marshal(map[string]interface{}{
		"userId":    userID,
		"total":     totalAtual,
		"jogadores": ids,
	})
	listaMsg, _ := json.Marshal(WSMessage{
		Tipo:    "sala_atual",
		Payload: listaPayload,
	})
	cliente.Send <- listaMsg

	// broadcast para todos que alguém entrou
	entrouPayload, _ := json.Marshal(map[string]string{"userId": userID})
	entrouMsg, _ := json.Marshal(WSMessage{
		Tipo:    "jogador_entrou",
		Payload: entrouPayload,
	})
	hub.Broadcast(entrouMsg)

	// goroutine para escrever mensagens
	go func() {
		defer conn.Close()
		for msg := range cliente.Send {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
	}()

	// cleanup ao desconectar
	defer func() {
		hub.mu.Lock()
		delete(hub.Clientes, userID)
		hub.mu.Unlock()
		close(cliente.Send)
		conn.Close()

		saiuPayload, _ := json.Marshal(map[string]string{"userId": userID})
		saiuMsg, _ := json.Marshal(WSMessage{
			Tipo:    "jogador_saiu",
			Payload: saiuPayload,
		})
		hub.Broadcast(saiuMsg)
	}()

	// loop de leitura
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

			votoPayload, _ := json.Marshal(voto)
			votoMsg, _ := json.Marshal(WSMessage{
				Tipo:    "voto_registrado",
				Payload: votoPayload,
			})
			hub.Broadcast(votoMsg)

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

				matchPayload, _ := json.Marshal(filmeMatch)
				matchMsg, _ := json.Marshal(WSMessage{
					Tipo:    "match",
					Payload: matchPayload,
				})
				hub.Broadcast(matchMsg)
			}
		case "iniciar_sala":
			iniciarMsg, _ := json.Marshal(WSMessage{
				Tipo:    "sala_iniciada",
				Payload: json.RawMessage(`{}`),
			})
			hub.Broadcast(iniciarMsg)
		}
	}
}
