package game

import (
	"fmt"
	"log"
	"strings"
)

func handleWhisper(g *Game, session *Session, params string, help bool) []OutputEvent {
	parts := strings.Split(params, " ")
	if help || len(parts) < 2 {
		return []OutputEvent{{
			SessionID: session.ID,
			Message:   "Say something privately\nUsage: /whisper <username> <message>",
		}}
	}
	targetUsername := parts[0]
	message := strings.Join(parts[1:], " ")
	
	log.Printf("User %s issued say command %+v", session.Name, parts)
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