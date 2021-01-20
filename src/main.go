package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket" //import websocket
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

//Message contains message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	//upgrade websocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	//close connection
	defer ws.Close()

	//register new client
	clients[ws] = true

	//continuously listen to new messages and post to board
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}

}

func handleMessages() {
	for {
		//grab message
		msg := <-broadcast
		//broadcast to all clients
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

func main() {
	//serve public assets
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	//websocket route
	http.HandleFunc("/ws", handleConnections)

	//listen to incoming messages concurrently
	go handleMessages()

	//start web server on port 8080
	log.Println("Listening on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
