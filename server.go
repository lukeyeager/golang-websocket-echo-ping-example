package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 1 * time.Second
	pingPeriod = 1 * time.Second
	pongWait   = 3 * time.Second
)

var upgrader = websocket.Upgrader{}

func serveWS(w http.ResponseWriter, r *http.Request) {
	client, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer client.Close()
	client.SetReadDeadline(time.Now().Add(pongWait))
	client.SetPongHandler(func(string) error {
		log.Println("Received pong.")
		client.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	done := make(chan bool)
	defer close(done)
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Sending ping.")
				if err := client.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
					log.Println("Ping error:", err)
				}
			case <-done:
				log.Println("Stopping ping routine.")
				return
			}
		}
	}()

	for {
		messageType, payload, err := client.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received message type=%d, payload=\"%s\"\n", messageType, payload)
		if err := client.WriteMessage(messageType, payload); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func main() {
	addr := flag.String("addr", ":8080", "address")
	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.ServeFile(response, request, "client.html")
	})
	http.HandleFunc("/ws", serveWS)
	log.Println("Listening at", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
