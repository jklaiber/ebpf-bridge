package service

import (
	"fmt"
	"net"
)

type Writer interface {
	Write(message string) error
}

type MessageWriter struct {
	conn net.Conn
}

// NewMessageWriter creates a new MessageWriter
//
// Returns:
//   - *MessageWriter: The MessageWriter instance.
//   - error: An error if the connection could not be established.
func NewMessageWriter() (*MessageWriter, error) {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to socket: %w", err)
	}
	return &MessageWriter{conn: conn}, nil
}

// Write writes a message to the socket.
//
// Parameters:
//   - w *MessageWriter: The MessageWriter instance.
//   - message string: The message to write.
//
// Returns:
//   - error: An error if the message could not be written.
func (w *MessageWriter) Write(message string) error {
	defer w.conn.Close()
	_, err := w.conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("could not write to socket: %w", err)
	}
	return nil
}
