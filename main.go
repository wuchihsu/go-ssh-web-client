package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/web-socket/ssh", sshHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// https://i.imgur.com/EOHSWDc.jpg
// https://i.imgur.com/FqU0AU8.jpg