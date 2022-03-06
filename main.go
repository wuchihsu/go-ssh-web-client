package main

import (
	"io"
	"log"
	"os"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: checkOrigin,
}

func sshDail() (*ssh.Session, error) {
	config := &ssh.ClientConfig{
		User: "linuxserver.io",
		Auth: []ssh.AuthMethod{
			ssh.Password("cch"),
		},
		// InsecureIgnoreHostKey returns a function
		// that can be used for ClientConfig.HostKeyCallback
		// to accept any host key.
		// It should not be used for production code.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "127.0.0.1:2222", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	// defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	// defer session.Close()

	return session, nil
}

func wsWrite(sessOut io.Reader, conn *websocket.Conn) {
	defer log.Println("wsWrite: conn close")
	defer conn.Close()

	data := make([]byte, 1024, 1024)

	for {
		time.Sleep(10* time.Millisecond)
		n, readErr := sessOut.Read(data)
		if n > 0 {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data[:n]); err != nil {
				log.Println("wsWrite: conn.WriteMessage: ", err)
				return
			}
		}
		if readErr != nil {
			log.Println("wsWrite: sessOut.Read: ", readErr)
			return
		}
	}
}

func wsRead(sessIn io.WriteCloser, conn *websocket.Conn) {
	defer log.Println("wsRead: conn close")
	defer conn.Close()

	for {
		_, connReader, err := conn.NextReader()
		if err != nil {
			log.Println("wsRead: conn.NextReader: ", err)
			return
		}
		if _, err := io.Copy(sessIn, connReader); err != nil {
			log.Println("wsRead: io.Copy: ", err)
			return
		}
	}
}

// wsHandler handles websocket requests from the client.
func sshHandler(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Failed to upgrades the HTTP server connection to the WebSocket protocol: ", err)
		return
	}

	session, err := sshDail()
	if err != nil {
		log.Println("Failed to SSH dail: ", err)
		return
	}

	sessOut, err := session.StdoutPipe()
	if err != nil {
		log.Println("session.StdoutPipe: ", err)
		return
	}

	sessIn, err := session.StdinPipe()
	if err != nil {
		log.Println("session.StdinPipe: ", err)
		return
	}

	session.Stderr = os.Stderr
	go wsWrite(sessOut, conn)
	go wsRead(sessIn, conn)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing (different from the example in docs)
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		log.Fatal("request for pseudo terminal failed: ", err)
	}
	if err := session.Shell(); err != nil {
		log.Fatal("failed to start shell: ", err)
	}
	log.Println("start a login shell on the remote host")
}

func main() {
	http.HandleFunc("/ws", sshHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
