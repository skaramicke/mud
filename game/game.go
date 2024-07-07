package game

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type Game struct {
	sessions      map[string]*Session
	lobby         *Room
	mu            sync.Mutex
	inputChannel  chan InputEvent
	outputChannels map[string]chan OutputEvent
}

type Session struct {
	ID   string
	Name string
	Room *Room
}

type Room struct {
	Name     string
	Sessions map[string]*Session
}

type InputEvent struct {
	SessionID string
	Input     string
}

type OutputEvent struct {
	SessionID string
	Message   string
	Quit      bool
}

func NewGame() *Game {
	g := &Game{
			sessions:       make(map[string]*Session),
			lobby:          &Room{Name: "Lobby", Sessions: make(map[string]*Session)},
			inputChannel:   make(chan InputEvent, 100),
			outputChannels: make(map[string]chan OutputEvent),
	}
	go g.processEvents()
	return g
}

func (g *Game) processEvents() {
	for input := range g.inputChannel {
		g.handleInput(input)
	}
}

func (g *Game) handleInput(event InputEvent) {
	var messagesToSend []OutputEvent

	g.mu.Lock()
	session, exists := g.sessions[event.SessionID]
	if !exists {
		// New session
		session = &Session{
			ID:   event.SessionID,
			Name: event.Input, // Use the first input as the name
			Room: g.lobby,
		}
		g.sessions[event.SessionID] = session
		g.lobby.Sessions[event.SessionID] = session

		messagesToSend = append(messagesToSend, g.collectBroadcastMessages(session.Room, fmt.Sprintf("%s has entered the lobby.", session.Name), session.ID)...)
		messagesToSend = append(messagesToSend, OutputEvent{
			SessionID: event.SessionID,
			Message:   "Welcome to the lobby!",
		})
		// Inform the new user about others in the lobby
		for _, s := range g.lobby.Sessions {
			if s.ID != event.SessionID {
				messagesToSend = append(messagesToSend, OutputEvent{
					SessionID: event.SessionID,
					Message:   fmt.Sprintf("%s is already in the lobby.", s.Name),
				})
			}
		}
	} else {
		// Check if the input is a command or chat
		if strings.HasPrefix(event.Input, "/") {
			output, quit := g.handleCommand(session, event.Input[1:])
			messagesToSend = append(messagesToSend, output)
			if quit {
				delete(g.sessions, session.ID)
				delete(session.Room.Sessions, session.ID)
				messagesToSend = append(messagesToSend, g.collectBroadcastMessages(session.Room, fmt.Sprintf("%s has left the room.", session.Name), "")...)
			}
		} else {
			// Treat as chat and broadcast to the room
			messagesToSend = append(messagesToSend, g.collectBroadcastMessages(session.Room, fmt.Sprintf("%s says: %s", session.Name, event.Input), "")...)
		}
	}
	g.mu.Unlock()

	// Send collected messages after releasing the lock
	for _, msg := range messagesToSend {
		g.sendOutput(msg)
	}
}

func (g *Game) handleCommand(session *Session, command string) (OutputEvent, bool) {
	parts := strings.SplitN(command, " ", 2)
	cmd := parts[0]
	switch cmd {
	case "who":
		log.Printf("User %s requested list of users in room %s", session.Name, session.Room.Name)
		return g.listUsersInRoom(session), false
	case "quit":
		return OutputEvent{
			SessionID: session.ID,
			Message:   "Goodbye!",
			Quit:      true,
		}, true
	default:
		return OutputEvent{
			SessionID: session.ID,
			Message:   fmt.Sprintf("Unknown command: %s", cmd),
		}, false
	}
}

func (g *Game) listUsersInRoom(session *Session) OutputEvent {
	userList := []string{}
	for _, s := range session.Room.Sessions {
		userList = append(userList, s.Name)
	}
	return OutputEvent{
		SessionID: session.ID,
		Message:   fmt.Sprintf("Users in this room: %s", strings.Join(userList, ", ")),
	}
}

func (g *Game) collectBroadcastMessages(room *Room, message string, excludeID string) []OutputEvent {
	var messages []OutputEvent
	for sessionID := range room.Sessions {
		if sessionID != excludeID {
			messages = append(messages, OutputEvent{SessionID: sessionID, Message: message})
		}
	}
	return messages
}

func (g *Game) sendOutput(event OutputEvent) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if event.SessionID == "" {
			// Broadcast to all sessions
			for _, ch := range g.outputChannels {
					select {
					case ch <- event:
					default:
							log.Printf("Output channel full, discarding message: %s", event.Message)
					}
			}
	} else {
			// Send to specific session
			if ch, exists := g.outputChannels[event.SessionID]; exists {
					select {
					case ch <- event:
					default:
							log.Printf("Output channel full, discarding message: %s", event.Message)
					}
			}
	}
}

func (g *Game) GetInputChannel() chan<- InputEvent {
	return g.inputChannel
}

func (g *Game) GetOutputChannel(sessionID string) <-chan OutputEvent {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.outputChannels[sessionID]; !exists {
			g.outputChannels[sessionID] = make(chan OutputEvent, 100)
	}
	return g.outputChannels[sessionID]
}
