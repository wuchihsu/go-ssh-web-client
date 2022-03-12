package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write or read a message.
	messageWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type windowSize struct {
	High int `json:"high"`
	Width int `json:"width"`
}

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
			conn.SetWriteDeadline(time.Now().Add(messageWait))
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

	var zeroTime time.Time
	conn.SetReadDeadline(zeroTime)

	for {
		msgType, connReader, err := conn.NextReader()
		if err != nil {
			log.Println("wsRead: conn.NextReader:", err)
			return
		}
		if msgType != websocket.BinaryMessage {
			if _, err := io.Copy(sessIn, connReader); err != nil {
				log.Println("wsRead: io.Copy:", err)
				return
			}
			continue
		}
		data := make([]byte, maxMessageSize, maxMessageSize)
		n, err := connReader.Read(data)
		if err != nil {
			log.Println("wsRead: connReader.Read:", err)
			return
		}
		log.Println("data:", string(data))
		var wdSize windowSize
		if err := json.Unmarshal(data[:n], &wdSize); err != nil {
			log.Println("wsRead: json.Unmarshal:", err)
			return
		}
		if err := sshCli.PtyWindowChange(wdSize.High, wdSize.Width); err != nil {
			log.Println("wsRead: sshCli.PtyWindowChange:", err)
			return
		}
	}
}

// sshHandler handles websocket requests for SSH from the clients.
func sshHandler(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("sshHandler: upgrader.Upgrade:", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(messageWait))
	msgType, msg, err := conn.ReadMessage()
	if msgType != websocket.BinaryMessage {
		log.Println("sshHandler: conn.ReadMessage: message type is not binary")
		conn.Close()
		return
	}
	if err != nil {
		log.Println("sshHandler: conn.ReadMessage:", err)
		conn.Close()
		return
	}

	log.Println("msg:", string(msg))

	var wdSize windowSize
	if err := json.Unmarshal(msg, &wdSize); err != nil {
		log.Println("sshHandler: json.Unmarshal:", err)
		conn.Close()
		return
	}

	sshCli, err := newSSHClient("127.0.0.1:2222", "linuxserver.io", "cch")
	if err != nil {
		log.Println("sshHandler: newSSHClient:", err)
		conn.Close()
		return
	}

	go wsWrite(sshCli, conn)
	go wsRead(sshCli, conn)

	log.Println("wdSize:", wdSize)

	if err := sshCli.PtyAndShell(wdSize.High, wdSize.Width); err != nil {
		log.Println("sshHandler: sshCli.PtyAndShell:", err)
		sshCli.Close()
		conn.Close()
		return
	}
	log.Println("started a login shell on the remote host")
}
