package service

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Listener interface {
	Listen() error
}

type MessageListener struct {
	messages  chan<- string
	responses <-chan string
}

// NewListener creates a new Listener
//
// Parameters:
//   - messages chan<- string: The channel to send messages to.
//   - responses <-chan string: The channel to receive responses from.
//
// Returns:
//   - *Listener: The Listener instance.
func NewListener(messages chan<- string, responses <-chan string) *MessageListener {
	return &MessageListener{messages: messages, responses: responses}
}

// Listen listens on the socket and handles incoming connections.
//
// Parameters:
//   - l *Listener: The Listener instance.
//
// Returns:
//   - error: An error if the socket could not be opened or a connection could not be accepted.
func (l *MessageListener) Listen() error {
	os.Remove(SocketPath)
	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return fmt.Errorf("could not listen on socket: %w", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("could not accept connection: %w", err)
		}
		go l.handleConnection(conn)
	}
}

// handleConnection handles a connection.
//
// Parameters:
//   - l *Listener: The Listener instance.
//   - conn net.Conn: The connection.
func (l *MessageListener) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		l.messages <- line
		resp := <-l.responses
		conn.Write([]byte(resp))
	}
}
