package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

const (
	// Time allowed to write or read a message.
	messageWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	addr = "127.0.0.1:2222"
	user = "linuxserver.io"
	password = "cch"
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

func getWindowSize(conn *websocket.Conn) (wdSize *windowSize, err error) {
	conn.SetReadDeadline(time.Now().Add(messageWait))
	msgType, msg, err := conn.ReadMessage()
	if msgType != websocket.BinaryMessage {
		err = fmt.Errorf("conn.ReadMessage: message type is not binary")
		return
	}
	if err != nil {
		err = fmt.Errorf("conn.ReadMessage: %w", err)
		return
	}

	// log.Println("msg:", string(msg))

	wdSize = new(windowSize)
	if err = json.Unmarshal(msg, wdSize); err != nil {
		err = fmt.Errorf("json.Unmarshal: %w", err)
		return
	}
	return
}

func wsWrite(conn *websocket.Conn, sess *ssh.Session, sessOut io.Reader) error {
	data := make([]byte, maxMessageSize, maxMessageSize)

	for {
		time.Sleep(10 * time.Millisecond)
		n, readErr := sessOut.Read(data)
		if n > 0 {
			conn.SetWriteDeadline(time.Now().Add(messageWait))
			if err := conn.WriteMessage(websocket.TextMessage, data[:n]); err != nil {
				return fmt.Errorf("conn.WriteMessage: %w", err)
			}
		}
		if readErr != nil {
			return fmt.Errorf("sessOut.Read: %w", readErr)
		}
	}
}

func wsRead(conn *websocket.Conn, sess *ssh.Session, sessIn io.WriteCloser) error {
	var zeroTime time.Time
	conn.SetReadDeadline(zeroTime)

	for {
		msgType, connReader, err := conn.NextReader()
		if err != nil {
			return fmt.Errorf("conn.NextReader: %w", err)
		}
		if msgType != websocket.BinaryMessage {
			if _, err := io.Copy(sessIn, connReader); err != nil {
				return fmt.Errorf("io.Copy: %w", err)
			}
			continue
		}

		data := make([]byte, maxMessageSize, maxMessageSize)
		n, err := connReader.Read(data)
		if err != nil {
			return fmt.Errorf("connReader.Read: %w", err)
		}

		// log.Println("data:", string(data))

		var wdSize windowSize
		if err := json.Unmarshal(data[:n], &wdSize); err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}

		// log.Println("wdSize:", wdSize)

		if err := sess.WindowChange(wdSize.High, wdSize.Width); err != nil {
			return fmt.Errorf("sess.WindowChange: %w", err)
		}
	}
}

func bridgeWSAndSSH(conn *websocket.Conn) {
	defer conn.Close()

	wdSize, err := getWindowSize(conn)
	if err != nil {
		log.Println("bridgeWSAndSSH: getWindowSize:", err)
		return
	}

	// log.Println("wdSize:", wdSize)

	// TODO: get addr, user and password from args or a config file
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// InsecureIgnoreHostKey returns a function
		// that can be used for ClientConfig.HostKeyCallback
		// to accept any host key.
		// It should not be used for production code.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Println("bridgeWSAndSSH: ssh.Dial:", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Println("bridgeWSAndSSH: client.NewSession:", err)
		return
	}
	defer session.Close()

	session.Stderr = os.Stderr  // TODO: check proper Stderr output
	sessOut, err := session.StdoutPipe()
	if err != nil {
		log.Println("bridgeWSAndSSH: session.StdoutPipe:", err)
		return
	}

	sessIn, err := session.StdinPipe()
	if err != nil {
		log.Println("bridgeWSAndSSH: session.StdinPipe:", err)
		return
	}

	if err := session.RequestPty("xterm", wdSize.High, wdSize.Width, terminalModes); err != nil {
		log.Println("bridgeWSAndSSH: session.RequestPty:", err)
		return
	}
	if err := session.Shell(); err != nil {
		log.Println("bridgeWSAndSSH: session.Shell:", err)
		return
	}

	log.Println("started a login shell on the remote host")

	go func() {
		if err := wsRead(conn, session, sessIn); err != nil {
			log.Println("bridgeWSAndSSH: wsRead:", err)
		}
	}()

	if err := wsWrite(conn, session, sessOut); err != nil {
		log.Println("bridgeWSAndSSH: wsWrite:", err)
	}
}

// handleSSHWebSocket handles websocket requests for SSH from the clients.
func handleSSHWebSocket(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("upgrader.Upgrade:", err)
		return
	}

	go bridgeWSAndSSH(conn)
}
