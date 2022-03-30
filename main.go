package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/web-socket/ssh", handleSSHWebSocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}