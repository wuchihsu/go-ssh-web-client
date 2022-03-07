package main

import (
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

var terminalModes = ssh.TerminalModes{
	ssh.ECHO:          1,     // enable echoing (different from the example in docs)
	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
}

type sshClient struct {
	config  *ssh.ClientConfig
	client  *ssh.Client
	session *ssh.Session
}

func newSSHClient(addr, user, password string) (*sshClient, error) {
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
		return nil, fmt.Errorf("failed to SSH dial: %w", err)
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	return &sshClient{config: config, client: client, session: session}, nil
}

func (c *sshClient) PtyAndShell(ptyH, ptyW int) error {
	if err := c.session.RequestPty("xterm", ptyH, ptyW, terminalModes); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}
	if err := c.session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}
	return nil
}

func (c *sshClient) SessionStderrPipe() (io.Reader, error) {
	return c.session.StderrPipe()
}

func (c *sshClient) SessionStdinPipe() (io.WriteCloser, error) {
	return c.session.StdinPipe()
}

func (c *sshClient) SessionStdoutPipe() (io.Reader, error) {
	return c.session.StdoutPipe()
}

func (c *sshClient) SetSessionStderr(writer io.Writer) {
	c.session.Stderr = writer
}

func (c *sshClient) SetSessionStdin(reader io.Reader) {
	c.session.Stdin = reader
}

func (c *sshClient) SetSessionStdout(writer io.Writer) {
	c.session.Stdout = writer
}

func (c *sshClient) Close() error {
	sessionErr := c.session.Close()
	clientErr := c.client.Close()
	if sessionErr != nil && clientErr != nil {
		return fmt.Errorf("failed to close session: %s & failed to close client: %s", sessionErr, clientErr)
	}
	if sessionErr != nil {
		return fmt.Errorf("failed to close session: %w", sessionErr)
	}
	if clientErr != nil {
		return fmt.Errorf("failed to close client: %w", clientErr)
	}
	return nil
}
