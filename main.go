package main

import (
	"context"
	"fmt"
	"net/http"
)

var port = "8080"

func main() {
	serverHandler()

	fmt.Printf("Https server started on port %v\n", port)
	fmt.Println(http.ListenAndServeTLS(fmt.Sprintf("localhost:%s", port), "./server.crt", "./server.key", nil))
}

func serverHandler() {
	ctx := context.Background()

	manager := NewManager(ctx)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serveWS)
	http.HandleFunc("/login", manager.loginHandler)
}