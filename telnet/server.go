package telnet

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"mud/game"
)

const (
	HOST = "localhost"
	PORT = "2323"
)

type Server struct {
	game *game.Game
}

func NewServer(game *game.Game) *Server {
	return &Server{game: game}
}
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	sessionID := conn.RemoteAddr().String()
	fmt.Fprintf(conn, "Welcome to the MUD server!\nPlease enter your name: ")

	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading name: %v", err)
		return
	}
	name = strings.TrimSpace(name)

	s.game.GetInputChannel() <- game.InputEvent{SessionID: sessionID, Input: name}

	// Get a dedicated output channel for this session
	outputChan := s.game.GetOutputChannel(sessionID)

	// Create a channel to signal when to quit
	quitChan := make(chan bool)

	// Start a goroutine to handle outgoing messages
	go s.handleOutgoing(conn, outputChan, quitChan)

	// Main input loop
	for {
		// Remove the deadline for blocking reads
		conn.SetReadDeadline(time.Time{})

		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF || isClosedConnError(err) {
				// Connection was closed
				break
			} else {
				// This is a real error
				log.Printf("Error reading from connection: %v", err)
				break
			}
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		log.Printf("Received from %s: %s", name, input)
		s.game.GetInputChannel() <- game.InputEvent{SessionID: sessionID, Input: input}
	}

	// Signal handleOutgoing to stop
	close(quitChan)

	log.Printf("Connection closed for %s", name)
}

// isClosedConnError checks if the error is due to a closed connection
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}

func (s *Server) handleOutgoing(conn net.Conn, outputChan <-chan game.OutputEvent, quitChan <-chan bool) {
	for {
			select {
			case <-quitChan:
					return
			case output, ok := <-outputChan:
					if !ok {
							return
					}
					_, err := fmt.Fprintf(conn, "%s\n", output.Message)
					if err != nil {
							log.Printf("Error writing to connection: %v", err)
							return
					}
					if output.Quit {
							// Send Telnet End of Session command
							_, err := conn.Write([]byte{255, 244, 255, 253, 6})
							if err != nil {
									log.Printf("Error sending End of Session command: %v", err)
							}
							// Flush the connection
							if flusher, ok := conn.(interface{ Flush() error }); ok {
									flusher.Flush()
							}
							// Close the connection
							conn.Close()
							return
					}
			}
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", HOST+":"+PORT)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server listening on %s:%s", HOST, PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		log.Printf("New connection from %s", conn.RemoteAddr().String())

		go s.handleConnection(conn)
	}
}
