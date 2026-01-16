package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mildlybrutal/websocketGo/internal/server"
)

func main() {
	go server.MainHub.Run()

	http.HandleFunc("/ws", server.HandleConnections)

	fmt.Println("Websocket server started at port 8080")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
