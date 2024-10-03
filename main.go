package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"xhchat.tsantana.dev/views"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

var availableNames = []string{
	"Tamanduá Ágil",
	"Arara Brilhante",
	"Onça Corajosa",
	"Jaguar Feroz",
	"Sapo Gentil",
	"Tatu Ligeiro",
	"Bicho-Preguiça Sábio",
	"Capivara Valente",
	"Gato-do-mato Veloz",
	"Pica-pau Amável",
	"Peixe-boi Sereno",
	"Quati Estiloso",
	"Mico-leão-dourado Bravo",
	"Tamanduá-bandeira Destemido",
	"Guará Inteligente",
}
var mu sync.Mutex

type Message struct {
	Content string `json:"content"`
	User    string `json:"user"`
}

var messages []Message

func main() {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("views/css"))
	mux.Handle("/views/css/", http.StripPrefix("/views/css/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username := assignUsername()
		component := views.Index(username)
		component.Render(r.Context(), w)
	})

	// Rota para WebSocket
	mux.HandleFunc("/ws", handleConnections)

	// Inicia o servidor e goroutine para lidar com as mensagens
	go handleMessages()

	log.Println("Servidor iniciado na porta :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func assignUsername() string {
	mu.Lock()
	defer mu.Unlock()

	if len(availableNames) == 0 {
		return "Guest" // Nome padrão caso acabem os nomes
	}

	name := availableNames[0]
	availableNames = availableNames[1:] // Remove o nome atribuído da lista
	return name
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Faz upgrade da conexão HTTP para WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Adiciona o cliente ao map de clientes conectados
	clients[ws] = true

	// Escuta mensagens do cliente
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Envia a mensagem para o canal broadcast
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Recebe mensagem do canal broadcast
		msg := <-broadcast
		// Envia a mensagem para todos os clientes conectados
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
