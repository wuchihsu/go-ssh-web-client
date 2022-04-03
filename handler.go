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
)

var terminalModes = ssh.TerminalModes{
	ssh.ECHO:          1,     // enable echoing (different from the example in docs)
	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
}

type windowSize struct {
	High  int `json:"high"`
	Width int `json:"width"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize,
	WriteBufferSize: maxMessageSize,
}

type sshClient struct {
	conn     *websocket.Conn
	addr     string
	user     string
	secret   string
	client   *ssh.Client
	sess     *ssh.Session
	sessIn   io.WriteCloser
	sessOut  io.Reader
	closeSig chan struct{}
}

func (c *sshClient) getWindowSize() (wdSize *windowSize, err error) {
	c.conn.SetReadDeadline(time.Now().Add(messageWait))
	msgType, msg, err := c.conn.ReadMessage()
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

func (c *sshClient) wsWrite() error {
	defer func() {
		c.closeSig <- struct{}{}
	}()

	data := make([]byte, maxMessageSize, maxMessageSize)

	for {
		time.Sleep(10 * time.Millisecond)
		n, readErr := c.sessOut.Read(data)
		if n > 0 {
			c.conn.SetWriteDeadline(time.Now().Add(messageWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, data[:n]); err != nil {
				return fmt.Errorf("conn.WriteMessage: %w", err)
			}
		}
		if readErr != nil {
			return fmt.Errorf("sessOut.Read: %w", readErr)
		}
	}
}

func (c *sshClient) wsRead() error {
	defer func() {
		c.closeSig <- struct{}{}
	}()

	var zeroTime time.Time
	c.conn.SetReadDeadline(zeroTime)

	for {
		msgType, connReader, err := c.conn.NextReader()
		if err != nil {
			return fmt.Errorf("conn.NextReader: %w", err)
		}
		if msgType != websocket.BinaryMessage {
			if _, err := io.Copy(c.sessIn, connReader); err != nil {
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

		if err := c.sess.WindowChange(wdSize.High, wdSize.Width); err != nil {
			return fmt.Errorf("sess.WindowChange: %w", err)
		}
	}
}

func (c *sshClient) bridgeWSAndSSH() {
	defer c.conn.Close()

	wdSize, err := c.getWindowSize()
	if err != nil {
		log.Println("bridgeWSAndSSH: getWindowSize:", err)
		return
	}

	// log.Println("wdSize:", wdSize)

	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.secret),
		},
		// InsecureIgnoreHostKey returns a function
		// that can be used for ClientConfig.HostKeyCallback
		// to accept any host key.
		// It should not be used for production code.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	c.client, err = ssh.Dial("tcp", c.addr, config)
	if err != nil {
		log.Println("bridgeWSAndSSH: ssh.Dial:", err)
		return
	}
	defer c.client.Close()

	c.sess, err = c.client.NewSession()
	if err != nil {
		log.Println("bridgeWSAndSSH: client.NewSession:", err)
		return
	}
	defer c.sess.Close()

	c.sess.Stderr = os.Stderr // TODO: check proper Stderr output
	c.sessOut, err = c.sess.StdoutPipe()
	if err != nil {
		log.Println("bridgeWSAndSSH: session.StdoutPipe:", err)
		return
	}

	c.sessIn, err = c.sess.StdinPipe()
	if err != nil {
		log.Println("bridgeWSAndSSH: session.StdinPipe:", err)
		return
	}
	defer c.sessIn.Close()

	if err := c.sess.RequestPty("xterm", wdSize.High, wdSize.Width, terminalModes); err != nil {
		log.Println("bridgeWSAndSSH: session.RequestPty:", err)
		return
	}
	if err := c.sess.Shell(); err != nil {
		log.Println("bridgeWSAndSSH: session.Shell:", err)
		return
	}

	log.Println("started a login shell on the remote host")
	defer log.Println("closed a login shell on the remote host")

	go func() {
		if err := c.wsRead(); err != nil {
			log.Println("bridgeWSAndSSH: wsRead:", err)
		}
	}()

	go func() {
		if err := c.wsWrite(); err != nil {
			log.Println("bridgeWSAndSSH: wsWrite:", err)
		}
	}()

	<-c.closeSig
}

type sshHandler struct {
	addr   string
	user   string
	secret string
}

// webSocket handles WebSocket requests for SSH from the clients.
func (h *sshHandler) webSocket(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("upgrader.Upgrade:", err)
		return
	}

	sshCli := &sshClient{
		conn:     conn,
		addr:     h.addr,
		user:     h.user,
		secret:   h.secret,
		closeSig: make(chan struct{}, 1),
	}
	go sshCli.bridgeWSAndSSH()
}
