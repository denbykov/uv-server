package main

import (
	"fmt"
	"net/http"
	"server/common/loggers"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(c *http.Request) bool { return true },
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("read error:", err)
			break
		}

		fmt.Printf("Received: %s\n", msg)

		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			fmt.Println("Write error:", err)
			break
		}
	}
}

func main() {
	loggers.Init()
	defer loggers.CloseLogFile()

	loggers.ApplicationLogger.Info("Starting...")

	// http.HandleFunc("/ws", handleConnection)

	// fmt.Println("Websocket server started on :3000")
	// err := http.ListenAndServe(":3000", nil)
	// if err != nil {
	// 	fmt.Println("ListenAndServe:", err)
	// }
}
