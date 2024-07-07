package game

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type Game struct {
	sessions     map[string]*Session
	usernames    map[string]*Session // maps username to Session
	lobby        *Room
	mu           sync.Mutex
	inputChannel chan InputEvent
}

type Session struct {
	ID            string
	Name          string
	Room          *Room
	OutputChannel chan OutputEvent
}

type Room struct {
	Name     string
	Sessions map[string]*Session
}

type InputEvent struct {
	SessionID    string
	Input        string
	ResponseChan chan bool
}

type OutputEvent struct {
	SessionID string
	Message   string
	Quit      bool
}

func NewGame() *Game {
	g := &Game{
		sessions:     make(map[string]*Session),
		usernames:    make(map[string]*Session),
		lobby:        &Room{Name: "Lobby", Sessions: make(map[string]*Session)},
		inputChannel: make(chan InputEvent, 100),
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
			ID:            event.SessionID,
			Name:          "",
			Room:          g.lobby,
			OutputChannel: make(chan OutputEvent, 100),
		}
		g.sessions[event.SessionID] = session
		g.lobby.Sessions[event.SessionID] = session

		messagesToSend = append(messagesToSend, OutputEvent{
			SessionID: event.SessionID,
			Message:   "Who are you?",
		})

		// Send confirmation that the session was created
		if event.ResponseChan != nil {
			event.ResponseChan <- true
		}
	} else {

		if session.Name == "" {
			// Check if the input is a valid username
			if _, exists := g.usernames[event.Input]; exists {
				messagesToSend = append(messagesToSend, OutputEvent{
					SessionID: event.SessionID,
					Message:   fmt.Sprintf("Username '%s' is already taken. Please enter a different username.", event.Input),
				})
			} else {
				// Set the username
				session.Name = event.Input
				g.usernames[event.Input] = session
				messagesToSend = append(messagesToSend, OutputEvent{
					SessionID: event.SessionID,
					Message:   fmt.Sprintf("Welcome, %s!", session.Name),
				})
				messagesToSend = append(messagesToSend, g.collectBroadcastMessages(session.Room, fmt.Sprintf("%s has joined the room.", session.Name), "")...)
			}
		} else if strings.HasPrefix(event.Input, "/") { // Check if the input is a command or chat
			output, quit := g.handleCommand(session, event.Input[1:])
			messagesToSend = append(messagesToSend, output...)
			if quit {
				go func() {
					delete(g.sessions, session.ID)
					delete(session.Room.Sessions, session.ID)
					close(session.OutputChannel)
				}()
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

func (g *Game) handleCommand(session *Session, command string) ([]OutputEvent, bool) {
	parts := strings.Split(command, " ")
	cmd := parts[0]
	switch cmd {
	case "whisper":
		log.Printf("User %s issued say command %+v", session.Name, parts)
		if len(parts) < 2 {
			return []OutputEvent{{
				SessionID: session.ID,
				Message:   "Usage: /say <username> <message>",
			}}, false
		}
		targetUsername := parts[1]
		message := strings.Join(parts[2:], " ")
		return g.handleWhisper(session, targetUsername, message), false
	case "who":
		log.Printf("User %s requested list of users in room %s", session.Name, session.Room.Name)
		return []OutputEvent{g.handleListUsersInRoom(session)}, false
	case "help":
		if len(parts) > 1 {
			switch parts[1] {
			case "whisper":
				return []OutputEvent{{
					SessionID: session.ID,
					Message:   "Usage: /whisper <username> <message>",
				}}, false
			case "who":
				return []OutputEvent{{
					SessionID: session.ID,
					Message:   "Usage: /who",
				}}, false
			case "quit":
				return []OutputEvent{{
					SessionID: session.ID,
					Message:   "Usage: /quit",
				}}, false
			default:
				return []OutputEvent{{
					SessionID: session.ID,
					Message:   fmt.Sprintf("Unknown command: %s", parts[1]),
				}}, false
			}
		} else {
			return []OutputEvent{{
				SessionID: session.ID,
				Message:   "Available commands: /whisper <username> <message>, /who, /quit",
			}}, false
		}
	case "quit":
		return []OutputEvent{{
			SessionID: session.ID,
			Message:   "Goodbye!",
			Quit:      true,
		}}, true
	default:
		return []OutputEvent{{
			SessionID: session.ID,
			Message:   fmt.Sprintf("Unknown command: %s", cmd),
		}}, false
	}
}

func (g *Game) handleListUsersInRoom(session *Session) OutputEvent {
	userList := []string{}
	for _, s := range session.Room.Sessions {
		userList = append(userList, s.Name)
	}
	return OutputEvent{
		SessionID: session.ID,
		Message:   fmt.Sprintf("Users in this room: %s", strings.Join(userList, ", ")),
	}
}

func (g *Game) handleWhisper(session *Session, targetUsername, message string) []OutputEvent {
	log.Print("handleSayCommand")

	targetSession, exists := g.usernames[targetUsername]
	log.Printf("User %s wants to send a private message to %s: %s", session.Name, targetUsername, message)
	if !exists {
		return []OutputEvent{{
			SessionID: session.ID,
			Message:   fmt.Sprintf("User '%s' not found.", targetUsername),
		}}
	}

	// Send message to target user
	return []OutputEvent{
		{
			SessionID: targetSession.ID,
			Message:   fmt.Sprintf("%s whispers: %s", session.Name, message),
		},
		{
			SessionID: session.ID,
			Message:   fmt.Sprintf("You whispered to %s: %s", targetUsername, message),
		},
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
		for _, session := range g.sessions {
			select {
			case session.OutputChannel <- event:
			default:
				log.Printf("Output channel full for user %s, discarding message: %s", session.Name, event.Message)
			}
		}
	} else {
		// Send to specific session
		if session, exists := g.sessions[event.SessionID]; exists {
			select {
			case session.OutputChannel <- event:
			default:
				log.Printf("Output channel full for user %s, discarding message: %s", session.Name, event.Message)
			}
		}
	}
}

func (g *Game) GetInputChannel() chan<- InputEvent {
	return g.inputChannel
}

func (g *Game) GetOutputChannel(sessionID string) (<-chan OutputEvent, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	session, exists := g.sessions[sessionID]
	if !exists {
		return nil, false
	}

	return session.OutputChannel, true
}
