package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize,
	WriteBufferSize: maxMessageSize,
	CheckOrigin:     checkOrigin,
}

func wsWrite(sshCli *sshClient, conn *websocket.Conn) {
	defer func() {
		sshCli.Close()
		conn.Close()
	}()

	sshCli.SetSessionStderr(os.Stderr)
	sessOut, err := sshCli.SessionStdoutPipe()
	if err != nil {
		log.Println("wsWrite: sshCli.SessionStdoutPipe:", err)
		return
	}

	data := make([]byte, maxMessageSize, maxMessageSize)

	for {
		time.Sleep(10 * time.Millisecond)
		n, readErr := sessOut.Read(data)
		if n > 0 {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data[:n]); err != nil {
				log.Println("wsWrite: conn.WriteMessage:", err)
				return
			}
		}
		if readErr != nil {
			log.Println("wsWrite: sessOut.Read:", readErr)
			return
		}
	}
}

func wsRead(sshCli *sshClient, conn *websocket.Conn) {
	defer func() {
		sshCli.Close()
		conn.Close()
	}()

	sessIn, err := sshCli.SessionStdinPipe()
	if err != nil {
		log.Println("wsRead: sshCli.SessionStdinPipe:", err)
		return
	}

	for {
		_, connReader, err := conn.NextReader()
		if err != nil {
			log.Println("wsRead: conn.NextReader:", err)
			return
		}
		if _, err := io.Copy(sessIn, connReader); err != nil {
			log.Println("wsRead: io.Copy:", err)
			return
		}
	}
}

// wsHandler handles websocket requests from the client.
func sshHandler(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("upgrader.Upgrade:", err)
		return
	}
	sshCli, err := newSSHClient("127.0.0.1:2222", "linuxserver.io", "cch")
	if err != nil {
		log.Println("newSSHClient:", err)
		return
	}

	go wsWrite(sshCli, conn)
	go wsRead(sshCli, conn)

	if err := sshCli.PtyAndShell(40, 80); err != nil {
		log.Println("sshCli.PtyAndShell:", err)
		return
	}
	log.Println("started a login shell on the remote host")
}

func main() {
	http.HandleFunc("/ws", sshHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
